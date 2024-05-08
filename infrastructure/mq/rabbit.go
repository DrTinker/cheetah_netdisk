package mq

import (
	"NetDisk/conf"
	"NetDisk/models"
	"fmt"
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
}

// 一个进程对应一个connection，一个线程对应一个channel
func NewMQClientImpl(url string) (*MQClientImpl, error) {
	// 开启connection
	mq := &MQClientImpl{}
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, err
	}
	// 创建队列
	// 监听机制
	closeReciver := make(chan *amqp.Error)
	blockReciver := make(chan amqp.Blocking)
	// 注入mq对象
	mq.url = url
	mq.conn = conn
	mq.closeReciver = closeReciver
	mq.blockReciver = blockReciver
	// 注册关闭事件监听
	mq.conn.NotifyClose(mq.closeReciver)
	// 注册阻塞事件监听
	mq.conn.NotifyBlocked(mq.blockReciver)
	return mq, nil
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
		//logrus.Warn("keep alive running")
		select {
		case close := <-m.closeReciver:
			// 不可恢复则输出日志
			logrus.Error(fmt.Sprintf("mq disconnected!!! code: %v reason: %v", close.Code, close.Reason))
			// 如果是可以恢复的，则进行重连
			conn, err := amqp.Dial(m.url)
			for err != nil {
				conn, err = amqp.Dial(m.url)
				// 每秒尝试重连
				time.Sleep(time.Second)
			}
			m.conn = conn
			if conn != nil {
				logrus.Info("mq reconnected!!!")
			}
		case block := <-m.blockReciver:
			// 输出阻塞原因
			logrus.Warn("mq blocked by: ", block)
		default:
			// do nothing
		}
	}
}

func (m *MQClientImpl) InitTransfer(exchange, key string) (*models.TransferSetting, error) {
	// 开启channel
	channel, err := m.conn.Channel()
	if err != nil {
		return nil, errors.Wrap(err, "[MQClientImpl] InitTransfer err:")
	}
	// 不要在代码中声明队列，交换机，binding 参考：https://juejin.cn/post/7125719003510603783
	// 初始化队列，交换机，binding规则
	// 声明交换机
	// err = channel.ExchangeDeclare(conf.Exchange, "direct",
	// 	true,  // 持久化
	// 	false, // 自动删除
	// 	false, // 内置交换机
	// 	true,  // noWait
	// 	nil,   // 其他配置
	// )
	// if err != nil {
	// 	return nil, errors.Wrap(err, "[MQClientImpl] InitTransfer declare exchange err:")
	// }
	// // 声明队列，存在则忽视
	// qInfo, err := channel.QueueDeclare(conf.TransferCOSQueue,
	// 	true,  // 持久化
	// 	false, // 自动删除，false指断开connection本队列不会自动删除
	// 	false, // 排他性，为true只为本connection中channel共享，conn断开后自动删除
	// 	true,  // noWait true表示创建不需要等服务器确认
	// 	nil,   // 其他配置，暂时不用
	// )
	// logrus.Info("[MQClientImpl] InitTransfer queue info: ", qInfo)
	// if err != nil {
	// 	return nil, errors.Wrap(err, "[MQClientImpl] InitTransfer declare queue err:")
	// }
	// // 绑定交换机和队列
	// channel.QueueBind()
	settings := &models.TransferSetting{
		Channel:   channel,
		Exchange:  exchange,
		RoutinKey: key,
	}
	return settings, nil
}

// 一个线程一个channel，用完关闭
func (m *MQClientImpl) ReleaseChannel(s *models.TransferSetting) {
	if s == nil || s.Channel == nil {
		return
	}
	err := s.Channel.Close()
	if err != nil {
		logrus.Error("channel close err: ", err)
	}
}

func (m *MQClientImpl) Publish(setting *models.TransferSetting, msg []byte) error {
	// 检查连接
	if m.conn.IsClosed() {
		return errors.Wrap(conf.MQConnectionClosedError, "[MQClientImpl] Publish err:")
	}
	// 发送消息
	err := setting.Channel.Publish(
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
	channel := setting.Channel
	msgs, err := channel.Consume(
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

	// 用于阻塞循环
	done := make(chan bool)

	go func() {
		for msg := range msgs {
			if success := callback(msg.Body); !success {
				// TODO 失败转入死信队列
			}
		}
	}()
	// 循环监听消息队列
	<-done
	// 没有消息则结束
	err = channel.Close()
	if err != nil {
		return errors.Wrap(err, "[MQClientImpl] close channel err:")
	}
	return nil
}
