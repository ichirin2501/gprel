package gprel

import (
	"os"

	log "github.com/sirupsen/logrus"
)

func init() {
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp:    true,
		DisableTimestamp: false,
		TimestampFormat:  "2006-01-02T15:04:05.000Z07:00",
	})
	log.SetOutput(os.Stdout)
	if d := os.Getenv("DEBUG"); d == "1" {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.InfoLevel)
	}
}
