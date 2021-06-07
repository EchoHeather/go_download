$mkdir -p /www/rabbitmq
$docker run -d --hostname rabbit-node1 --name rabbit-node1 -p 5672:5672 -p15672:15672 -v /www/rabbitmq:/var/lib/rabbitmq rabbitmq:management

$docker ps | grep rabbit
c676e3890c88        rabbitmq:management   "docker-entrypoint.s…"   3 minutes ago       Up 3 minutes        4369/tcp, 5671/tcp, 0.0.0.0:5672->5672/tcp, 15671/tcp, 25672/tcp, 0.0.0.0:15672->15672/tcp   rabbit-node1

浏览器打开登录rabbitmq, 入口:http://localhost:15672
默认用户名: guest 密码: guest