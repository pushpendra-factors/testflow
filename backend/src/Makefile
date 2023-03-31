.PHONY: build-api serve-api build-ps serve-ps build-ds serve-ds build-sdk serve-sdk serve-redis serve-predis clean pack-api upload-api pack-ps upload-ps pack-ds upload-ds pack-sdk upload-sdk pack-redis upload-redis pack-predis upload-predis pack-debug upload-debug pack-build-seq upload-build-seq pack-hubspot-enrich upload-hubspot-enrich pack-sdk-request-worker upload-sdk-request-worker pack-integration-request-worker upload-integration-request-worker pack-monitoring upload-monitoring pack-archive-events upload-archive-events pack-adhoc-archive-events upload-adhoc-archive-events pack-bigquery-upload upload-bigquery-upload pack-onboard-to-bigquery upload-onboard-to-bigquery pack-add-session upload-add-session pack-dashboard-caching upload-dashboard-caching pack-monthly-dashboard-caching upload-monthly-dashboard-caching pack-beam-dashboard-caching upload-beam-dashboard-caching pack-beam-dashboard-caching-now upload-beam-dashboard-caching-now pack-web-analytics-dashboard upload-web-analytics-dashboard pack-instantiate-event-user-cache upload-instantiate-event-user-cache pack-journey-mining upload-journey-mining build-api-doc pack-pull-events upload-pull-events go-fmt pack-beam-add-session upload-beam-add-session pack-memsql-hubspot-sync-fields upload-memsql-hubspot-sync-fields pack-replicate-properties upload-replicate-properties pack-precompiles-queries upload-precompile-queries pack-sdk-bundle upload-sdk-bundle pack-event-bundle upload-event-bundle pack-analytics-bundle upload-analytics-bundle test pack-convert_epoch_saved_queries_tz1_to_tz2 upload-convert_epoch_saved_queries_tz1_to_tz2 pack-channels-to-kpi-migration upload-channels-to-kpi-migration pack-predict-pull-events upload-predict-pull-events pack-form-fills upload-form-fills pack-delete-older-clickable-elements upload-delete-older-clickable-elements pack-dashboard-caching-queries-comparision upload-dashboard-caching-queries-comparision

# Update tag with the latest release version
# Ex format: v0.05-beta_d0e9ed9
# git rev-list HEAD --max-count=1 --abbrev-commit
TAG?=
export TAG

# staging | production
ENV?=staging
export ENV

# Should be set in ~/.bashrc or ~/.profile or passed every time.
export CHANNEL_TOKEN

build-api:
	rm $(GOPATH)/bin/app; go build -o $(GOPATH)/bin/app $(GOPATH)/src/factors/app/app.go

serve-api:
	$(GOPATH)/bin/app

serve-api-memsql:
	$(GOPATH)/bin/app --primary_datastore=memsql

build-ps:
	go build -o $(GOPATH)/bin/pattern-app $(GOPATH)/src/factors/pattern_server/cmd/pattern-app.go

serve-ps:
	$(GOPATH)/bin/pattern-app

build-ds:
	rm $(GOPATH)/bin/data-service; go build -o $(GOPATH)/bin/data-service $(GOPATH)/src/factors/data_service/data-service.go

serve-ds:
	$(GOPATH)/bin/data-service

build-sdk:
	rm $(GOPATH)/bin/sdk-service; go build -o $(GOPATH)/bin/sdk-service $(GOPATH)/src/factors/sdk/service/service.go

serve-sdk:
	$(GOPATH)/bin/sdk-service

serve-redis:
	redis-server $(GOPATH)/src/redis.conf

serve-predis:
	redis-server $(GOPATH)/src/persistent_redis.conf

serve-qredis:
	redis-server $(GOPATH)/src/queue_redis.conf

clean:
	rm $(GOPATH)/bin/pattern-app $(GOPATH)/bin/app $(GOPATH)/bin/data-service

pack-api:
	docker build -t us.gcr.io/factors-$(ENV)/app-server:$(TAG) -f Dockerfile.appserver .

upload-api: export IMAGE_NAME=app-server
upload-api: notify-deployment
	docker push us.gcr.io/factors-$(ENV)/app-server:$(TAG)

pack-ps:
	docker build -t us.gcr.io/factors-$(ENV)/pattern-server:$(TAG) -f Dockerfile.patternserver .

upload-ps: export IMAGE_NAME=pattern-server
upload-ps: notify-deployment
	docker push us.gcr.io/factors-$(ENV)/pattern-server:$(TAG)

pack-ds:
	docker build -t us.gcr.io/factors-$(ENV)/data-server:$(TAG) -f Dockerfile.dataserver .

upload-ds: export IMAGE_NAME=data-server
upload-ds: notify-deployment
	docker push us.gcr.io/factors-$(ENV)/data-server:$(TAG)

pack-sdk:
	docker build -t us.gcr.io/factors-$(ENV)/sdk-server:$(TAG) -f Dockerfile.sdkserver .

upload-sdk: export IMAGE_NAME=sdk-server
upload-sdk: notify-deployment
	docker push us.gcr.io/factors-$(ENV)/sdk-server:$(TAG)

