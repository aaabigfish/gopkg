# asynq 

队列库，基于 [asynq](github.com/hibiken/asynq) 封装， 依赖gitlab.ipcloud.cc/go/gopkg/config 默认配置在项目conf/config.toml，配置模版如下：
```
[asynq]
# redis地址，格式（host:port）
Addr = "192.168.1.157:6379"
# redis用户名
UserName = ""
# redis密码
PassWord = ""
# redis的DB
DB = 3
```


# 生产者示例
```go
import "gitlab.ipcloud.cc/go/gopkg/mq/asynq"

// 创建写入实例,可以指定写入的redis队列的key,这里以order_list为例,(注意key值必须全局唯一，以免出现出现消费错乱问题)
mq := asynq.NewWriter("order_list")

// 如果初始化队列，已经初始化了写入队列的key，可以使用Push方法，不需要指定key值
mq.Push(order)

// 指定写入某个队列,这里指定写入list_key队列
mq.Send("list_key", "test")

```


# 消费者示例
```go
import "gitlab.ipcloud.cc/go/gopkg/mq/asynq"

// 创建消费实例,可以传入asynq.Config的精细配置队列
mq := asynq.NewReader()

// 增加处理函数（注意key不能重复，重复会导致函数覆盖）
mq.AddHook("key", func(ctx context.Context, m*asynq.Message) error {
	log.PP(string(m.Payload()))
	return nil
})

// 批量增加处理函数
mq.AddHooks(Handler{
    "key1":handlerFunc1,
    "key2":handlerFunc2,
})

// 启动消费服务
mq.Run()

```