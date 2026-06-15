package cmd

import (
    "fmt"
    "log"
    "os"
    "path/filepath"
)

const (
    colorReset  = "\033[0m"
    colorGreen  = "\033[32m"
    colorYellow = "\033[33m"
    colorRed    = "\033[31m"
)

var Logger *log.Logger

func Info(format string, v ...any) {
    msg := fmt.Sprintf("[INFO] "+format, v...)
    logToFile(msg)
    fmt.Println(colorGreen + msg + colorReset)
}

func Warn(format string, v ...any) {
    msg := fmt.Sprintf("[WARN] "+format, v...)
    logToFile(msg)
    fmt.Println(colorYellow + msg + colorReset)
}

func Fatal(format string, v ...any) {
    msg := fmt.Sprintf("[FATAL] "+format, v...)
    logToFile(msg)
    fmt.Fprintln(os.Stderr, colorRed+msg+colorReset)
    os.Exit(1)
}

func Log() {
    exe, err := os.Executable()
    if err != nil {
        fmt.Println("[FATAL] Failed to resolve executable path:", err)
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

func logToFile(msg string) {
    if Logger != nil {
        Logger.Println(msg)
    }
}