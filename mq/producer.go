package mq

import (
	"github.com/streadway/amqp"
	cfg "goWork/config"
	"log"
)

var coon *amqp.Connection
var channel *amqp.Channel

//异常关闭会接收通知
var notifyClose chan *amqp.Error

func init() {
	//判断是否需要异步
	if !cfg.AsyncTransferEnable {
		return
	}

	if initChannel() {
		channel.NotifyClose(notifyClose)
	}

	//断线重连
	go func() {
		for {
			select {
			case msg := <-notifyClose:
				coon = nil
				channel = nil
				log.Printf("onNotifyChannelClosed: %+v\n", msg)
				initChannel()
			}
		}
	}()
}

//initChannel 初始化连接rabbitmq
func initChannel() bool {
	//判断channel是否为空
	if channel != nil {
		return true
	}

	//获取rabbitmq链接
	coon, err := amqp.Dial(cfg.RabbitUrl)
	if err != nil {
		log.Fatal(err.Error())
		return false
	}

	//打开channel,接收消息与发送消息
	channel, err = coon.Channel()
	if err != nil {
		log.Fatal(err.Error())
		return false
	}

	return true
}

func Publish(exchange, routingKey string, msg []byte) bool {
	//判断连接是否正常
	if !initChannel() {
		return false
	}

	//执行消息发布动作
	err := channel.Publish(
		exchange,
		routingKey,
		false,
		false,
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        msg,
		},
	)
	if err != nil {
		log.Fatal(err.Error())
		return false
	}
	return true
}