pack-redis:
	docker build -t us.gcr.io/factors-$(ENV)/redis:$(TAG) -f Dockerfile.redis .

upload-redis:
	docker push us.gcr.io/factors-$(ENV)/redis:$(TAG)

pack-predis:
	docker build -t us.gcr.io/factors-$(ENV)/persistent-redis:$(TAG) -f Dockerfile.persistent_redis .

upload-predis:
	docker push us.gcr.io/factors-$(ENV)/persistent-redis:$(TAG)

pack-qredis:
	docker build -t us.gcr.io/factors-$(ENV)/queue-redis:$(TAG) -f Dockerfile.queue_redis .

upload-qredis:
	docker push us.gcr.io/factors-$(ENV)/queue-redis:$(TAG)

pack-debug:
	docker build --no-cache -t us.gcr.io/factors-$(ENV)/debug:$(TAG) -f Dockerfile.debug .

upload-debug: export IMAGE_NAME=debug
upload-debug: notify-deployment
	docker push us.gcr.io/factors-$(ENV)/debug:$(TAG)

pack-build-seq:
	docker build -t us.gcr.io/factors-$(ENV)/build-seq-job:$(TAG) -f Dockerfile.build_seq_job .

upload-build-seq: export IMAGE_NAME=build-seq-job
upload-build-seq: notify-deployment
	docker push us.gcr.io/factors-$(ENV)/build-seq-job:$(TAG)

pack-pull-events:
	docker build -t us.gcr.io/factors-$(ENV)/pull-events-job:$(TAG) -f Dockerfile.pull_events_job .

upload-pull-events: export IMAGE_NAME=pull-events-job
upload-pull-events: notify-deployment
	docker push us.gcr.io/factors-$(ENV)/pull-events-job:$(TAG)

pack-project-events:
	docker build -t us.gcr.io/factors-$(ENV)/project-events-job:$(TAG) -f Dockerfile.project_events_job .

upload-project-events: export IMAGE_NAME=project-events-job
upload-project-events: notify-deployment
	docker push us.gcr.io/factors-$(ENV)/project-events-job:$(TAG)

pack-pattern-mine:
	docker build -t us.gcr.io/factors-$(ENV)/pattern-mine-job:$(TAG) -f Dockerfile.pattern_mine_job .

upload-pattern-mine: export IMAGE_NAME=pattern-mine-job
upload-pattern-mine: notify-deployment
	docker push us.gcr.io/factors-$(ENV)/pattern-mine-job:$(TAG)

pack-explain:
	docker build -t us.gcr.io/factors-$(ENV)/explain-job:$(TAG) -f Dockerfile.explain_v2 .

upload-explain: export IMAGE_NAME=explain-job
upload-explain: notify-deployment
	docker push us.gcr.io/factors-$(ENV)/explain-job:$(TAG)

pack-add-session:
	docker build -t us.gcr.io/factors-$(ENV)/add-session-job:$(TAG) -f Dockerfile.add_session_job .

upload-add-session: export IMAGE_NAME=add-session-job
upload-add-session: notify-deployment
	docker push us.gcr.io/factors-$(ENV)/add-session-job:$(TAG)

pack-import-ads:
	docker build -t us.gcr.io/factors-$(ENV)/import-ads-job:$(TAG) -f Dockerfile.import_ads_job .

upload-import-ads: export IMAGE_NAME=import-ads-job
upload-import-ads: notify-deployment
	docker push us.gcr.io/factors-$(ENV)/import-ads-job:$(TAG)

pack-delete-dangling-sessions:
	docker build -t us.gcr.io/factors-$(ENV)/delete-dangling-sessions-job:$(TAG) -f Dockerfile.delete_dangling_sessions_job .

upload-delete-dangling-sessions: export IMAGE_NAME=delete-dangling-sessions-job
upload-delete-dangling-sessions: notify-deployment
	docker push us.gcr.io/factors-$(ENV)/delete-dangling-sessions-job:$(TAG)

pack-hubspot-enrich:
	docker build -t us.gcr.io/factors-$(ENV)/hubspot-enrich-job:$(TAG) -f Dockerfile.hubspot_enrich_job .

upload-hubspot-enrich: export IMAGE_NAME=hubspot-enrich-job
upload-hubspot-enrich: notify-deployment
	docker push us.gcr.io/factors-$(ENV)/hubspot-enrich-job:$(TAG)

pack-otp-hubspot:
	docker build -t us.gcr.io/factors-$(ENV)/otp-hubspot-job:$(TAG) -f Dockerfile.otp_hubspot_job .

upload-otp-hubspot: export IMAGE_NAME=otp-hubspot-job
upload-otp-hubspot: notify-deployment
	docker push us.gcr.io/factors-$(ENV)/otp-hubspot-job:$(TAG)


pack-marketo-enrich:
	docker build -t us.gcr.io/factors-$(ENV)/marketo-enrich-job:$(TAG) -f Dockerfile.marketo_enrich_job .

upload-marketo-enrich: export IMAGE_NAME=marketo-enrich-job
upload-marketo-enrich: notify-deployment
	docker push us.gcr.io/factors-$(ENV)/marketo-enrich-job:$(TAG)

