package log

import (
	"os"

	"github.com/fatih/color"
)

func Info(msg string) {
	color.Cyan(msg)
}

func Error(msg string) {
	color.Red(msg)
}

func Success(msg string) {
	color.Green(msg)
}

func Warn(msg string) {
	color.Yellow(msg)
}

func Fatal(msg string) {
	color.Red(msg)
	os.Exit(0)
}
