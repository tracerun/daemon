package lg

import (
	"os"

	"go.uber.org/zap"
)

var (
	// L the zap logger
	L *zap.Logger
)

type key int

const requestIDKey key = 0

// InitLogger must be first to be called.
func InitLogger(debug, nostd bool, logPath string) {
	var cfg zap.Config

	dyn := zap.NewAtomicLevel()
	if debug {
		dyn.SetLevel(zap.DebugLevel)
	}
	cfg.Level = dyn
	cfg.EncoderConfig.LevelKey = "lvl"
	cfg.EncoderConfig.MessageKey = "msg"
	cfg.EncoderConfig.TimeKey = "timestamp"
	cfg.Encoding = "console"
	cfg.EncoderConfig = zap.NewDevelopmentEncoderConfig()

	var paths []string
	if !nostd {
		paths = append(paths, "stderr")
	}
	if len(logPath) != 0 {
		os.OpenFile(logPath, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0666)
		paths = append(paths, logPath)
	}

	cfg.OutputPaths = paths
	cfg.ErrorOutputPaths = paths

	var err error
	if L, err = cfg.Build(); err != nil {
		panic(err)
	}
	L.Debug("log path", zap.Strings("paths", paths))
}
