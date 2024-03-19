package log

import (
	"os"
	"time"

	"github.com/k0kubun/pp/v3"
	"github.com/mattn/go-isatty"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"

	"gitlab.ipcloud.cc/go/gopkg/config"
)

var (
	_logger *logger
)

type logger struct {
	cfg    *config.LogConfig
	sugar  *zap.SugaredLogger
	_level zapcore.Level
}

func initPP() {
	out := os.Stdout
	pp.SetDefaultOutput(out)

	if !isatty.IsTerminal(out.Fd()) {
		pp.ColoringEnabled = false
	}
}

func init() {
	initPP()

	_cfg := config.NewLogConfig()

	_logger = &logger{
		cfg: _cfg,
	}
	lumber := _logger.newLumber()
	writeSyncer := zapcore.NewMultiWriteSyncer(zapcore.AddSync(lumber))
	zapOpt := zap.Fields(zap.String("appname", config.App))
	if config.Env != config.EnvProd {
		zapOpt = zap.Fields(zap.String("appname", config.App), zap.String("env", config.Env))
	}

	sugar := zap.New(_logger.newCore(writeSyncer),
		zap.ErrorOutput(writeSyncer),
		zap.AddCaller(),
		zap.AddCallerSkip(1),
		zapOpt).
		Sugar()

	_logger.sugar = sugar

}

// PP 类似 PHP 的 var_dump
func PP(args ...interface{}) {
	pp.Println(args...)
}

func (l *logger) newCore(ws zapcore.WriteSyncer) zapcore.Core {
	// 默认日志级别
	atomicLevel := zap.NewAtomicLevel()
	defaultLevel := zapcore.DebugLevel
	// 会解码传递的日志级别，生成新的日志级别
	_ = (&defaultLevel).UnmarshalText([]byte(l.cfg.Level))
	atomicLevel.SetLevel(defaultLevel)
	l._level = defaultLevel

	// encoder 这部分没有放到配置文件，因为一般配置一次就不会改动
	encoder := zapcore.EncoderConfig{
		MessageKey:     "msg",
		LevelKey:       "level",
		TimeKey:        "time",
		NameKey:        "logger",
		CallerKey:      "file",
		StacktraceKey:  "stack",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalColorLevelEncoder,
		EncodeTime:     l.customTimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
		EncodeName:     zapcore.FullNameEncoder,
	}
	var writeSyncer zapcore.WriteSyncer
	if l.cfg.Console {
		writeSyncer = zapcore.NewMultiWriteSyncer(zapcore.AddSync(os.Stdout))
		return zapcore.NewCore(zapcore.NewConsoleEncoder(encoder),
			writeSyncer,
			atomicLevel)
	}

	// 输出到文件时，不使用彩色日志，否则会出现乱码
	encoder.EncodeLevel = zapcore.LowercaseLevelEncoder
	writeSyncer = ws
	return zapcore.NewCore(zapcore.NewJSONEncoder(encoder),
		writeSyncer,
		atomicLevel)
}

// CustomTimeEncoder 实现了 zapcore.TimeEncoder
// 实现对日期格式的自定义转换
func (l *logger) customTimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	format := l.cfg.TimeFormat
	if len(format) <= 0 {
		format = "2006-01-02 15:04:05.000"
	}
	enc.AppendString(t.Format(format))
}

func (l *logger) newLumber() *lumberjack.Logger {
	return &lumberjack.Logger{
		Filename:   l.cfg.FileName,
		MaxSize:    l.cfg.MaxSize,
		MaxAge:     l.cfg.MaxAge,
		MaxBackups: l.cfg.MaxBackups,
		LocalTime:  l.cfg.LocalTime,
		Compress:   l.cfg.Compress,
	}
}

// Fields fields
type Field = zap.Field

func With(key string, val interface{}) *zap.SugaredLogger {
	return _logger.sugar.With(key, val)
}

func New() *zap.SugaredLogger {
	return _logger.sugar
}

func Any(key string, val interface{}) Field {
	return zap.Any(key, val)
}

// Debug 打印debug级别信息
func Debug(message string, kvs ...interface{}) {
	_logger.sugar.Debugw(message, kvs...)
}

// Info 打印info级别信息
func Info(message string, kvs ...interface{}) {
	_logger.sugar.Infow(message, kvs...)
}

// Warn 打印warn级别信息
func Warn(message string, kvs ...interface{}) {
	_logger.sugar.Warnw(message, kvs...)
}

// Error 打印error级别信息
func Error(message string, kvs ...interface{}) {
	_logger.sugar.Errorw(message, kvs...)
}

// Panic 打印错误信息，然后panic
func Panic(message string, kvs ...interface{}) {
	_logger.sugar.Panicw(message, kvs...)
}

// Fatal 打印错误信息，然后退出
func Fatal(message string, kvs ...interface{}) {
	_logger.sugar.Fatalw(message, kvs...)
}

// Debugf 格式化输出debug级别日志
func Debugf(template string, args ...interface{}) {
	_logger.sugar.Debugf(template, args...)
}

// Infof 格式化输出info级别日志
func Infof(template string, args ...interface{}) {
	_logger.sugar.Infof(template, args...)
}

// Warnf 格式化输出warn级别日志
func Warnf(template string, args ...interface{}) {
	_logger.sugar.Warnf(template, args...)
}

// Errorf 格式化输出error级别日志
func Errorf(template string, args ...interface{}) {
	_logger.sugar.Errorf(template, args...)
}

// Panicf 格式化输出日志，并panic
func Panicf(template string, args ...interface{}) {
	_logger.sugar.Panicf(template, args...)
}

// Fatalf 格式化输出日志，并退出
func Fatalf(template string, args ...interface{}) {
	_logger.sugar.Fatalf(template, args...)
}

// Sync 关闭时需要同步日志到输出
func Sync() {
	_ = _logger.sugar.Sync()
}
