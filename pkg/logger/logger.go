package logger

import (
	"log"
	"os"
	"path"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type (
	// Config is logger config.
	Config struct {
		// SystemLogfilePath is the path of logfile,
		// if it is empty, it would be ./logs/containeragent.log
		SystemLogfilePath string
		// Debug specifies the lowest log level is Debug if true,
		// otherwise Info is the one.
		Debug bool

		finalSystemLogfilePath string
	}
)

// Init initializes logger.
func Init(config *Config) {
	config.finalSystemLogfilePath = config.SystemLogfilePath
	if config.finalSystemLogfilePath == "" {
		config.finalSystemLogfilePath = diskSystemLoggerFilePath
	}

	initSystem(config)
}

const (
	diskSystemLoggerFilePath = "logs/containeragent.log"

	// no cache for system log
	systemLogMaxCacheCount = 0

	// RFC3339Milli is the logger time format.
	RFC3339Milli = "2006-01-02T15:04:05.999Z07:00"
)

var (
	compoundSystemLogger *zap.SugaredLogger // equal stderrLogger + diskSystemLogger
	stderrSystemLogger   *zap.SugaredLogger
	diskSystemLogger     *zap.SugaredLogger
)

func systemEncoderConfig() zapcore.EncoderConfig {
	timeEncoder := func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(t.Format(RFC3339Milli))
	}

	return zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "", // no need
		CallerKey:      "caller",
		MessageKey:     "message",
		StacktraceKey:  "", // no need
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalColorLevelEncoder,
		EncodeTime:     timeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}
}

func initSystem(config *Config) {
	encoderConfig := systemEncoderConfig()

	lowestLevel := zap.InfoLevel
	if config.Debug {
		lowestLevel = zap.DebugLevel
	}

	dir, _ := path.Split(config.finalSystemLogfilePath)

	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err = os.Mkdir(dir, os.ModePerm)
		if err != nil {
			log.Printf(err.Error())
			os.Exit(1)
		}
	}

	lf, err := newLogFile(config.finalSystemLogfilePath, systemLogMaxCacheCount)
	if err != nil {
		log.Printf(err.Error())
		os.Exit(1)
	}

	opts := []zap.Option{zap.AddCaller(), zap.AddCallerSkip(1)}

	stderrSyncer := zapcore.AddSync(os.Stderr)
	stderrCore := zapcore.NewCore(zapcore.NewConsoleEncoder(encoderConfig), stderrSyncer, lowestLevel)
	stderrSystemLogger = zap.New(stderrCore, opts...).Sugar()

	gatewaySyncer := zapcore.AddSync(lf)
	gatewayCore := zapcore.NewCore(zapcore.NewConsoleEncoder(encoderConfig), gatewaySyncer, lowestLevel)
	diskSystemLogger = zap.New(gatewayCore, opts...).Sugar()

	defaultCore := zapcore.NewTee(gatewayCore, stderrCore)
	compoundSystemLogger = zap.New(defaultCore, opts...).Sugar()
}
