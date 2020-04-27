package debug

import (
	"github.com/mattn/go-colorable"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// setupLog creates logger
func NewLogger(debug bool) *zap.Logger {

	var l *zap.Logger

	aa := zap.NewDevelopmentEncoderConfig()
	aa.EncodeLevel = zapcore.CapitalColorLevelEncoder
	l = zap.New(zapcore.NewCore(
		zapcore.NewConsoleEncoder(aa),
		zapcore.AddSync(colorable.NewColorableStdout()),
		zapcore.DebugLevel,
	),
		zap.AddCaller(),
	)

	//		l, _ = zap.NewProduction()

	//	r.Use(ginzap.Ginzap(l, time.RFC3339, false))

	// Logs all panic to error log
	//   - stack means whether output the stack info.
	//	r.Use(ginzap.RecoveryWithZap(l, true))

	return l
}