pack-marketo-sync:
	docker build -t us.gcr.io/factors-$(ENV)/marketo-sync-job:$(TAG) -f Dockerfile.marketo_sync_job .

upload-marketo-sync: export IMAGE_NAME=marketo-sync-job
upload-marketo-sync: notify-deployment
	docker push us.gcr.io/factors-$(ENV)/marketo-sync-job:$(TAG)

pack-leadsquared-sync:
	docker build -t us.gcr.io/factors-$(ENV)/leadsquared-sync-job:$(TAG) -f Dockerfile.leadsquared_sync_job .

upload-leadsquared-sync: export IMAGE_NAME=leadsquared-sync-job
upload-leadsquared-sync: notify-deployment
	docker push us.gcr.io/factors-$(ENV)/leadsquared-sync-job:$(TAG)

pack-leadsquared-pull:
	docker build -t us.gcr.io/factors-$(ENV)/leadsquared-pull-job:$(TAG) -f Dockerfile.leadsquared_pull_job .

upload-leadsquared-pull: export IMAGE_NAME=leadsquared-pull-job
upload-leadsquared-pull: notify-deployment
	docker push us.gcr.io/factors-$(ENV)/leadsquared-pull-job:$(TAG)

pack-leadsquared-enrich:
	docker build -t us.gcr.io/factors-$(ENV)/leadsquared-enrich-job:$(TAG) -f Dockerfile.leadsquared_enrich_job .

upload-leadsquared-enrich: export IMAGE_NAME=leadsquared-enrich-job
upload-leadsquared-enrich: notify-deployment
	docker push us.gcr.io/factors-$(ENV)/leadsquared-enrich-job:$(TAG)

pack-enrich-smart-properties:
	docker build -t us.gcr.io/factors-$(ENV)/enrich-smart-properties-job:$(TAG) -f Dockerfile.enrich_smart_properties_job .

upload-enrich-smart-properties: export IMAGE_NAME=enrich-smart-properties-job
upload-enrich-smart-properties: notify-deployment
	docker push us.gcr.io/factors-$(ENV)/enrich-smart-properties-job:$(TAG)

pack-ingest-leadgen:
	docker build -t us.gcr.io/factors-$(ENV)/ingest-leadgen-job:$(TAG) -f Dockerfile.ingest_leadgen_job .

upload-ingest-leadgen: export IMAGE_NAME=ingest-leadgen-job
upload-ingest-leadgen: notify-deployment
	docker push us.gcr.io/factors-$(ENV)/ingest-leadgen-job:$(TAG)

pack-salesforce-enrich:
	docker build -t us.gcr.io/factors-$(ENV)/salesforce-enrich-job:$(TAG) -f Dockerfile.salesforce_enrich_job .

upload-salesforce-enrich: export IMAGE_NAME=salesforce-enrich-job
upload-salesforce-enrich: notify-deployment
	docker push us.gcr.io/factors-$(ENV)/salesforce-enrich-job:$(TAG)

pack-otp-salesforce:
	docker build -t us.gcr.io/factors-$(ENV)/otp-salesforce-job:$(TAG) -f Dockerfile.otp_salesforce_job .

upload-otp-salesforce: export IMAGE_NAME=otp-salesforce-job
upload-otp-salesforce: notify-deployment
	docker push us.gcr.io/factors-$(ENV)/otp-salesforce-job:$(TAG)

pack-backfill-salesforce-datetime-job:
	docker build -t us.gcr.io/factors-$(ENV)/backfill-salesforce-datetime-job:$(TAG) -f Dockerfile.backfill_salesforce_datetime_job .

upload-backfill-salesforce-datetime-job: export IMAGE_NAME=backfill-salesforce-datetime-job
upload-backfill-salesforce-datetime-job: notify-deployment
	docker push us.gcr.io/factors-$(ENV)/backfill-salesforce-datetime-job:$(TAG)

pack-backfill-salesforce-smart-event-job:
	docker build -t us.gcr.io/factors-$(ENV)/backfill-salesforce-smart-event-job:$(TAG) -f Dockerfile.backfill_salesforce_datetime_job .

upload-backfill-salesforce-smart-event-job: export IMAGE_NAME=backfill-salesforce-smart-event-job
upload-backfill-salesforce-smart-event-job: notify-deployment
	docker push us.gcr.io/factors-$(ENV)/backfill-salesforce-datetime-job:$(TAG)

pack-fix-salesforce-identify-campaign-with-contact-and-lead-association:
	docker build -t us.gcr.io/factors-$(ENV)/fix-salesforce-identify-campaign-with-contact-and-lead-association:$(TAG) -f Dockerfile.fix_salesforce_campaign_contact_and_lead_association_job .

upload-fix-salesforce-identify-campaign-with-contact-and-lead-association: export IMAGE_NAME=fix-salesforce-identify-campaign-with-contact-and-lead-association
upload-fix-salesforce-identify-campaign-with-contact-and-lead: notify-deployment
	docker push us.gcr.io/factors-$(ENV)/fix-salesforce-identify-campaign-with-contact-and-lead-association:$(TAG)

