# log

日志库，基于 [zap](https://github.com/uber-go/zap) 封装， 依赖gitlab.ipcloud.cc/go/gopkg/config 默认配置在项目conf/config.toml，配置模版如下：
```
[log]
# 日志(debug,info,warn,error,fatal,panic)
Level = "debug"
# 文件名
FileName = "logs/info.log"
# 单个日志文件大小MB
MaxSize = 500
# 至多保留多少个日志文件
MaxBackups = 20
# 至多保留多少天的日志文件
MaxAge = 30
# 压缩
Compress = true
# 本地时间
LocalTime = true
# 是否打印到控制台,true打印到控制台，false记录到文件
Console = false
```

# 示例
```go
import "gitlab.ipcloud.cc/go/gopkg/log"

// 调试信息打印,可以打印多个
log.PP(order, err, config.Config)

// 采用键值对的方式打印
log.Info("msg", key1, val1, key2, val2)

// 采用printf方式打印
log.Infof("msg(%+v) err(%v)", msg, err)

// 新建键值对"key", val的日志
log.With("key", val ).Errorf("not found err(%v)", err)

// 使用zap的日志
log.New().Error("file not found", zap.Any("file", "config.json"))
```
