package logger

import (
	rotatelogs "github.com/lestrrat/go-file-rotatelogs"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"io"
	"log"
	"os"
	"path"
	"strings"
	"time"
)

type LevelJudge func(zapcore.Level) bool

func (l LevelJudge) Enabled(level zapcore.Level) bool {
	return l(level)
}

var (
	DebugLevel LevelJudge = func(level zapcore.Level) bool {
		return level == zapcore.DebugLevel
	}
	InfoLevel LevelJudge = func(level zapcore.Level) bool {
		return level == zapcore.InfoLevel
	}

	WarnLevel LevelJudge = func(level zapcore.Level) bool {
		return level == zapcore.WarnLevel
	}

	FatalLevel LevelJudge = func(level zapcore.Level) bool {
		return level == zapcore.FatalLevel
	}

	ErrorLevel LevelJudge = func(level zapcore.Level) bool {
		return level == zapcore.ErrorLevel
	}
)

func init() {
	config := zapcore.EncoderConfig{}
	config.MessageKey = "Msg"
	config.LevelKey = "Level"
	config.TimeKey = "Time"
	config.NameKey = "Name"
	config.CallerKey = "Caller"
	config.FunctionKey = "Function"
	config.StacktraceKey = "StackTrace"
	config.EncodeLevel = zapcore.CapitalLevelEncoder
	config.EncodeTime = zapcore.TimeEncoderOfLayout(time.ANSIC)
	config.EncodeDuration = zapcore.SecondsDurationEncoder
	config.EncodeCaller = zapcore.ShortCallerEncoder

	consoleEncoder := zapcore.NewConsoleEncoder(config)

	var err error

	debugWriter, err := getWriter("Debug.log")
	if err != nil {
		log.Fatalf("init logger fatal,err is %v", err)
	}
	debugCore := zapcore.NewCore(consoleEncoder, zapcore.AddSync(debugWriter), DebugLevel)
	debugStdCore := zapcore.NewCore(consoleEncoder, os.Stdout, zapcore.DebugLevel)

	infoWriter, err := getWriter("Info.log")
	if err != nil {
		log.Fatalf("init logger fatal,err is %v", err)
	}
	infoCore := zapcore.NewCore(consoleEncoder, zapcore.AddSync(infoWriter), InfoLevel)

	warnWriter, err := getWriter("Warn.log")
	if err != nil {
		log.Fatalf("init logger fatal,err is %v", err)
	}
	warnCore := zapcore.NewCore(consoleEncoder, zapcore.AddSync(warnWriter), WarnLevel)

	errorWriter, err := getWriter("Error.log")
	if err != nil {
		log.Fatalf("init logger fatal,err is %v", err)
	}
	errorCore := zapcore.NewCore(consoleEncoder, zapcore.AddSync(errorWriter), ErrorLevel)

	fatalWriter, err := getWriter("Fatal.log")
	if err != nil {
		log.Fatalf("init logger fatal,err is %v", err)
	}
	fatalCore := zapcore.NewCore(consoleEncoder, zapcore.AddSync(fatalWriter), FatalLevel)

	multiCore := zapcore.NewTee(debugCore, infoCore, warnCore, errorCore, fatalCore, debugStdCore)

	logger := zap.New(multiCore, zap.AddStacktrace(zapcore.ErrorLevel), zap.AddCaller(), zap.OnFatal(zapcore.WriteThenPanic))
	logger = logger.Named("Logger")
	zap.ReplaceGlobals(logger)
}

func getWriter(fileName string) (io.Writer, error) {
	hook, err := rotatelogs.New(
		path.Join("/home/lqf/golang/GoProject/cache-go/log", strings.ReplaceAll(fileName, ".log", "")+"-%Y-%m-%d.log"),
		rotatelogs.WithRotationTime(time.Hour),
		rotatelogs.WithMaxAge(time.Hour*24*30),
	)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	return hook, nil
}
