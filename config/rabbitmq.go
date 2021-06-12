package config

const (
	//开启为异步
	AsyncTransferEnable = true
	//rabbitmq host
	RabbitUrl = "amqp://guest:guest@localhost:5672/"
	//交换机name
	TransExchangeName = "uploadsearver.trans"
	//上传视频异步队列名
	TransQueueName = "uploadsearver.trans.oss"
	//上传视频失败异步队列名
	TransQueueErrName = "uploadsearver.trans.oss.err"
	//Routing key
	TransRoutingkey = "oss"
)
