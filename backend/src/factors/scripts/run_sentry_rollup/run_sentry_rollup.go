package main

import (
	C "factors/config"
	"flag"
	"time"

	log "github.com/sirupsen/logrus"
)

func main() {

	env := flag.String("env", C.DEVELOPMENT, "")

	sentryDSN := flag.String("sentry_dsn", "", "Sentry DSN")
	overrideAppName := flag.String("app_name", "", "Override default app_name.")
	useSentryRollup := flag.Bool("use_sentry_rollup", false, "Enables rollup support for sentry")
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

	log.WithField("sentry_dsn", sentryDSN).Error("sentry error capturing started")
	for i := 0; i < 30; i++ {
		log.Error("Test: sample errrors for sentry")
		time.Sleep(1 * time.Second)
	}
	log.Error("sentry error capturing done")

}
