# 如何使用

- 环境变量(debug模式打印所有的sql，本地开发使用)
```
export DB_MODE=debug
```

- 引入orm包
```go
import "gitlab.ipcloud.cc/go/gopkg/database/orm"
```

- 配置数据库连接
```
# DB链接配置，格式为 db_${name}，通过 ${name} 可以获取DB连接池
[db_buffnetwork]
# 格式如[username[:password]@][protocol[(address)]]/dbname[?param1=value1&...&paramN=valueN]
Dsn = "mylock_root-test:mylock_root123-test@tcp(35.78.219.246:3306)/buff_network?charset=utf8mb4&parseTime=true&loc=Asia%2FShanghai"

```

- 实例化mysql对象,name和配置文件的${name}对应，针对上面的配置name=buffnetwork
```go
db := orm.Get("buffnetwork")
```