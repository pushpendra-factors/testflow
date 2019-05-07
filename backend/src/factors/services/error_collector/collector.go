package error_collector

import (
	"encoding/json"
	"factors/interfaces/maileriface"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type stackTracer interface {
	StackTrace() errors.StackTrace
}

type Entry struct {
	e *logrus.Entry
}

type Entity struct {
	message []byte
}

type Collector struct {
	mailer                  maileriface.Mailer
	fromEmail, toEmail, env string
	entriesLock             sync.RWMutex
	entries                 []*Entry
	ticker                  *time.Ticker
}

func New(m maileriface.Mailer, reportingInterval time.Duration, env, toMail, fromMail string) *Collector {
	collector := Collector{
		mailer:    m,
		entries:   make([]*Entry, 0, 0),
		fromEmail: fromMail,
		toEmail:   toMail,
		env:       env,
	}

	go collector.reportAtIntervals(reportingInterval)

	return &collector
}

func (c *Collector) Notice(entry *logrus.Entry) {

	c.entriesLock.Lock()
	defer c.entriesLock.Unlock()

	entries := c.entries
	entries = append(entries, &Entry{
		e: entry,
	})
	c.entries = entries
}

func (c *Collector) reportAtIntervals(t time.Duration) {
	c.ticker = time.NewTicker(t)
	for {
		select {
		case <-c.ticker.C:
			c.Flush()
		}
	}
}

func (c *Collector) Flush() {

	c.entriesLock.Lock()
	defer c.entriesLock.Unlock()

	if len(c.entries) == 0 {
		return
	}

	var dataToSend strings.Builder
	stackStrace := ""
	for _, entry := range c.entries {
		reqId := entry.e.Data["reqId"]
		err := entry.e.Data[logrus.ErrorKey]

		if errWithStacktrace, ok := err.(stackTracer); ok {
			stackStrace = fmt.Sprintf("%+v", errWithStacktrace)
		}

		delete(entry.e.Data, logrus.ErrorKey)

		allEntries, _ := json.Marshal(entry.e.Data)
		// not logging error to avoid cycling hook calls
		dataToSend.WriteString(fmt.Sprintf("ReqId: %v\n, Error: %v\n, Stacktrace: %v\n, Data: %v\n\n", reqId, err, stackStrace, string(allEntries)))
	}

	str := dataToSend.String()

	if err := c.mailer.SendMail(c.toEmail, c.fromEmail, c.env+" Errors Noticed", str, str); err != nil {
	}

	emptyEntries := make([]*Entry, 0, 0)
	c.entries = emptyEntries
}

func (c *Collector) Stop() {
	c.Flush()
	c.ticker.Stop()
}
