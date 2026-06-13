package main

import (
	"log"
	"math/rand/v2"
	"os"
	"time"
)

const (
	colorGray      = "\033[90m"
	colorWhite     = "\033[37m"
	colorYellow    = "\033[33m"
	colorRed       = "\033[31m"
	colorBrightRed = "\033[91m"
	colorReset     = "\033[0m"
)

var LogLevel = 3 // 0=FATAL, 1=ERROR, 2=WARN, 3=INFO, 4=VERBOSE

const easter string = "Man, this system is so\n\033[31m" +
	"   ▄████████    ▄████████    ▄████████\n" +
	"  ███    ███   ███    ███   ███    ███\n" +
	"  ███    ███   ███    █▀    ███    █▀ \n" +
	"  ███    ███   ███          ███       \n" +
	"▀███████████ ▀███████████ ▀███████████\n" +
	"  ███    ███          ███          ███\n" +
	"  ███    ███    ▄█    ███    ▄█    ███\n" +
	"  ███    █▀   ▄████████▀   ▄████████▀ \n" +
	"\033[0mthat the tg-ws-proxy-go terminated."

var std = log.New(os.Stderr, "", 0)

func logf(level, color string, format string, v ...interface{}) {
	timestamp := time.Now().Format("02 Jan 2006 15:04:05.000")
	msg := colorGray + "[" + timestamp + "]" + color + " [" + level + "] " + format + colorReset
	std.Printf(msg, v...)
}

func Verbose(format string, v ...interface{}) {
	if LogLevel >= 4 {
		logf("VERBOSE", colorGray, format, v...)
	}
}

func Info(format string, v ...interface{}) {
	if LogLevel >= 3 {
		logf("INFO", colorWhite, format, v...)
	}
}

func Warn(format string, v ...interface{}) {
	if LogLevel >= 2 {
		logf("WARN", colorYellow, format, v...)
	}
}

func Error(format string, v ...interface{}) {
	if LogLevel >= 1 {
		logf("ERROR", colorBrightRed, format, v...)
	}
}

func Fatal(format string, v ...interface{}) {
	logf("FATAL", colorRed, format, v...)
	if rand.IntN(2) == 1 {
		logf("Golang", colorWhite, easter)
	}
	os.Exit(1)
}

func PrintLogo() {
	Info(colorYellow + "            ━┏┛┏━┛      ┃┃┃┏━┛      ┏━┃┏━┃┏━┃┃ ┃┃ ┃      ┏━┛┏━┃            " + colorReset)
	Info(colorYellow + "             ┃ ┃ ┃  ━┛  ┃┃┃━━┃  ━┛  ┏━┛┏┏┛┃ ┃ ┛ ━┏┛  ━┛  ┃ ┃┃ ┃            " + colorReset)
	Info(colorYellow + "             ┛ ━━┛      ━━┛━━┛      ┛  ┛ ┛━━┛┛ ┛ ┛       ━━┛━━┛            " + colorReset)
}
