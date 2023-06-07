package main

import (
	"errors"
	C "factors/config"
	"flag"
	"time"

	log "github.com/sirupsen/logrus"
)

func main() {

	env := flag.String("env", C.DEVELOPMENT, "")

	sentryDSN := flag.String("sentry_dsn", "https://81f48ea1f7604e6eb98871c04f68f9d4@o435495.ingest.sentry.io/5394896", "Sentry DSN")
	overrideAppName := flag.String("app_name", "sentry_rollup_test", "Override default app_name.")
	useSentryRollup := flag.Bool("use_sentry_rollup", true, "Enables rollup support for sentry")
	sentryRollupSyncInSecs := flag.Int("sentry_rollup_sync_in_seconds", 10, "Enables to send errors to sentry in given interval in seconds.")

	defaultAppName := "sentry_rollup_job"
	appName := C.GetAppName(defaultAppName, *overrideAppName)
	flag.Parse()

	config := &C.Configuration{
		AppName:                appName,
		Env:                    *env,
		SentryDSN:              *sentryDSN,
		UseSentryRollup:        *useSentryRollup,
		SentryRollupSyncInSecs: *sentryRollupSyncInSecs,
	}
	C.InitConf(config)

	C.InitSentryLogging(*sentryDSN, appName)

	log.WithField("sentry_dsn", sentryDSN).Info("sentry error capturing started")
	for i := 0; i < 40; i++ {
		log.WithField("project_id", 1).WithError(errors.New("sample error")).Error("Test: sample errrors for sentry 123.")
		time.Sleep(1 * time.Second)
	}
	log.Info("sentry error capturing done")
}
