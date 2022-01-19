package zapenc

import (
	"testing"

	"go.uber.org/zap"
)

func TestInvokeLogger(t *testing.T) {
	Debug("Debug fields", zap.String("field key", "Debug"))
	Debug("Debug fields", "field key", "Debug")

	Info("Info fields", zap.String("field key", "Info"))
	Info("Info fields", "field key", "Info")

	Warn("Warn fields", zap.String("field key", "Warn"))
	Warn("Warn fields", "field key", "Warn")

	Error("Error fields", zap.String("field key", "Error"))
	Error("Error fields", "field key", "Error")

	DPanic("DPanic fields", zap.String("field key", "DPanic"))
	DPanic("DPanic fields", "field key", "DPanic")

	Panic("Panic fields", zap.String("field key", "Panic"))
	Panic("Panic fields", "field key", "Panic")
}
