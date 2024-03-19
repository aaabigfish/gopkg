# nsq 

队列库，基于 [go-nsq](github.com/nsqio/go-nsq) 封装，默认配置在项目conf/config.toml，配置模版如下：
```
[nsq]
Addr = "192.168.1.157:34234"

```


# 生产者示例
```go
import "gitlab.ipcloud.cc/go/gopkg/mq/nsq"

// 创建写入链接池
// balancer平衡器，默认循环模式（RoundRobin）。在nsq集群上将消息路由到可用节点的逻辑，详见balancer.go
mq := nsq.NewProducerPool(addrs, balancer)

// 创建写入实例。
mq := nsq.NewProducer(addr)


// 创建生产者的时候指定了topic，可以使用Push方法，不需要指定topic值
mq.Push(body)

// body写入topic队列
mq.PushTopic("topic", body)

// 指定写入[]byte到某个队列,这里指定写入order_list1队列
mq.Publish("order_list1", []byte{"body"})

// 指定写入[]byte到延迟队列,这里指定写入order_list1队列
mq.PublishDelay("order_list1", []byte{"body"}， time.Second)

```


# 消费者示例
```go
import "gitlab.ipcloud.cc/go/gopkg/mq/nsq"

// 创建消费实例,可以传入nsq.Config配置
mq := nsq.NewConsumer(topic, channel, addrs, config)

// 增加处理函数
mq.AddHandler(hookFunc)


// 启动消费服务
mq.Run()

```