package maingoroutine

import (
	"go.uber.org/zap"
)

type Config struct {
	Logger *zap.Logger

	EnableHTTP    bool
	HTTPHost      string
	HTTPPort      uint16
	EnableHTTPRPC bool
}
