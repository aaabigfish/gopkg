// Package config 提供最基础的配置加载功能
package config

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/cast"
	"github.com/spf13/viper"
)

const (
	EnvDev    = "dev"
	EnvTest   = "test"
	EnvMirror = "mirror"
	EnvProd   = "prod"
)

// Get 查询配置/环境变量
var Get func(string) string

var (
	// Host 主机名
	Host = "localhost"
	// App 服务标识
	App = "localapp"
	// Env 运行环境
	Env = EnvProd
	// Zone 服务区域
	Zone = "sh001"

	DbMode = "release"

	files = map[string]*Client{}

	defaultFile = "config.toml"

	configErr error
)

type DBConfig struct {
	Type        string
	Dsn         string
	MaxIdle     int
	MaxActive   int
	IdleTimeout int
}

// LogConfig is used to configure uber zap
type LogConfig struct {
	Level      string
	FileName   string
	TimeFormat string
	MaxSize    int
	MaxBackups int
	MaxAge     int
	Compress   bool
	LocalTime  bool
	Console    bool
}

func init() {
	Host, _ = os.Hostname()
	if appID := os.Getenv("APP_ID"); appID != "" {
		App = appID
	}

	if env := os.Getenv("ENV"); env != "" {
		Env = env
	}

	if zone := os.Getenv("ZONE"); zone != "" {
		Zone = zone
	}

	if name := os.Getenv("CONF_NAME"); name != "" {
		defaultFile = name
	}

	if dbmode := os.Getenv("DB_MODE"); dbmode != "" {
		DbMode = dbmode
	}

	path := os.Getenv("CONF_PATH")
	if path == "" {
		var err error
		path, err = os.Getwd()
		if err != nil {
			configErr = err
			return
		}
	}

	configPath := filepath.Join(path, "conf")
	if !fileExists(configPath) {
		appPath, err := filepath.Abs(filepath.Dir(os.Args[0]))
		if err != nil {
			configErr = err
			return
		}
		configPath = filepath.Join(appPath, "conf")
	}

	fs, err := os.ReadDir(configPath)
	if err != nil {
		return
	}

	for _, f := range fs {
		if !(strings.HasSuffix(f.Name(), ".toml") || strings.HasSuffix(f.Name(), ".json")) {
			continue
		}

		v := viper.New()
		v.SetConfigFile(filepath.Join(configPath, f.Name()))
		if err := v.ReadInConfig(); err != nil {
			configErr = err
			return
		}
		v.AutomaticEnv()

		files[f.Name()] = &Client{v}
	}

	if File(defaultFile) == nil {
		return
	}

	Get = GetString

	if appName := GetString("app.appname"); appName != "" {
		App = appName
	}

	if env := GetString("app.env"); env != "" {
		Env = env
	}

	if GetBool("log.Console") {
		DbMode = "debug"
	}
}

func Error() string {
	return configErr.Error()
}

func NewDbConfig(key string) *DBConfig {
	maxIdle := 10
	maxActive := 1000
	idleTimeout := 600
	if v := GetInt("db_" + key + ".maxidle"); v > 0 {
		maxIdle = v
	}
	if v := GetInt("db_" + key + ".maxactive"); v > 0 {
		maxActive = v
	}
	if v := GetInt("db_" + key + ".idletimeout"); v > 0 {
		idleTimeout = v
	}

	return &DBConfig{
		Type:        GetString("db_" + key + ".type"),
		Dsn:         GetString("db_" + key + ".dsn"),
		MaxIdle:     maxIdle,
		MaxActive:   maxActive,
		IdleTimeout: idleTimeout,
	}
}

func NewLogConfig() *LogConfig {
	cfg := &LogConfig{
		Level:     "debug",
		LocalTime: true,
		Console:   true,
	}

	if File(defaultFile) == nil {
		return cfg
	}

	return &LogConfig{
		Level:      GetString("log.Level"),
		FileName:   GetString("log.FileName"),
		TimeFormat: GetString("log.TimeFormat"),
		MaxSize:    GetInt("log.MaxSize"),
		MaxBackups: GetInt("log.MaxBackups"),
		MaxAge:     GetInt("log.MaxAge"),
		Compress:   GetBool("log.Compress"),
		LocalTime:  GetBool("log.LocalTime"),
		Console:    GetBool("log.Console"),
	}
}

func fileExists(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

type Client struct {
	*viper.Viper
}

// File 根据文件名获取对应配置对象
// 目前仅支持 toml和json 文件
// 如果要读取 foo.toml 配置，可以 File("foo.toml").Get("bar")
func File(name string) *Client {
	return files[name]
}

// OnConfigChange 注册配置文件变更回调
// 需要在 WatchConfig 之前调用
func OnConfigChange(run func()) {
	for _, v := range files {
		v.OnConfigChange(func(in fsnotify.Event) { run() })
	}
}

// WatchConfig 启动配置变更监听，业务代码不要调用。
func WatchConfig() {
	for _, v := range files {
		v.WatchConfig()
	}
}

// Set 设置配置，仅用于测试
func Set(key string, value interface{}) { File(defaultFile).Set(key, value) }

func GetBool(key string) bool              { return File(defaultFile).GetBool(key) }
func GetDuration(key string) time.Duration { return File(defaultFile).GetDuration(key) }
func GetFloat64(key string) float64        { return File(defaultFile).GetFloat64(key) }
func GetInt(key string) int                { return File(defaultFile).GetInt(key) }
func GetInt32(key string) int32            { return File(defaultFile).GetInt32(key) }
func GetInt64(key string) int64            { return File(defaultFile).GetInt64(key) }
func GetIntSlice(key string) []int         { return File(defaultFile).GetIntSlice(key) }
func GetSizeInBytes(key string) uint       { return File(defaultFile).GetSizeInBytes(key) }
func GetString(key string) string          { return File(defaultFile).GetString(key) }
func GetStringSlice(key string) []string   { return File(defaultFile).GetStringSlice(key) }
func GetTime(key string) time.Time         { return File(defaultFile).GetTime(key) }
func GetUint(key string) uint              { return File(defaultFile).GetUint(key) }
func GetUint32(key string) uint32          { return File(defaultFile).GetUint32(key) }
func GetUint64(key string) uint64          { return File(defaultFile).GetUint64(key) }

func GetStringMap(key string) map[string]interface{} { return File(defaultFile).GetStringMap(key) }

func GetStringMapInt(key string) map[string]int {
	stringMap := File(defaultFile).GetStringMap(key)
	mapInt := make(map[string]int, len(stringMap))
	for key, val := range stringMap {
		mapInt[key] = cast.ToInt(val)
	}
	return mapInt
}

func GetStringMapString(key string) map[string]string {
	return File(defaultFile).GetStringMapString(key)
}
func GetStringMapStringSlice(key string) map[string][]string {
	return File(defaultFile).GetStringMapStringSlice(key)
}