pack-remove-empty-smart-events:
	docker build -t us.gcr.io/factors-$(ENV)/remove-empty-smart-events:$(TAG) -f Dockerfile.remove_empty_smart_events_job .

upload-remove-empty-smart-events: export IMAGE_NAME=remove-empty-smart-events
upload-remove-empty-smart-events: notify-deployment
	docker push us.gcr.io/factors-$(ENV)/remove-empty-smart-events:$(TAG)

pack-hubspot-smart-event-validator-job:
	docker build -t us.gcr.io/factors-$(ENV)/hubspot-smart-event-validator-job:$(TAG) -f Dockerfile.hubspot_smart_event_validator_job .

upload-hubspot-smart-event-validator-job: export IMAGE_NAME=hubspot-smart-event-validator-job
upload-hubspot-smart-event-validator-job: notify-deployment
	docker push us.gcr.io/factors-$(ENV)/hubspot-smart-event-validator-job:$(TAG)

pack-sdk-request-worker:
	docker build -t us.gcr.io/factors-$(ENV)/sdk-request-worker:$(TAG) -f Dockerfile.sdk_request_worker .

upload-sdk-request-worker: export IMAGE_NAME=sdk-request-worker
upload-sdk-request-worker: notify-deployment
	docker push us.gcr.io/factors-$(ENV)/sdk-request-worker:$(TAG)

pack-integration-request-worker:
	docker build -t us.gcr.io/factors-$(ENV)/integration-request-worker:$(TAG) -f Dockerfile.integration_request_worker .

upload-integration-request-worker: export IMAGE_NAME=integration-request-worker
upload-integration-request-worker: notify-deployment
	docker push us.gcr.io/factors-$(ENV)/integration-request-worker:$(TAG)

pack-monitoring:
	docker build -t us.gcr.io/factors-$(ENV)/monitoring:$(TAG) -f Dockerfile.monitoring .

upload-monitoring: export IMAGE_NAME=monitoring
upload-monitoring: notify-deployment
	docker push us.gcr.io/factors-$(ENV)/monitoring:$(TAG)

pack-analyze:
	docker build -t us.gcr.io/factors-$(ENV)/analyze:$(TAG) -f Dockerfile.analyze_job .

upload-analyze: export IMAGE_NAME=analyze
upload-analyze: notify-deployment
	docker push us.gcr.io/factors-$(ENV)/analyze:$(TAG)

pack-archive-events:
	docker build -t us.gcr.io/factors-$(ENV)/archive-events-job:$(TAG) -f Dockerfile.archive_events_job .

upload-archive-events: export IMAGE_NAME=archive-events-job
upload-archive-events: notify-deployment
	docker push us.gcr.io/factors-$(ENV)/archive-events-job:$(TAG)

pack-adhoc-archive-events:
	docker build -t us.gcr.io/factors-$(ENV)/adhoc-archive-events-job:$(TAG) -f Dockerfile.adhoc_archive_events_job .

upload-adhoc-archive-events: export IMAGE_NAME=adhoc-archive-events-job
upload-adhoc-archive-events: notify-deployment
	docker push us.gcr.io/factors-$(ENV)/adhoc-archive-events-job:$(TAG)

pack-bigquery-upload:
	docker build -t us.gcr.io/factors-$(ENV)/bigquery-upload-job:$(TAG) -f Dockerfile.bigquery_upload_job .

upload-bigquery-upload: export IMAGE_NAME=bigquery-upload-job
upload-bigquery-upload: notify-deployment
	docker push us.gcr.io/factors-$(ENV)/bigquery-upload-job:$(TAG)

pack-onboard-to-bigquery:
	docker build -t us.gcr.io/factors-$(ENV)/onboard-to-bigquery-job:$(TAG) -f Dockerfile.onboard_to_bigquery_job .

upload-onboard-to-bigquery: export IMAGE_NAME=onboard-to-bigquery-job
upload-onboard-to-bigquery: notify-deployment
	docker push us.gcr.io/factors-$(ENV)/onboard-to-bigquery-job:$(TAG)

pack-merge-user-properties:
	docker build -t us.gcr.io/factors-$(ENV)/merge-user-properties-job:$(TAG) -f Dockerfile.merge_user_properties_job .

upload-merge-user-properties: export IMAGE_NAME=merge-user-properties-job
upload-merge-user-properties: notify-deployment
	docker push us.gcr.io/factors-$(ENV)/merge-user-properties-job:$(TAG)

pack-dashboard-caching:
	docker build -t us.gcr.io/factors-$(ENV)/dashboard-caching-job:$(TAG) -f Dockerfile.dashboard_caching_job .

upload-dashboard-caching: export IMAGE_NAME=dashboard-caching-job
upload-dashboard-caching: notify-deployment
	docker push us.gcr.io/factors-$(ENV)/dashboard-caching-job:$(TAG)

pack-monthly-dashboard-caching:
	docker build -t us.gcr.io/factors-$(ENV)/monthly-dashboard-caching-job:$(TAG) -f Dockerfile.monthly_dashboard_caching_job .

upload-monthly-dashboard-caching: export IMAGE_NAME=monthly-dashboard-caching-job
upload-monthly-dashboard-caching: notify-deployment
	docker push us.gcr.io/factors-$(ENV)/monthly-dashboard-caching-job:$(TAG)

