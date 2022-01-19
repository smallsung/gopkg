package zapenc

import (
	"os"

	"github.com/smallsung/gopkg/errors"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Core struct {
	zapcore.Core
}

func (c Core) Build(options ...zap.Option) *zap.Logger {
	return New(c.Core, options...)
}

func NewCore(level zapcore.LevelEnabler, writer zapcore.WriteSyncer, encoder zapcore.Encoder) Core {
	if level == nil {
		panic(errors.New("nil pointer dereference"))
	}
	if writer == nil {
		panic(errors.New("nil pointer dereference"))
	}
	if encoder == nil {
		panic(errors.New("nil pointer dereference"))
	}
	return Core{zapcore.NewCore(encoder, WrapWriter(writer), level)}
}

func NewWriterCore(level zapcore.LevelEnabler, writer zapcore.WriteSyncer) Core {
	return NewCore(level, writer, zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()))
}

func NewFileCore(level zapcore.LevelEnabler, fileName string) Core {
	return NewWriterCore(level, Lumberjack(fileName))
}

var Stderr = WrapWriter(os.Stderr)

func newConsoleCore(level zapcore.LevelEnabler) Core {
	config := zap.NewDevelopmentEncoderConfig()
	config.EncodeLevel = zapcore.CapitalColorLevelEncoder
	return NewCore(level, Stderr, zapcore.NewConsoleEncoder(config))
}

func NewMultiCore(cores ...zapcore.Core) Core {
	return Core{zapcore.NewTee(cores...)}
}
