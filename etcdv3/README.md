# etcdv3

etcdv3库，基于 [etcdv3](go.etcd.io/etcd/client/v3) 封装，默认配置在项目conf/config.toml，配置模版如下：
```
[etcd]
Endpoints = ["192.168.1.157:34234"]
Username = ""
Password = ""

```


# 示例
```go
import "gitlab.ipcloud.cc/go/gopkg/etcdv3"

client := etcdv3.NewClient(context.Background, []string{"addr:port1", "addr2:port2", "addr3:port3"}, etcdv3.ClientOptions{})

client.Put("test_key", "value")
client.Get("test_key")

```