pack-beam-dashboard-caching:
	docker build -t us.gcr.io/factors-$(ENV)/beam-dashboard-caching-job:$(TAG) -f Dockerfile.beam_dashboard_caching_job .

upload-beam-dashboard-caching: export IMAGE_NAME=beam-dashboard-caching-job
upload-beam-dashboard-caching: notify-deployment
	docker push us.gcr.io/factors-$(ENV)/beam-dashboard-caching-job:$(TAG)

pack-beam-dashboard-caching-now:
	docker build -t us.gcr.io/factors-$(ENV)/beam-dashboard-caching-job-now:$(TAG) -f Dockerfile.beam_dashboard_caching_job_now .

upload-beam-dashboard-caching-now: export IMAGE_NAME=beam-dashboard-caching-job-now
upload-beam-dashboard-caching-now: notify-deployment
	docker push us.gcr.io/factors-$(ENV)/beam-dashboard-caching-job-now:$(TAG)

pack-sixsignal-report:
	docker build -t us.gcr.io/factors-$(ENV)/sixsignal-report-job:$(TAG) -f Dockerfile.sixsignal_report_job .

upload-sixsignal-report: export IMAGE_NAME=sixsignal-report-job
upload-sixsignal-report: notify-deployment
	docker push us.gcr.io/factors-$(ENV)/sixsignal-report-job:$(TAG)

pack-beam-add-session:
	docker build -t us.gcr.io/factors-$(ENV)/beam-add-session-job:$(TAG) -f Dockerfile.beam_add_session_job .

upload-beam-add-session: export IMAGE_NAME=beam-add-session-job
upload-beam-add-session: notify-deployment
	docker push us.gcr.io/factors-$(ENV)/beam-add-session-job:$(TAG)

pack-web-analytics-dashboard:
	docker build -t us.gcr.io/factors-$(ENV)/web-analytics-dashboard-job:$(TAG) -f Dockerfile.web_analytics_dashboard_job .

upload-web-analytics-dashboard: export IMAGE_NAME=web-analytics-dashboard-job
upload-web-analytics-dashboard: notify-deployment
	docker push us.gcr.io/factors-$(ENV)/web-analytics-dashboard-job:$(TAG)

pack-journey-mining:
	docker build -t us.gcr.io/factors-$(ENV)/journey-mining-job:$(TAG) -f Dockerfile.journey_mining_job .

upload-journey-mining: export IMAGE_NAME=journey-mining-job
upload-journey-mining: notify-deployment
	docker push us.gcr.io/factors-$(ENV)/journey-mining-job:$(TAG)

pack-yourstory-add-properties:
	docker build -t us.gcr.io/factors-$(ENV)/yourstory-add-properties-job:$(TAG) -f Dockerfile.yourstory_add_properties_job .

upload-yourstory-add-properties: export IMAGE_NAME=yourstory-add-properties-job
upload-yourstory-add-properties: notify-deployment
	docker push us.gcr.io/factors-$(ENV)/yourstory-add-properties-job:$(TAG)

pack-instantiate-event-user-cache:
	docker build -t us.gcr.io/factors-$(ENV)/instantiate-event-user-cache-job:$(TAG) -f Dockerfile.instantiate_event_user_cache_job .

upload-instantiate-event-user-cache: export IMAGE_NAME=instantiate-event-user-cache-job
upload-instantiate-event-user-cache: notify-deployment
	docker push us.gcr.io/factors-$(ENV)/instantiate-event-user-cache-job:$(TAG)

pack-migrate-model-metadata:
	docker build -t us.gcr.io/factors-$(ENV)/migrate-model-metadata-todb-job:$(TAG) -f Dockerfile.migrate_model_metadata_job .

upload-migrate-model-metadata: export IMAGE_NAME=migrate-model-metadata-todb-job
upload-migrate-model-metadata: notify-deployment
	docker push us.gcr.io/factors-$(ENV)/migrate-model-metadata-todb-job:$(TAG)

pack-cleanup-eventuser-cache:
	docker build -t us.gcr.io/factors-$(ENV)/cleanup-eventuser-cache-job:$(TAG) -f Dockerfile.cleanup_eventuser_cache_job .

upload-cleanup-eventuser-cache: export IMAGE_NAME=cleanup-eventuser-cache-job
upload-cleanup-eventuser-cache: notify-deployment
	docker push us.gcr.io/factors-$(ENV)/cleanup-eventuser-cache-job:$(TAG)

pack-cleanup-sortedset-cache:
	docker build -t us.gcr.io/factors-$(ENV)/cleanup-sortedset-cache-job:$(TAG) -f Dockerfile.cleanup_sortedset_cache_job .

upload-cleanup-sortedset-cache: export IMAGE_NAME=cleanup-sortedset-cache-job
upload-cleanup-sortedset-cache: notify-deployment
	docker push us.gcr.io/factors-$(ENV)/cleanup-sortedset-cache-job:$(TAG)

pack-dashboard-caching-queries-comparision:
	docker build -t us.gcr.io/factors-$(ENV)/dashboard-caching-queries-comparision:$(TAG) -f Dockerfile.dashboard_caching_queries_comparision .

