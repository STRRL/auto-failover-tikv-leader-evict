package log

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
)

var logger *zap.Logger
var logLevel = zap.NewAtomicLevel()

func init() {
	logger = zap.New(
		zapcore.NewCore(
			zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
			zapcore.AddSync(os.Stdout),
			logLevel,
		),
	)
}

func L() *zap.Logger {
	return logger
}
func EnableDebug() {
	logLevel.SetLevel(zap.DebugLevel)
}
