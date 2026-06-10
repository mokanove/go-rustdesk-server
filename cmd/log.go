package cmd

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
)
const (
	colorReset  = "\033[0m"
	colorYellow = "\033[33m"
	colorRed    = "\033[31m"
)

var Logger *log.Logger

func Log() {
	exe, err := os.Executable()
	if err != nil {
		fmt.Println("[FATAL] Failed to get executable path:", err)
		os.Exit(1)
	}
	logPath := filepath.Join(filepath.Dir(exe), "rdks.log")
	f, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println("[FATAL] Failed to open log file:", err)
		os.Exit(1)
	}
	Logger = log.New(f, "", log.LstdFlags)
}

func Info(format string, v ...any) {
	if Logger != nil {
		Logger.Printf("[INFO] "+format, v...)
	}
	fmt.Printf("[INFO] "+format+"\n", v...)
}

func Warn(format string, v ...any) {
	msg := fmt.Sprintf("[WARN] "+format, v...)
	if Logger != nil {
		Logger.Println(msg)
	}
	fmt.Println(colorYellow + msg + colorReset)
}

func Fatal(format string, v ...any) {
	msg := fmt.Sprintf("[FATAL] "+format, v...)
	if Logger != nil {
		Logger.Println(msg)
	}
	fmt.Println(colorRed + msg + colorReset)
	os.Exit(1)
}