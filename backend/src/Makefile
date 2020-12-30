.PHONY: build-api serve-api build-ps serve-ps build-ds serve-ds build-sdk serve-sdk serve-redis serve-predis clean pack-api upload-api pack-ps upload-ps pack-ds upload-ds pack-sdk upload-sdk pack-redis upload-redis pack-predis upload-predis pack-debug upload-debug pack-build-seq upload-build-seq pack-build-reports upload-build-reports pack-hubspot-enrich upload-hubspot-enrich pack-sdk-request-worker upload-sdk-request-worker pack-integration-request-worker upload-integration-request-worker pack-monitoring upload-monitoring pack-archive-events upload-archive-events pack-adhoc-archive-events upload-adhoc-archive-events pack-bigquery-upload upload-bigquery-upload pack-onboard-to-bigquery upload-onboard-to-bigquery pack-add-session upload-add-session pack-dashboard-caching upload-dashboard-caching pack-web-analytics-dashboard upload-web-analytics-dashboard pack-instantiate-event-user-cache upload-instantiate-event-user-cache pack-journey-mining upload-journey-mining build-api-doc test

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

build-ps:
	go build -o $(GOPATH)/bin/pattern-app $(GOPATH)/src/factors/pattern_server/cmd/pattern-app.go

serve-ps:
	$(GOPATH)/bin/pattern-app

build-ds:
	rm $(GOPATH)/bin/data-service; go build -o $(GOPATH)/bin/data-service $(GOPATH)/src/factors/data_service/data-service.go

serve-ds:
	$(GOPATH)/bin/data-service

build-sdk:
	rm $(GOPATH)/bin/sdk-service; go build -o $(GOPATH)/bin/sdk-service $(GOPATH)/src/factors/sdk_service/sdk-service.go

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

pack-build-reports:
	docker build -t us.gcr.io/factors-$(ENV)/build-reports-job:$(TAG) -f Dockerfile.build_reports .

upload-build-reports: export IMAGE_NAME=build-reports-job
upload-build-reports: notify-deployment
	docker push us.gcr.io/factors-$(ENV)/build-reports-job:$(TAG)

pack-add-session:
	docker build -t us.gcr.io/factors-$(ENV)/add-session-job:$(TAG) -f Dockerfile.add_session_job .

upload-add-session: export IMAGE_NAME=add-session-job
upload-add-session: notify-deployment
	docker push us.gcr.io/factors-$(ENV)/add-session-job:$(TAG)

pack-hubspot-enrich:
	docker build -t us.gcr.io/factors-$(ENV)/hubspot-enrich-job:$(TAG) -f Dockerfile.hubspot_enrich_job .

upload-hubspot-enrich: export IMAGE_NAME=hubspot-enrich-job
upload-hubspot-enrich: notify-deployment
	docker push us.gcr.io/factors-$(ENV)/hubspot-enrich-job:$(TAG)

pack-salesforce-enrich:
	docker build -t us.gcr.io/factors-$(ENV)/salesforce-enrich-job:$(TAG) -f Dockerfile.salesforce_enrich_job .

upload-salesforce-enrich: export IMAGE_NAME=salesforce-enrich-job
upload-salesforce-enrich: notify-deployment
	docker push us.gcr.io/factors-$(ENV)/salesforce-enrich-job:$(TAG)

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

pack-cleanup-eventuser-cache: 
	docker build -t us.gcr.io/factors-$(ENV)/cleanup-eventuser-cache-job:$(TAG) -f Dockerfile.cleanup_eventuser_cache_job .

upload-cleanup-eventuser-cache: export IMAGE_NAME=cleanup-eventuser-cache-job
upload-cleanup-eventuser-cache: notify-deployment
	docker push us.gcr.io/factors-$(ENV)/cleanup-eventuser-cache-job:$(TAG)

pack-adwords-migration-job: 
	docker build -t us.gcr.io/factors-$(ENV)/adwords-migration-job:$(TAG) -f Dockerfile.adwords_migration_job .

upload-adwords-migration-job: export IMAGE_NAME=adwords-migration-job
upload-adwords-migration-job: notify-deployment
	docker push us.gcr.io/factors-$(ENV)/adwords-migration-job:$(TAG)

notify-deployment:
	$(GOPATH)/src/factors/scripts/notify_deployment.sh

build-api-doc:
	cd $(GOPATH)/src/factors;../../bin/swag init -g app/app.go

test:
	go test -failfast -v $(GOPATH)/src/factors/tests && go test -failfast -v $(GOPATH)/src/factors/pattern_server/store
