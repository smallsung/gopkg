package zapenc

import "go.uber.org/zap"

var (
	invokeLogger        *zap.Logger
	invokeSugaredLogger *zap.SugaredLogger
)

func init() {
	SetSugarLogger(newInvokeLogger())
}

func newInvokeLogger() *zap.Logger {
	return NewConsole(
		zap.DebugLevel,
		zap.AddCaller(),
		zap.Development(),
		zap.AddStacktrace(zap.WarnLevel),
		zap.AddCallerSkip(1),
	).Named("invoke.logger")
}

func SetSugarLogger(l *zap.Logger) {
	invokeLogger = l
	invokeSugaredLogger = l.Sugar()
}

func Debug(msg string, fields ...interface{}) {
	if newFields, ok := convertFields(fields...); ok {
		invokeLogger.Debug(msg, newFields...)
	} else {
		invokeSugaredLogger.Debug(convertInterfaces(msg, fields...))
	}
}

func Info(msg string, fields ...interface{}) {
	if newFields, ok := convertFields(fields...); ok {
		invokeLogger.Info(msg, newFields...)
	} else {
		invokeSugaredLogger.Info(convertInterfaces(msg, fields...))
	}
}
func Warn(msg string, fields ...interface{}) {
	if newFields, ok := convertFields(fields...); ok {
		invokeLogger.Warn(msg, newFields...)
	} else {
		invokeSugaredLogger.Warn(convertInterfaces(msg, fields...))
	}
}

func Error(msg string, fields ...interface{}) {
	if newFields, ok := convertFields(fields...); ok {
		invokeLogger.Error(msg, newFields...)
	} else {
		invokeSugaredLogger.Error(convertInterfaces(msg, fields...))
	}
}

func DPanic(msg string, fields ...interface{}) {
	if newFields, ok := convertFields(fields...); ok {
		invokeLogger.DPanic(msg, newFields...)
	} else {
		invokeSugaredLogger.DPanic(convertInterfaces(msg, fields...))
	}
}

func Panic(msg string, fields ...interface{}) {
	if newFields, ok := convertFields(fields...); ok {
		invokeLogger.Panic(msg, newFields...)
	} else {
		invokeSugaredLogger.Panic(convertInterfaces(msg, fields...))
	}
}

func Fatal(msg string, fields ...interface{}) {
	if newFields, ok := convertFields(fields...); ok {
		invokeLogger.Fatal(msg, newFields...)
	} else {
		invokeSugaredLogger.Fatal(convertInterfaces(msg, fields...))
	}
}

func convertFields(ctx ...interface{}) ([]zap.Field, bool) {
	var fields []zap.Field
	for _, x := range ctx {
		if xx, ok := x.(zap.Field); !ok {
			return nil, false
		} else {
			fields = append(fields, xx)
		}
	}
	return fields, true
}

func convertInterfaces(msg string, fields ...interface{}) []interface{} {
	cxt := make([]interface{}, 0, len(fields)+1)
	cxt = append(cxt, msg)
	cxt = append(cxt, fields...)
	return cxt
}
