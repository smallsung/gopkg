package zapenc

import (
	"io"

	"github.com/smallsung/gopkg/errors"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

func WrapWriter(w io.Writer) zapcore.WriteSyncer {
	return zapcore.Lock(
		zapcore.AddSync(w),
	)
}

func Lumberjack(fileName string) zapcore.WriteSyncer {
	return WrapWriter(&lumberjack.Logger{
		Filename:   fileName,
		MaxSize:    128,   // 每个日志文件保存的大小 单位:M
		MaxAge:     180,   // 文件最多保存多少天
		MaxBackups: 180,   // 日志文件最多保存多少个备份
		LocalTime:  false, //
		Compress:   false, // 是否压缩
	})
}

func MultiWriter(mw ...interface{}) zapcore.WriteSyncer {
	var ws []zapcore.WriteSyncer
	for _, x := range mw {
		var w io.Writer

		switch x.(type) {
		case zapcore.WriteSyncer:
			w, _ = x.(zapcore.WriteSyncer)
		case io.Writer:
			w, _ = x.(io.Writer)
		case string:
			w = Lumberjack(x.(string))
		default:
			panic(errors.New("NotImplementErr"))
		}

		ws = append(ws, WrapWriter(w))
	}

	return zapcore.NewMultiWriteSyncer(ws...)
}
