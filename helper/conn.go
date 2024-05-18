package helper

import (
	"NetDisk/conf"
	"errors"
	"fmt"
	"sync"
	"time"
)

var (
	ErrInvaildFactory = errors.New("must specify a factory")

	ErrConnClosed           = errors.New("connection closed")
	ErrMaxActiveConnReached = errors.New("max active conn reached")

	ErrNilConn = errors.New("operating a nil conn")
	ErrTimeout = errors.New("get conn timeout")
)

// Pool 基本方法
type Pool interface {
	// 获取资源
	Get() (interface{}, error)
	// 资源放回去
	Put(interface{}) error
	// 关闭资源
	Close(interface{}) error
	// 释放所有资源
	Release()
	// 返回当前池子内有效连接数量
	Len() int
}

// ConnectionFactory 连接工厂
type ConnectionFactory interface {
	//生成连接的方法
	Factory() (interface{}, error)
	//关闭连接的方法
	Close(interface{}) error
	//检查连接是否有效的方法
	Ping(interface{}) error
}

// 实现 Pool 接口
type connectionPool struct {
	mu                       sync.RWMutex
	conns                    chan *idleConn    // 存储最大空闲连接
	factory                  ConnectionFactory // 工厂
	idleTimeout, waitTimeOut time.Duration     // 连接空闲超时和等待超时
	maxActive                int               // 最大连接数
	openingConns             int               // 活跃的连接数
	connReqs                 []chan connReq    // 缓冲区, 存储因为无法获取连接而阻塞的请求
}

type idleConn struct {
	conn      interface{}
	startTime time.Time // 开始空闲的时间
}

type connReq struct {
	idleConn *idleConn
}

type PoolConfig struct {
	//连接池中拥有的最小连接数
	InitialCap int
	//最大并发存活连接数
	MaxCap int
	//最大空闲连接
	MaxIdle int
	// 工厂
	Factory ConnectionFactory
	//连接最大空闲时间，超过该事件则将失效
	IdleTimeout time.Duration
}

func NewConnectionPool(cfg PoolConfig) (Pool, error) {
	// 参数校验
	if cfg.InitialCap <= 0 {
		cfg.InitialCap = 1
	}
	if cfg.Factory == nil {
		return nil, ErrInvaildFactory
	}
	if cfg.MaxIdle <= 0 {
		cfg.MaxIdle = conf.MaxIdleConn
	}
	if cfg.MaxCap <= 0 {
		cfg.MaxCap = conf.MaxConn
	}
	if cfg.IdleTimeout <= 0 {
		cfg.IdleTimeout = conf.MaxIdleTime
	}

	// 初始化
	cp := &connectionPool{
		conns:        make(chan *idleConn, cfg.MaxIdle), // 最大空闲连接数
		factory:      cfg.Factory,
		idleTimeout:  cfg.IdleTimeout,
		maxActive:    cfg.MaxCap,
		openingConns: cfg.InitialCap,
		waitTimeOut:  conf.MaxIdleTime,
	}
	// 创建初始连接
	for i := 0; i < cfg.InitialCap; i++ {
		conn, err := cp.factory.Factory()
		if err != nil {
			return nil, fmt.Errorf("error occurs during creating initial conns: %+v", err)
		}
		cp.conns <- &idleConn{
			conn:      conn,
			startTime: time.Now(),
		}
	}
	return cp, nil
}

func (cp *connectionPool) Get() (interface{}, error) {
	conns := cp.conns
	if conns == nil {
		return nil, ErrConnClosed
	}
	timeoutTicker := time.After(cp.waitTimeOut)
	for {
		// 优先从空闲队列取连接
		select {
		case conn := <-conns:
			if conn == nil {
				return nil, ErrConnClosed
			}
			//判断是否超时，超时则丢弃
			if timeout := cp.idleTimeout; timeout > 0 {
				if conn.startTime.Add(timeout).Before(time.Now()) {
					//丢弃并关闭该连接
					_ = cp.Close(conn.conn)
					continue
				}
			}
			//判断是否失效，失效则丢弃，如果用户没有设定 ping 方法，就不检查
			if err := cp.factory.Ping(conn.conn); err != nil {
				_ = cp.Close(conn.conn)
				continue
			}
			return conn.conn, nil
		// 超时控制
		case <-timeoutTicker:
			return nil, ErrTimeout
		// 空闲队列没有时，挂载到阻塞队列上
		default:
			cp.mu.Lock()
			// 判断是否到达最大链接数量
			if cp.openingConns >= cp.maxActive {
				// 封装当前请求为 connReq, 缓冲区大小为1非阻塞
				req := make(chan connReq, 1)
				cp.connReqs = append(cp.connReqs, req)
				// 加入队列后即可解锁
				cp.mu.Unlock()
				// 等待放回的 conn
				res, ok := <-req
				// chan 已经被关闭则返回
				if !ok {
					return nil, ErrMaxActiveConnReached
				}
				// 判断是否超时
				if timeout := cp.idleTimeout; timeout > 0 {
					if res.idleConn.startTime.Add(timeout).Before(time.Now()) {
						//丢弃并关闭该连接
						_ = cp.Close(res.idleConn.conn)
						continue
					}
				}
				return res.idleConn.conn, nil
			}
			// 未达到上限, 创建新连接
			if cp.factory == nil {
				cp.mu.Unlock()
				return nil, ErrConnClosed
			}
			conn, err := cp.factory.Factory()
			if err != nil {
				cp.mu.Unlock()
				return nil, fmt.Errorf("error occurs during creating initial conns: %+v", err)
			}
			// 增加连接数
			cp.openingConns++
			cp.mu.Unlock()
			return conn, nil
		}
	}
}

func (cp *connectionPool) Put(conn interface{}) error {
	if conn == nil {
		return ErrNilConn
	}

	cp.mu.Lock()
	defer cp.mu.Unlock()

	// 先判空
	if cp.conns == nil {
		return ErrConnClosed
	}
	idleConn := &idleConn{
		conn:      conn,
		startTime: time.Now(),
	}
	// 尝试从阻塞队列中唤醒
	if l := len(cp.connReqs); l > 0 {
		// 获取队头的 req
		req := cp.connReqs[0]
		cp.connReqs = cp.connReqs[:l-1]
		req <- connReq{idleConn: idleConn}
	}
	// 阻塞队列为空，则放入空闲队列
	select {
	case cp.conns <- idleConn:
		return nil
	default:
		// 空闲队列满了直接关闭连接
		return cp.Close(conn)
	}
}

func (cp *connectionPool) Close(conn interface{}) error {
	if conn == nil {
		return ErrNilConn
	}

	cp.mu.Lock()
	defer cp.mu.Unlock()

	if cp.conns == nil {
		return ErrConnClosed
	}
	cp.openingConns--

	return cp.factory.Close(conn)
}

func (cp *connectionPool) Release() {
	cp.mu.Lock()
	conns := cp.conns
	cp.conns = nil
	cp.mu.Unlock()

	// 最后关闭工厂
	defer func() {
		cp.factory = nil
	}()

	if conns == nil {
		return
	}

	close(conns)
	// 调用工厂的关闭方法关闭每个 conn
	for wrapConn := range conns {
		_ = cp.factory.Close(wrapConn.conn)
	}
}

func (cp *connectionPool) Len() int {
	return cp.openingConns
}
