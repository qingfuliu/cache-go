package logger

import (
	"fmt"
	"go.uber.org/zap"
	"os"
	"testing"
)

func TestLogger(t *testing.T) {
	fmt.Println(os.Getwd())
	zap.L().Debug("err is:", zap.String("hhh", "lqf"))
	zap.L().Info("err is:", zap.String("hhh", "lqf"))
	zap.L().Warn("err is:", zap.String("hhh", "lqf"))
	zap.L().Error("err is:", zap.String("hhh", "lqf"))
	zap.L().Fatal("err is:", zap.String("hhh", "lqf"))
}
