package main

import (
	"context"
	"fmt"
	"os"
	"runtime/debug"

	"github.com/coze-dev/coze-studio/backend/pkg/lang/ternary"
	"github.com/coze-dev/coze-studio/backend/pkg/logs"
	"github.com/joho/godotenv"
)

func main() {
	ctx := context.Background()
	// Please do not change the order of the function calls below
	setCrashOutput()

}

func loadEnv() (err error) {
	appEnv := os.Getenv("APP_ENV")
	fileName := ternary.IFElse(appEnv == "", ".env", ".env."+appEnv)

	logs.Infof("load env file: %s", fileName)

	err = godotenv.Load(fileName)
	if err != nil {
		return fmt.Errorf("load env file(%s) failed, err=%w", fileName, err)
	}

	return err
}

func getEnv(key string, defaultValue string) string {
	v := os.Getenv(key)
	if v == "" {
		return defaultValue
	}
	return v
}

func setCrashOutput() {
	crashFile, _ := os.Create("crash.log")
	_ = debug.SetCrashOutput(crashFile, debug.CrashOptions{})
}
