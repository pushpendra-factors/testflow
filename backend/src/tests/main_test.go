package tests

import (
	C "config"
	"os"
	"testing"

	log "github.com/sirupsen/logrus"
)

func TestMain(m *testing.M) {
	// Setup.
	// Initialize configs and connections.
	if err := C.Init(); err != nil {
		log.Fatal("Failed to initialize config and services.")
		os.Exit(1)
	}
	if C.GetConfig().Env != C.DEVELOPMENT {
		log.Fatal("Environment is not Development.")
		os.Exit(1)
	}
	retCode := m.Run()
	os.Exit(retCode)
}
