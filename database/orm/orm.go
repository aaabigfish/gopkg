package orm

import (
	"sync"
	"time"

	"golang.org/x/sync/singleflight"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"github.com/aaabigfish/gopkg/config"
)

var (
	sfg singleflight.Group
	rwl sync.RWMutex

	dbs = map[string]*DB{}
)

// DB 扩展 sqlx.DB
type DB struct {
	*gorm.DB
}

// Get 获取数据库实例
// db := mysql.Get("foo")
func Get(name string, conf ...*gorm.Config) *DB {
	rwl.RLock()
	if db, ok := dbs[name]; ok {
		rwl.RUnlock()
		return db
	}
	rwl.RUnlock()

	v, _, _ := sfg.Do(name, func() (interface{}, error) {
		dbConf := config.NewDbConfig(name)
		if len(dbConf.Type) == 0 {
			dbConf.Type = "mysql"
		}

		c := &gorm.Config{}
		if len(conf) > 0 {
			c = conf[0]
		}

		db, err := gorm.Open(mysql.Open(dbConf.Dsn), c)
		if err != nil {
			panic(err.Error())
		}

		if config.DbMode == "debug" {
			db = db.Debug()
		}

		sdb, err := db.DB()
		if err != nil {
			panic(err.Error())
		}
		sdb.SetMaxIdleConns(dbConf.MaxIdle)
		sdb.SetMaxOpenConns(dbConf.MaxActive)
		sdb.SetConnMaxLifetime(time.Duration(dbConf.IdleTimeout) * time.Second)

		newDB := &DB{db}

		rwl.Lock()
		defer rwl.Unlock()
		dbs[name] = newDB

		return newDB, nil
	})

	return v.(*DB)
}