upload-dashboard-caching-queries-comparision: export IMAGE_NAME=dashboard-caching-queries-comparision
upload-dashboard-caching-queries-comparision: notify-deployment
	docker push us.gcr.io/factors-$(ENV)/dashboard-caching-queries-comparision:$(TAG)

pack-cleanup-dangling-keys:
	docker build -t us.gcr.io/factors-$(ENV)/cleanup-dangling-keys-job:$(TAG) -f Dockerfile.cleanup_dangling_keys_job .

upload-cleanup-dangling-keys: export IMAGE_NAME=cleanup-dangling-keys-job
upload-cleanup-dangling-keys: notify-deployment
	docker push us.gcr.io/factors-$(ENV)/cleanup-dangling-keys-job:$(TAG)

pack-rollup-sortedset-cache:
	docker build -t us.gcr.io/factors-$(ENV)/rollup-sortedset-cache-job:$(TAG) -f Dockerfile.rollup_sortedset_cache_job .

upload-rollup-sortedset-cache: export IMAGE_NAME=rollup-sortedset-cache-job
upload-rollup-sortedset-cache: notify-deployment
	docker push us.gcr.io/factors-$(ENV)/rollup-sortedset-cache-job:$(TAG)

pack-memsql-replicator:
	docker build -t us.gcr.io/factors-$(ENV)/memsql-replicator:$(TAG) -f Dockerfile.memsql_replicator .

upload-memsql-replicator: export IMAGE_NAME=memsql-replicator
upload-memsql-replicator: notify-deployment
	docker push us.gcr.io/factors-$(ENV)/memsql-replicator:$(TAG)

pack-memsql-bulk-download:
	docker build -t us.gcr.io/factors-$(ENV)/memsql-bulk-download:$(TAG) -f Dockerfile.memsql_bulk_download .

upload-memsql-bulk-download:
	docker push us.gcr.io/factors-$(ENV)/memsql-bulk-download:$(TAG)

pack-memsql-fill-user-properties:
	docker build -t us.gcr.io/factors-$(ENV)/memsql-fill-user-properties-job:$(TAG) -f Dockerfile.memsql_fill_user_properties_job .

upload-memsql-fill-user-properties:
	docker push us.gcr.io/factors-$(ENV)/memsql-fill-user-properties-job:$(TAG)

pack-copy-user-properties-migration:
	docker build -t us.gcr.io/factors-$(ENV)/copy-user-properties-migration:$(TAG) -f Dockerfile.copy_user_properties_migration .

upload-copy-user-properties-migration:
	docker push us.gcr.io/factors-$(ENV)/copy-user-properties-migration:$(TAG)

pack-pull-test-data:
	docker build -t us.gcr.io/factors-$(ENV)/pull-test-data-job:$(TAG) -f Dockerfile.pull_test_data_job .
upload-pull-test-data: export IMAGE_NAME=pull-test-data-job
upload-pull-test-data: notify-deployment
	docker push us.gcr.io/factors-$(ENV)/pull-test-data-job:$(TAG)

pack-ingest-test-data:
	docker build -t us.gcr.io/factors-$(ENV)/ingest-test-data-job:$(TAG) -f Dockerfile.ingest_test_data_job .
upload-ingest-test-data: export IMAGE_NAME=ingest-test-data-job
upload-ingest-test-data: notify-deployment
	docker push us.gcr.io/factors-$(ENV)/ingest-test-data-job:$(TAG)
pack-memsql-hubspot-sync-fields:
	docker build -t us.gcr.io/factors-$(ENV)/fix-memsql-hubspot-sync-fields:$(TAG) -f Dockerfile.fix_memsql_hubspot_sync_fields .

upload-memsql-hubspot-sync-fields:
	docker push us.gcr.io/factors-$(ENV)/fix-memsql-hubspot-sync-fields:$(TAG)

notify-deployment:
	TAG=$(TAG) $(GOPATH)/src/factors/scripts/notify_deployment.sh

build-api-doc:
	cd $(GOPATH)/src/factors;../../bin/swag init -g app/app.go

go-fmt:
	cd $(GOPATH)/src/factors;gofmt -s -w -l model handler scripts tests util sdk pattern pattern_client pattern_server data_service config

test:
	go test -failfast -v $(GOPATH)/src/factors/tests && go test -failfast -v $(GOPATH)/src/factors/pattern_server/store

pack-weekly-insights:
	docker build -t us.gcr.io/factors-$(ENV)/weekly-insights-job:$(TAG) -f Dockerfile.weekly_insights_job .

upload-weekly-insights: export IMAGE_NAME=weekly-insights-job
upload-weekly-insights: notify-deployment
	docker push us.gcr.io/factors-$(ENV)/weekly-insights-job:$(TAG)

pack-weekly-insights-mailer:
	docker build -t us.gcr.io/factors-$(ENV)/weekly-insights-mailer-job:$(TAG) -f Dockerfile.weekly_insights_mailer_job .

upload-weekly-insights-mailer: export IMAGE_NAME=weekly-insights-mailer-job
upload-weekly-insights-mailer: notify-deployment
	docker push us.gcr.io/factors-$(ENV)/weekly-insights-mailer-job:$(TAG)

