# gopkg


## 目录

- [简介](#简介)
- [如何使用](#如何使用)
- [目录说明](#目录说明)
- [License](#License)

## 简介

`gopkg` 是 Go 的通用实用程序集合
## 如何使用

```go
go get -u github.com/go/gopkg@master
```

## 目录说明
```
├── README.md
├── cache
│   ├── asynccache
│   ├── mcache
│   └── redis
├── cloud
│   ├── circuitbreaker
│   └── metainfo
├── collection
│   ├── hashset
│   ├── lscq
│   ├── skipmap
│   ├── skipset
│   └── zset
├── config
├── database
│   ├── mongo
│   └── orm
├── go.mod
├── internal
│   ├── benchmark
│   ├── hack
│   ├── runtimex
│   └── wyhash
├── lang
│   ├── fastrand
│   ├── stringx
│   └── syncx
├── log
├── stat
│   └── counter
├── timex
└── util
    ├── gctuner
    ├── gopool
    └── xxhash3
```

## License

`gopkg` is licensed under the terms of the Apache license 2.0. See [LICENSE](LICENSE) for more information.