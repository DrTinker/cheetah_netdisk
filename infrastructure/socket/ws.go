package socket

import (
	"NetDesk/conf"
	"fmt"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type SocketClientImpl struct {
	upgrader websocket.Upgrader
	// websocket客户端链接池
	client map[string]*websocket.Conn
	// 互斥锁，防止程序对统一资源同时进行读写
	mutex sync.Mutex
}

func NewSocketClientImpl() *SocketClientImpl {
	// websocket Upgrader
	wsupgrader := websocket.Upgrader{
		ReadBufferSize:   conf.Buffer_Size,
		WriteBufferSize:  conf.Buffer_Size,
		HandshakeTimeout: conf.Handshake_Timeout,
		// 取消ws跨域校验
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	return &SocketClientImpl{
		upgrader: wsupgrader,
		client:   make(map[string]*websocket.Conn),
	}
}

// 向连接池中添加ws链接
func (s *SocketClientImpl) AddConn(w http.ResponseWriter, r *http.Request, id string) error {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		return errors.Wrap(err, "[NewSocketClientImpl] AddSocketConn connect err: ")
	}
	s.mutex.Lock()
	s.client[id] = conn
	s.mutex.Unlock()
	return nil
}

// 发送数据
func (s *SocketClientImpl) SendMsg(id string, msg interface{}) error {
	s.mutex.Lock()
	conn, ok := s.client[id]
	s.mutex.Unlock()
	if !ok || conn == nil {
		return errors.Wrap(conf.SocketNilError, fmt.Sprintf("[NewSocketClientImpl] id: %s\n", id))
	}

	err := conn.WriteJSON(msg)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("[NewSocketClientImpl] send msg id: %s\n", id))
	}
	return nil
}

// msg应传指针类型
func (s *SocketClientImpl) ReadMsg(id string, msg interface{}) error {
	s.mutex.Lock()
	conn, ok := s.client[id]
	s.mutex.Unlock()
	if !ok || conn == nil {
		return errors.Wrap(conf.SocketNilError, fmt.Sprintf("[NewSocketClientImpl] id: %s\n", id))
	}

	err := conn.ReadJSON(msg)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("[NewSocketClientImpl] read msg id: %s\n", id))
	}
	return nil
}

// 检查服务端连接是否正常
func (s *SocketClientImpl) CheckConn(id string) bool {
	s.mutex.Lock()
	conn, ok := s.client[id]
	s.mutex.Unlock()
	if conn == nil || !ok {
		return false
	}
	return true
}

// 从连接池中清除
func (s *SocketClientImpl) DeleteConn(id string) {
	s.mutex.Lock()
	conn := s.client[id]
	if conn != nil {
		conn.Close()
	}
	delete(s.client, id)
	s.mutex.Unlock()
	logrus.Info("[NewSocketClientImpl] socket: ", id, " has been deleted")
}