pack-form-fills:
	docker build -t us.gcr.io/factors-$(ENV)/form-fills-job:$(TAG) -f Dockerfile.form_fills_job .

upload-form-fills: export IMAGE_NAME=form-fills-job
upload-form-fills: notify-deployment
	docker push us.gcr.io/factors-$(ENV)/form-fills-job:$(TAG)

pack-pathanalysis:
	docker build -t us.gcr.io/factors-$(ENV)/pathanalysis-job:$(TAG) -f Dockerfile.path_analysis_job .

upload-pathanalysis: export IMAGE_NAME=pathanalysis-job
upload-pathanalysis: notify-deployment
	docker push us.gcr.io/factors-$(ENV)/pathanalysis-job:$(TAG)

pack-bingads-integration:
	docker build -t us.gcr.io/factors-$(ENV)/bing-ads-integration-job:$(TAG) -f Dockerfile.bingads_integration_job .

upload-bingads-integration: export IMAGE_NAME=bing-ads-integration-job
upload-bingads-integration: notify-deployment
	docker push us.gcr.io/factors-$(ENV)/bing-ads-integration-job:$(TAG)

pack-currency-upload:
	docker build -t us.gcr.io/factors-$(ENV)/currency-upload-job:$(TAG) -f Dockerfile.currency_upload_job .

upload-currency-upload: export IMAGE_NAME=currency-upload-job
upload-currency-upload: notify-deployment
	docker push us.gcr.io/factors-$(ENV)/currency-upload-job:$(TAG)

pack-numerical-bucketing:
	docker build -t us.gcr.io/factors-$(ENV)/numerical-bucketing-job:$(TAG) -f Dockerfile.numerical_bucketing_job .

upload-numerical-bucketing: export IMAGE_NAME=numerical-bucketing-job
upload-numerical-bucketing: notify-deployment
	docker push us.gcr.io/factors-$(ENV)/numerical-bucketing-job:$(TAG)

pack-cleanup-stale-project:
	docker build -t us.gcr.io/factors-$(ENV)/cleanup-stale-project-job:$(TAG) -f Dockerfile.cleanup_stale_project .

upload-cleanup-stale-project: export IMAGE_NAME=cleanup-stale-project-job
upload-cleanup-stale-project: notify-deployment
	docker push us.gcr.io/factors-$(ENV)/cleanup-stale-project-job:$(TAG)

pack-replicate-properties:
	docker build -t us.gcr.io/factors-$(ENV)/replicate-properties-job:$(TAG) -f Dockerfile.replicate_properties_job .

upload-replicate-properties: export IMAGE_NAME=replicate-properties-job
upload-replicate-properties: notify-deployment
	docker push us.gcr.io/factors-$(ENV)/replicate-properties-job:$(TAG)

pack-queries-id-text-patch:
	docker build -t us.gcr.io/factors-$(ENV)/query-id-text-patch-job:$(TAG) -f Dockerfile.queries_id_text_patch_job .
upload-queries-id-text-patch: export IMAGE_NAME=query-id-text-patch-job
upload-queries-id-text-patch: notify-deployment
	docker push us.gcr.io/factors-$(ENV)/query-id-text-patch-job:$(TAG)

pack-precompile-queries:
	docker build -t us.gcr.io/factors-$(ENV)/precompile-queries-job:$(TAG) -f Dockerfile.precompile_queries_job .

upload-precompile-queries: export IMAGE_NAME=precompile-queries-job
upload-precompile-queries: notify-deployment
	docker push us.gcr.io/factors-$(ENV)/precompile-queries-job:$(TAG)

pack-convert_epoch_saved_queries_tz1_to_tz2:
	docker build -t us.gcr.io/factors-$(ENV)/convert-epoch-saved-queries-tz1-to-tz2:$(TAG) -f Dockerfile.convert_epoch_saved_queries_tz1_to_tz2 .
upload-convert_epoch_saved_queries_tz1_to_tz2: export IMAGE_NAME=convert-epoch-saved-queries-tz1-to-tz2
upload-convert_epoch_saved_queries_tz1_to_tz2: notify-deployment
	docker push us.gcr.io/factors-$(ENV)/convert-epoch-saved-queries-tz1-to-tz2:$(TAG)

pack-sdk-bundle: pack-sdk-request-worker
pack-sdk-bundle: pack-integration-request-worker
pack-sdk-bundle: pack-hubspot-enrich
pack-sdk-bundle: pack-salesforce-enrich
pack-sdk-bundle: pack-marketo-enrich
pack-sdk-bundle: pack-leadsquared-enrich
upload-sdk-bundle: upload-sdk-request-worker
upload-sdk-bundle: upload-integration-request-worker
upload-sdk-bundle: upload-hubspot-enrich
upload-sdk-bundle: upload-salesforce-enrich
upload-sdk-bundle: upload-marketo-enrich
upload-sdk-bundle: upload-leadsquared-enrich

pack-event-bundle: pack-sdk-bundle
pack-event-bundle: pack-add-session

