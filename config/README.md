# config
读取程序根目录的conf下所有 toml和json， 默认从config.toml加载配置。

注意：
- 目前只支持json和toml文件
- 只会读取conf下的所有json和toml文件，不支持目录递归
- 只能设置 k-v 型配置
- 配置名不区分大小写字母


框架还会自动监听conf目录下所有 toml和json 内容变更，发现变更会自动加载。

指定配置文件路径，可以设置环境变量CONF_PATH
linux环境配置
```
export CONF_PATH="/project/root/admin"
```

window环境配置
```
set CONF_PATH="/project/root/admin"
```

# 示例
```go
import "gitlab.ipcloud.cc/go/gopkg/config"

func init() {
    // OnConfigChange 注册配置文件变更回调
    config.OnConfigChange(func() {})
    // WatchConfig 启动配置变更监听，业务代码不要调用。
    config.WatchConfig()
}

// 获取默认配置文件config.toml，[app]段的AppName配置
appName := config.Get("app.AppName")

// 获取文件foo.toml的WORKER_NUM配置
b := config.File("foo.toml").GetInt32("WORKER_NUM")
```

