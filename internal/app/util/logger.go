package util

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/SafetyCulture/iauditor-exporter/internal/app/version"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/utils"
)

var (
	infoStr      = "%s\n[info] "
	warnStr      = "%s\n[warn] "
	errStr       = "%s\n[error] "
	traceStr     = "%s\n[%.3fms] [rows:%v] %s"
	traceWarnStr = "%s %s\n[%.3fms] [rows:%v] %s"
	traceErrStr  = "%s %s\n[%.3fms] [rows:%v] %s"
	trace        = false
)

// GormLogger wraps logger with some gorm specific functionality
type GormLogger struct {
	*zap.SugaredLogger

	SlowThreshold time.Duration
}

// LogMode log mode.
func (l *GormLogger) LogMode(level logger.LogLevel) logger.Interface {
	newlogger := *l

	return &newlogger
}

// Info print info.
func (l GormLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	l.Infof(infoStr+msg, append([]interface{}{utils.FileWithLineNum()}, data...)...)
}

// Warn print warn messages.
func (l GormLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	l.Warnf(warnStr+msg, append([]interface{}{utils.FileWithLineNum()}, data...)...)
}

// Error print error messages.
func (l GormLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	l.Errorf(errStr+msg, append([]interface{}{utils.FileWithLineNum()}, data...)...)
}

// Trace print sql message.
func (l GormLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	elapsed := time.Since(begin)
	switch {
	case err != nil:
		sql, rows := fc()
		if rows == -1 {
			l.Errorf(traceErrStr, utils.FileWithLineNum(), err, float64(elapsed.Nanoseconds())/1e6, "-", sql)
		} else {
			l.Errorf(traceErrStr, utils.FileWithLineNum(), err, float64(elapsed.Nanoseconds())/1e6, rows, sql)
		}
	case elapsed > l.SlowThreshold && l.SlowThreshold != 0:
		sql, rows := fc()
		slowLog := fmt.Sprintf("SLOW SQL >= %v", l.SlowThreshold)
		if rows == -1 {
			l.Debugf(traceWarnStr, utils.FileWithLineNum(), slowLog, float64(elapsed.Nanoseconds())/1e6, "-", sql)
		} else {
			l.Debugf(traceWarnStr, utils.FileWithLineNum(), slowLog, float64(elapsed.Nanoseconds())/1e6, rows, sql)
		}
	default:
		if !trace {
			return
		}

		sql, rows := fc()

		if rows == -1 {
			l.Debugf(traceStr, utils.FileWithLineNum(), "-", sql)
		} else {
			l.Debugf(traceStr, utils.FileWithLineNum(), rows, sql)
		}
	}
}

// GetLogger returns a configured instance of the logger
func GetLogger() *zap.SugaredLogger {
	logFileEncoder := zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig())
	consoleEncoder := zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig())

	file, err := os.OpenFile("logs.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("unable to open log file %v", err)
	}

	logFileWriter := zapcore.Lock(file)
	consoleWriter := zapcore.Lock(os.Stderr)

	// Log to both console and the log file. This allows for succinct console logs and
	// Verbose detailed logs to review is something goes wrong.
	core := zapcore.NewTee(
		zapcore.NewCore(consoleEncoder, consoleWriter, zap.InfoLevel),
		zapcore.NewCore(logFileEncoder, logFileWriter, zap.DebugLevel),
	)

	// From a zapcore.Core, it's easy to construct a Logger.
	l := zap.New(core).Named(fmt.Sprintf("iauditor-exporter@%s", version.GetVersion()))
	defer l.Sync()

	// redirects output from the standard library's package-global logger to the supplied logger
	zap.RedirectStdLog(l)

	return l.Sugar()
}