upload-event-bundle: upload-sdk-bundle
upload-event-bundle: upload-add-session

pack-analytics-bundle: pack-api
pack-analytics-bundle: pack-dashboard-caching
pack-analytics-bundle: pack-monthly-dashboard-caching
pack-analytics-bundle: pack-monthly-beam-dashboard-caching
pack-analytics-bundle: pack-monthly-beam-dashboard-caching-now

upload-analytics-bundle: upload-api
upload-analytics-bundle: upload-dashboard-caching
upload-analytics-bundle: upload-monthly-dashboard-caching
upload-analytics-bundle: upload-monthly-beam-dashboard-caching
upload-analytics-bundle: upload-monthly-beam-dashboard-caching-now


pack-k8-backup:
	docker build -t us.gcr.io/factors-$(ENV)/k8-backup:$(TAG) -f Dockerfile.k8_backup .

upload-k8-backup: export IMAGE_NAME=k8-backup
upload-k8-backup: notify-deployment
	docker push us.gcr.io/factors-$(ENV)/k8-backup:$(TAG)

pack-compute-and-send-alerts:
	docker build -t us.gcr.io/factors-$(ENV)/compute-and-send-alerts-job:$(TAG) -f Dockerfile.compute_and_send_alerts_job .

upload-compute-and-send-alerts: export IMAGE_NAME=compute-and-send-alerts-job
upload-compute-and-send-alerts: notify-deployment
	docker push us.gcr.io/factors-$(ENV)/compute-and-send-alerts-job:$(TAG)

pack-event-trigger-alerts:
	docker build -t us.gcr.io/factors-$(ENV)/event-trigger-alerts-job:$(TAG) -f Dockerfile.event_trigger_alerts_job .

upload-event-trigger-alerts: export IMAGE_NAME=event-trigger-alerts-job
upload-event-trigger-alerts: notify-deployment
	docker push us.gcr.io/factors-$(ENV)/event-trigger-alerts-job:$(TAG)

pack-channels-to-kpi-migration:
	docker build -t us.gcr.io/factors-$(ENV)/channels-to-kpi-migration:$(TAG) -f Dockerfile.channels_v1_to_kpi_query_migration .

upload-channels-to-kpi-migration: export IMAGE_NAME=channels-to-kpi-migration
upload-channels-to-kpi-migration: notify-deployment
	docker push us.gcr.io/factors-$(ENV)/channels-to-kpi-migration:$(TAG)

pack-adhoc-db-query:
	docker build -t us.gcr.io/factors-$(ENV)/adhoc-db-query-job:$(TAG) -f Dockerfile.adhoc_db_query .
upload-adhoc-db-query: export IMAGE_NAME=adhoc-db-query-job
upload-adhoc-db-query: notify-deployment
	docker push us.gcr.io/factors-$(ENV)/adhoc-db-query-job:$(TAG)

pack-delete-older-clickable-elements:
	docker build -t us.gcr.io/factors-$(ENV)/delete-older-clickable-elements-job:$(TAG) -f Dockerfile.delete_older_clickable_elements_job .

upload-delete-older-clickable-elements: export IMAGE_NAME=delete-older-clickable-elements-job
upload-delete-older-clickable-elements: notify-deployment
	docker push us.gcr.io/factors-$(ENV)/delete-older-clickable-elements-job:$(TAG)

pack-predict-pull-events:
	docker build -t us.gcr.io/factors-$(ENV)/predict-pull-events-job:$(TAG) -f Dockerfile.predict_pull_events_job .

upload-predict-pull-events: export IMAGE_NAME=predict-pull-events-job
upload-predict-pull-events: notify-deployment
	docker push us.gcr.io/factors-$(ENV)/predict-pull-events-job:$(TAG)

pack-adhoc-yellowai-identify-fix:
	docker build -t us.gcr.io/factors-$(ENV)/adhoc-yellowai-identify-fix:$(TAG) -f Dockerfile.adhoc_yellowai_identify_fix .
upload-adhoc-yellowai-identify-fix: export IMAGE_NAME=adhoc-yellowai-identify-fix
upload-adhoc-yellowai-identify-fix: notify-deployment
	docker push us.gcr.io/factors-$(ENV)/adhoc-yellowai-identify-fix:$(TAG)

pack-adhoc-sensehq-count-touchpoints:
	docker build -t us.gcr.io/factors-$(ENV)/adhoc-sensehq-count-touchpoints:$(TAG) -f Dockerfile.adhoc_sensehq_count_touchpoints .
upload-adhoc-sensehq-count-touchpoints: export IMAGE_NAME=adhoc-sensehq-count-touchpoints
upload-adhoc-sensehq-count-touchpoints: notify-deployment
	docker push us.gcr.io/factors-$(ENV)/adhoc-sensehq-count-touchpoints:$(TAG)

pack-delete-sessions-job:
	docker build -t us.gcr.io/factors-$(ENV)/delete-sessions-job:$(TAG) -f Dockerfile.delete_sessions_job .
upload-delete-sessions-job: export IMAGE_NAME=delete-sessions-job
upload-delete-sessions-job: notify-deployment
	docker push us.gcr.io/factors-$(ENV)/delete-sessions-job:$(TAG)