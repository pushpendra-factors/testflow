package beam

import (
	"bytes"
	"net"
	"os"
	"runtime"
	"strconv"

	log "github.com/sirupsen/logrus"
)

func GetGoroutineID() uint64 {
	b := make([]byte, 64)
	b = b[:runtime.Stack(b, false)]
	b = bytes.TrimPrefix(b, []byte("goroutine "))
	b = b[:bytes.IndexByte(b, ' ')]
	n, _ := strconv.ParseUint(string(b), 10, 64)
	return n
}

func GetMacAddr() []string {
	ifas, err := net.Interfaces()
	if err != nil {
		return nil
	}
	var as []string
	for _, ifa := range ifas {
		a := ifa.HardwareAddr.String()
		if a != "" {
			as = append(as, a)
		}
	}
	return as
}

func GetLogContext() *log.Entry {
	return log.WithFields(log.Fields{
		"Name":    os.Args[0],
		"PID":     os.Getpid(),
		"PPID":    os.Getppid(),
		"Routine": GetGoroutineID(),
		"MACAddr": GetMacAddr(),
	})
}
