package rpckit_log

import (
	"github.com/mattn/go-colorable"
	log "go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// setupLog creates logger
func New(debug bool) *log.Logger {

	var l *log.Logger

	aa := log.NewDevelopmentEncoderConfig()
	aa.EncodeLevel = zapcore.CapitalColorLevelEncoder
	l = log.New(zapcore.NewCore(
		zapcore.NewConsoleEncoder(aa),
		zapcore.AddSync(colorable.NewColorableStdout()),
		zapcore.DebugLevel,
	),
		log.AddCaller(),
	)

	//		l, _ = log.NewProduction()

	//	r.Use(ginzap.Ginzap(l, time.RFC3339, false))

	// Logs all panic to error log
	//   - stack means whether output the stack info.
	//	r.Use(ginzap.RecoveryWithZap(l, true))

	return l
}
