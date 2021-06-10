package mq

import "log"

var done = make(chan bool)

//StartConsumer 获取消息
func StartConsumer(qName, cName string, callback func(msg []byte) bool) {
	//获取信道
	msgs, err := channel.Consume(
		qName,
		cName,
		true,  // auto-ack
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,   // args
	)

	if err != nil {
		log.Fatal(err.Error())
		return
	}

	//
	go func() {
		for msg := range msgs{
			procssSuc := callback(msg.Body)
			if !procssSuc {
				//TODO:将任务写入另一个队列，用于重试
			}
		}
	}()

	//阻塞开启
	<- done

	channel.Close()
}
