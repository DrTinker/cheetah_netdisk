package mq

import (
	"NetDisk/conf"
	"NetDisk/helper"
	"NetDisk/models"
	"fmt"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
)

// 保持connection的长连接，channel随线程的创建而创建
type MQClientImpl struct {
	url          string
	conn         *amqp.Connection
	closeReciver chan *amqp.Error
	blockReciver chan amqp.Blocking

	channelPool helper.Pool
}

// 连接池工厂类
type rabbitFactory struct {
	conn *amqp.Connection
}

func (rf rabbitFactory) Factory() (interface{}, error) {
	// 开启channel
	channel, err := rf.conn.Channel()
	if err != nil {
		return nil, errors.Wrap(err, "[rabbitFactory] Factory err:")
	}
	return channel, nil
}

func (rf rabbitFactory) Close(channel interface{}) error {
	if channel == nil {
		return errors.New("[rabbitFactory] Close nil channel")
	}
	ch, ok := channel.(*amqp.Channel)
	if !ok {
		return errors.New("[rabbitFactory] Close wrong channel type")
	}
	return ch.Close()
}

func (rf rabbitFactory) Ping(interface{}) error {
	return nil
}

// 一个进程对应一个connection，一个线程对应一个channel
func NewMQClientImpl(url string) (*MQClientImpl, error) {
	// 开启connection
	mq := &MQClientImpl{}
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, err
	}
	// 注入mq对象
	mq.url = url
	// 创建 tcp 连接和 channel pool
	mq.setConn(conn)

	return mq, nil
}

func (m *MQClientImpl) setConn(conn *amqp.Connection) {
	m.conn = conn
	// 注册关闭事件监听
	m.closeReciver = m.conn.NotifyClose(make(chan *amqp.Error))
	// 注册阻塞事件监听
	m.blockReciver = m.conn.NotifyBlocked(make(chan amqp.Blocking))
	// 创建 channel 的连接池
	rf := rabbitFactory{conn: conn}
	poolConfig := helper.PoolConfig{
		InitialCap: 5,
		Factory:    rf,
	}
	channelPool, err := helper.NewConnectionPool(poolConfig)
	for err != nil {
		logrus.Error(fmt.Sprintf("[MQClientImpl] setConn err: %+v", err))
		channelPool, err = helper.NewConnectionPool(poolConfig)
	}
	m.channelPool = channelPool
}

func (m *MQClientImpl) KeepAlive() {
	go func() {
		defer func() {
			if err := recover(); err != nil {
				fmt.Printf("Keep alive panic caught: %v\n", err)
				m.keepAlive()
			}
		}()
		m.keepAlive()
	}()
}

func (m *MQClientImpl) keepAlive() {
	for {
		logrus.Info("keep alive running")
		select {
		case close := <-m.closeReciver:
			// 不可恢复则输出日志
			if close != nil {
				logrus.Error(fmt.Sprintf("mq disconnected!!! code: %v reason: %v", close.Code, close.Reason))
			}

			// 如果是可以恢复的，则进行重连
			conn, err := amqp.Dial(m.url)
			for err != nil {
				conn, err = amqp.Dial(m.url)
				// 每秒尝试重连
				time.Sleep(time.Second)
			}
			m.setConn(conn)
			if m.conn != nil && m.channelPool != nil {
				logrus.Info("mq reconnected!!!")
			}

		case block := <-m.blockReciver:
			// 输出阻塞原因
			logrus.Warn("mq blocked by: ", block)
		}
	}
}

func (m *MQClientImpl) Publish(setting *models.TransferSetting, msg []byte) error {
	// 检查连接
	if m.conn.IsClosed() {
		return errors.Wrap(conf.MQConnectionClosedError, "[MQClientImpl] Publish err:")
	}
	// 发送消息
	// 从连接池获取 channel
	channel, err := m.channelPool.Get()
	ch, ok := channel.(*amqp.Channel)
	if err != nil || !ok {
		return errors.Wrap(err, "[MQClientImpl] Publish err:")
	}
	// 放回连接池
	defer m.channelPool.Put(ch)
	err = ch.Publish(
		setting.Exchange,
		setting.RoutinKey,
		false, // 消息无法正确被路由则丢弃
		false, // 参数不起作用，原因未知
		amqp.Publishing{
			ContentType: conf.DefaultContentType,
			Body:        msg,
		},
	)
	if err != nil {
		return errors.Wrap(err, "[MQClientImpl] Publish err:")
	}

	return nil
}

func (m *MQClientImpl) Consume(setting *models.TransferSetting, queue, consumer string, callback func(msg []byte) bool) error {
	// 检查连接
	if m.conn.IsClosed() {
		return errors.Wrap(conf.MQConnectionClosedError, "[MQClientImpl] Consume err:")
	}
	// 从连接池获取 channel
	channel, err := m.channelPool.Get()
	ch, ok := channel.(*amqp.Channel)
	if err != nil || !ok {
		return errors.Wrap(err, "[MQClientImpl] Consume err:")
	}
	// 放回连接池
	defer m.channelPool.Put(ch)
	msgs, err := ch.Consume(
		queue,
		consumer,
		true,  // autoACK
		false, // exclusive
		false, // nolocal
		false, // nowait
		nil,
	)
	if err != nil {
		return errors.Wrap(err, "[MQClientImpl] Consume err:")
	}

	wg := sync.WaitGroup{}
	for msg := range msgs {
		wg.Add(1)
		// 每个消息开启一个协程处理
		go func(msg amqp.Delivery) {
			if success := callback(msg.Body); !success {
				// TODO 失败转入死信队列
			}
			wg.Done()
		}(msg)
	}

	wg.Wait()

	return nil
}
