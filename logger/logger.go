package logger

import (
	"io"
	"os"
	"path/filepath"
	"time"

	log "github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
)

func SetupLogger() {

	lumberjackLogger := &lumberjack.Logger{
		Filename:   filepath.Join("logs", "travel-advisor.log"),
		MaxSize:    5, // MB
		MaxBackups: 10,
		MaxAge:     30,    // days
		Compress:   false, // disabled by default
	}

	// Fork writing into two outputs
	multiWriter := io.MultiWriter(os.Stderr, lumberjackLogger)

	logFormatter := new(log.TextFormatter)
	logFormatter.TimestampFormat = time.RFC1123Z
	logFormatter.FullTimestamp = true

	log.SetFormatter(logFormatter)
	if os.Getenv("LOGGER_LEVEL") == "" {
		log.SetLevel(log.InfoLevel)
	} else {
		level, err := log.ParseLevel(os.Getenv("LOGGER_LEVEL"))
		if err != nil {
			log.Fatalf("Invalid LOGGER_LEVEL: %v", err)
		}
		log.SetLevel(level)
	}

	log.SetLevel(log.DebugLevel)
	log.SetOutput(multiWriter)
}
