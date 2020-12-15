#docker安装主从mysql

#####master节点
```
mkdir -p /data/mysql/datam
docker run -d --name mysql-master -p 3306:3306 -v master.conf:/etc/mysql/mysql.conf.d/mysqld.cnf -v /data/mysql/datam:/var/lib/mysql  -e MYSQL_ROOT_PASSWORD=root mysql:5.7

#####运行参数说明

–name mysql-master: 容器的名称设为mysql-master

-p 3306:3306: 将host的13306端口映射到容器的3306端口

-v master.conf:/etc/mysql/mysql.conf.d/mysqld.cnf ： master.conf配置文件挂载

-v /data/mysql/datam:/var/lib/mysql ： mysql容器内数据挂载到host的/data/mysql/datam， 用于持久化

-e MYSQL_ROOT_PASSWORD=root : mysql的root登录密码为root
```

#####slave节点
```
mkdir  -p /data/mysql/datas
docker run -d --name mysql-slave -p 3307:3306 -v slave.conf:/etc/mysql/mysql.conf.d/mysqld.cnf -v /data/mysql/datas:/var/lib/mysql  -e MYSQL_ROOT_PASSWORD=root mysql:5.7

#####运行参数说明

–name mysql-slave: 容器的名称设为mysql-slave

-p 3307:3306: 将host的13307端口映射到容器的3306端口

-v slave.conf:/etc/mysql/mysql.conf.d/mysqld.cnf ： slave.conf配置文件挂载

-v /data/mysql/datas:/var/lib/mysql ： mysql容器内数据挂载到host的/data/mysql/datas， 用于持久化

-e MYSQL_ROOT_PASSWORD=root : mysql的root登录密码为root
```

#####登陆master节点

```
mysql> create user slave identified by 'slave';
mysql> GRANT REPLICATION SLAVE ON *.* TO 'slave'@'%' IDENTIFIED BY 'slave';
mysql> flush privileges;
mysql> create database fileserver default character set utf8mb4;

再获取status, 得到类似如下的输出:
mysql> show master status \G;
*************************** 1. row ***************************
             File: log.000025
         Position: 155
     Binlog_Do_DB: 
 Binlog_Ignore_DB: 
Executed_Gtid_Set: 
1 row in set (0.00 sec)


查看mysql master的容器独立ip地址
docker inspect --format='{{.NetworkSettings.IPAddress}}' mysql-master
# 比如输出得到: 172.17.0.2
```

#####登陆slave节点

```
mysql> stop slave;
mysql> create database fileserver default character set utf8mb4;
#注意其中的日志文件和数值要和上面show master status的值对应
mysql> CHANGE MASTER TO MASTER_HOST='前两个步骤中获得的mysql master ip',MASTER_PORT=3306,MASTER_USER='slave',MASTER_PASSWORD='slave',MASTER_LOG_FILE='log.000025',MASTER_LOG_POS=155;
mysql> start slave;

mysql> show slave status \G;
// ...
Slave_IO_Running: Yes 
Slave_SQL_Running: Yes 
// ...

此时主从配置已完成
如有问题可参考
https://blog.csdn.net/weixin_39359455/article/details/104220723
```


