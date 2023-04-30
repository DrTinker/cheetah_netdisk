package mq

import (
	"NetDesk/common/conf"
	"NetDesk/common/models"

	"github.com/pkg/errors"
	"github.com/streadway/amqp"
)

// 保持connection的长连接，channel随线程的创建而创建
type MQClientImpl struct {
	conn *amqp.Connection
}

// 一个进程对应一个connection，一个线程对应一个channel
func NewMQClientImpl(url string) (*MQClientImpl, error) {
	// 开启connection
	mq := &MQClientImpl{}
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, err
	}
	mq.conn = conn
	return mq, nil
}

func (m *MQClientImpl) InitTransfer(exchange, key string) (*models.TransferSetting, error) {
	// 开启channel
	channel, err := m.conn.Channel()
	if err != nil {
		return nil, errors.Wrap(err, "[MQClientImpl] InitTransfer err:")
	}
	settings := &models.TransferSetting{
		Channel:   channel,
		Exchange:  exchange,
		RoutinKey: key,
	}
	return settings, nil
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
			ContentType: conf.Default_Content_Type,
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
