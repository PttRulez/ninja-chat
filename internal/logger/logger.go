package logger

import (
	"errors"
	"fmt"
	stdlog "log"
	"os"
	"syscall"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	level   zap.AtomicLevel
	lvlsMap = make(map[int8]string)
)

func init() {
	level = zap.NewAtomicLevel()
	lvlsMap[-1] = "debug"
	lvlsMap[0] = "info"
	lvlsMap[1] = "warn"
	lvlsMap[2] = "error"
}

//go:generate options-gen -out-filename=logger_options.gen.go -from-struct=Options -defaults-from=var
type Options struct {
	level          string `option:"mandatory" validate:"required,oneof=debug info warn error"`
	productionMode bool
	clock          zapcore.Clock
}

var defaultOptions = Options{
	clock: zapcore.DefaultClock,
}

func MustInit(opts Options) {
	if err := Init(opts); err != nil {
		panic(err)
	}
}

func Init(opts Options) error {
	if err := opts.Validate(); err != nil {
		return fmt.Errorf("validate options: %v", err)
	}
	err := level.UnmarshalText([]byte(opts.level))
	if err != nil {
		return fmt.Errorf("invalid log level: %v", err)
	}

	encoderConfig := zapcore.EncoderConfig{
		NameKey:    "component",
		TimeKey:    "T",
		EncodeTime: zapcore.ISO8601TimeEncoder,
		MessageKey: "msg",
		LevelKey:   "level",
	}

	var encoder zapcore.Encoder
	if opts.productionMode {
		encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	} else {
		encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	}

	stdout := zapcore.AddSync(os.Stdout)
	cores := []zapcore.Core{
		zapcore.NewCore(encoder, stdout, level),
	}
	l := zap.New(zapcore.NewTee(cores...), zap.WithClock(opts.clock))
	zap.ReplaceGlobals(l)

	return nil
}

func Sync() {
	if err := zap.L().Sync(); err != nil && !errors.Is(err, syscall.ENOTTY) {
		stdlog.Printf("cannot sync logger: %v", err)
	}
}

func GetAtomicLevel() zap.AtomicLevel {
	return level
}
