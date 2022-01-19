package zapenc

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func New(core zapcore.Core, options ...zap.Option) *zap.Logger {
	return zap.New(core, options...)
}

func NewWriter(level zapcore.LevelEnabler, writer zapcore.WriteSyncer, options ...zap.Option) *zap.Logger {
	return NewWriterCore(level, writer).Build(options...)
}

func NewRotateFile(level zapcore.LevelEnabler, fileName string, options ...zap.Option) *zap.Logger {
	return NewFileCore(level, fileName).Build(options...)
}

func NewFile(level zapcore.LevelEnabler, fileName string, options ...zap.Option) *zap.Logger {
	return NewRotateFile(level, fileName, options...)
}

func NewConsole(level zapcore.LevelEnabler, options ...zap.Option) *zap.Logger {
	options = append([]zap.Option{zap.AddCaller(), zap.Development(), zap.AddStacktrace(zap.ErrorLevel)}, options...)
	return newConsoleCore(level).Build(options...)
}

func NewWriterWithConsole(level zapcore.LevelEnabler, writer zapcore.WriteSyncer, options ...zap.Option) *zap.Logger {
	cores := []zapcore.Core{NewWriterCore(level, writer)}
	cores = append(cores, newConsoleCore(level))
	return NewMultiCore(cores...).Build(options...)
}

func NewFileWithConsole(level zapcore.LevelEnabler, fileName string, options ...zap.Option) *zap.Logger {
	cores := []zapcore.Core{NewFileCore(level, fileName)}
	cores = append(cores, newConsoleCore(level))
	return NewMultiCore(cores...).Build(options...)
}

func NewWithConsole(level zapcore.LevelEnabler, core zapcore.Core, options ...zap.Option) *zap.Logger {
	cores := []zapcore.Core{core}
	cores = append(cores, newConsoleCore(level))
	return NewMultiCore(cores...).Build(options...)
}

func NewDefaultLogger() *zap.Logger {
	return NewWithConsole(
		zap.DebugLevel,

		NewMultiCore(
			NewFileCore(zap.DebugLevel, "default.log"),
			NewFileCore(zap.ErrorLevel, "error.log"),
		),

		zap.AddCaller(),
		zap.Development(),
		zap.AddStacktrace(zap.ErrorLevel),
	)
}
