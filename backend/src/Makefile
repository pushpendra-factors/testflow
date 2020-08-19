.PHONY: build-api serve-api build-ps serve-ps build-ds serve-ds build-sdk serve-sdk serve-redis serve-predis clean pack-api upload-api pack-ps upload-ps pack-ds upload-ds pack-sdk upload-sdk pack-redis upload-redis pack-predis upload-predis pack-debug upload-debug pack-build-seq upload-build-seq pack-build-reports upload-build-reports pack-hubspot-enrich upload-hubspot-enrich pack-sdk-request-worker upload-sdk-request-worker pack-integration-request-worker upload-integration-request-worker pack-monitoring upload-monitoring pack-archive-events upload-archive-events pack-bigquery-upload upload-bigquery-upload pack-onboard-to-bigquery upload-onboard-to-bigquery pack-add-session upload-add-session pack-dashboard-caching upload-dashboard-caching pack-web-analytics-dashboard upload-web-analytics-dashboard test

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

clean:
	rm $(GOPATH)/bin/pattern-app $(GOPATH)/bin/app $(GOPATH)/bin/data-service

pack-api: export IMAGE_NAME=app-server
pack-api: notify-deployment
	docker build -t us.gcr.io/factors-$(ENV)/app-server:$(TAG) -f Dockerfile.appserver .

upload-api:
	docker push us.gcr.io/factors-$(ENV)/app-server:$(TAG)

pack-ps: export IMAGE_NAME=pattern-server
pack-ps: notify-deployment
	docker build -t us.gcr.io/factors-$(ENV)/pattern-server:$(TAG) -f Dockerfile.patternserver .

upload-ps:
	docker push us.gcr.io/factors-$(ENV)/pattern-server:$(TAG)

pack-ds: export IMAGE_NAME=data-server
pack-ds: notify-deployment
	docker build -t us.gcr.io/factors-$(ENV)/data-server:$(TAG) -f Dockerfile.dataserver .

upload-ds:
	docker push us.gcr.io/factors-$(ENV)/data-server:$(TAG)

pack-sdk: export IMAGE_NAME=sdk-server
pack-sdk: notify-deployment
	docker build -t us.gcr.io/factors-$(ENV)/sdk-server:$(TAG) -f Dockerfile.sdkserver .

upload-sdk:
	docker push us.gcr.io/factors-$(ENV)/sdk-server:$(TAG)

pack-redis:
	docker build -t us.gcr.io/factors-$(ENV)/redis:$(TAG) -f Dockerfile.redis .

upload-redis:
	docker push us.gcr.io/factors-$(ENV)/redis:$(TAG)

pack-predis:
	docker build -t us.gcr.io/factors-$(ENV)/persistent-redis:$(TAG) -f Dockerfile.persistent_redis .

upload-predis:
	docker push us.gcr.io/factors-$(ENV)/persistent-redis:$(TAG)

pack-debug: export IMAGE_NAME=debug
pack-debug: notify-deployment
	docker build --no-cache -t us.gcr.io/factors-$(ENV)/debug:$(TAG) -f Dockerfile.debug .

upload-debug:
	docker push us.gcr.io/factors-$(ENV)/debug:$(TAG)

pack-build-seq: export IMAGE_NAME=build-seq-job
pack-build-seq: notify-deployment
	docker build -t us.gcr.io/factors-$(ENV)/build-seq-job:$(TAG) -f Dockerfile.build_seq_job .

upload-build-seq:
	docker push us.gcr.io/factors-$(ENV)/build-seq-job:$(TAG)

pack-build-reports: export IMAGE_NAME=build-reports-job
pack-build-reports: notify-deployment
	docker build -t us.gcr.io/factors-$(ENV)/build-reports-job:$(TAG) -f Dockerfile.build_reports .

upload-build-reports:
	docker push us.gcr.io/factors-$(ENV)/build-reports-job:$(TAG)

pack-add-session: export IMAGE_NAME=add-session-job
pack-add-session: notify-deployment
	docker build -t us.gcr.io/factors-$(ENV)/add-session-job:$(TAG) -f Dockerfile.add_session_job .

upload-add-session:
	docker push us.gcr.io/factors-$(ENV)/add-session-job:$(TAG)

pack-hubspot-enrich: export IMAGE_NAME=hubspot-enrich-job
pack-hubspot-enrich: notify-deployment
	docker build -t us.gcr.io/factors-$(ENV)/hubspot-enrich-job:$(TAG) -f Dockerfile.hubspot_enrich_job .

upload-hubspot-enrich:
	docker push us.gcr.io/factors-$(ENV)/hubspot-enrich-job:$(TAG)

pack-sdk-request-worker: export IMAGE_NAME=sdk-request-worker
pack-sdk-request-worker: notify-deployment
	docker build -t us.gcr.io/factors-$(ENV)/sdk-request-worker:$(TAG) -f Dockerfile.sdk_request_worker .

upload-sdk-request-worker:
	docker push us.gcr.io/factors-$(ENV)/sdk-request-worker:$(TAG)

pack-integration-request-worker: export IMAGE_NAME=integration-request-worker
pack-integration-request-worker: notify-deployment
	docker build -t us.gcr.io/factors-$(ENV)/integration-request-worker:$(TAG) -f Dockerfile.integration_request_worker .

upload-integration-request-worker:
	docker push us.gcr.io/factors-$(ENV)/integration-request-worker:$(TAG)

pack-monitoring: export IMAGE_NAME=monitoring
pack-monitoring: notify-deployment
	docker build -t us.gcr.io/factors-$(ENV)/monitoring:$(TAG) -f Dockerfile.monitoring .

upload-monitoring:
	docker push us.gcr.io/factors-$(ENV)/monitoring:$(TAG)

pack-archive-events: export IMAGE_NAME=archive-events-job
pack-archive-events: notify-deployment
	docker build -t us.gcr.io/factors-$(ENV)/archive-events-job:$(TAG) -f Dockerfile.archive_events_job .

upload-archive-events:
	docker push us.gcr.io/factors-$(ENV)/archive-events-job:$(TAG)

pack-bigquery-upload: export IMAGE_NAME=bigquery-upload-job
pack-bigquery-upload: notify-deployment
	docker build -t us.gcr.io/factors-$(ENV)/bigquery-upload-job:$(TAG) -f Dockerfile.bigquery_upload_job .

upload-bigquery-upload:
	docker push us.gcr.io/factors-$(ENV)/bigquery-upload-job:$(TAG)

pack-onboard-to-bigquery: export IMAGE_NAME=onboard-to-bigquery-job
pack-onboard-to-bigquery: notify-deployment
	docker build -t us.gcr.io/factors-$(ENV)/onboard-to-bigquery-job:$(TAG) -f Dockerfile.onboard_to_bigquery_job .

upload-onboard-to-bigquery:
	docker push us.gcr.io/factors-$(ENV)/onboard-to-bigquery-job:$(TAG)

pack-merge-user-properties: export IMAGE_NAME=merge-user-properties-job
pack-merge-user-properties: notify-deployment
	docker build -t us.gcr.io/factors-$(ENV)/merge-user-properties-job:$(TAG) -f Dockerfile.merge_user_properties_job .

upload-merge-user-properties:
	docker push us.gcr.io/factors-$(ENV)/merge-user-properties-job:$(TAG)

pack-dashboard-caching: export IMAGE_NAME=dashboard-caching-job
pack-dashboard-caching: notify-deployment
	docker build -t us.gcr.io/factors-$(ENV)/dashboard-caching-job:$(TAG) -f Dockerfile.dashboard_caching_job .

upload-dashboard-caching:
	docker push us.gcr.io/factors-$(ENV)/dashboard-caching-job:$(TAG)

pack-web-analytics-dashboard: export IMAGE_NAME=web-analytics-dashboard-job
pack-web-analytics-dashboard: notify-deployment
	docker build -t us.gcr.io/factors-$(ENV)/web-analytics-dashboard-job:$(TAG) -f Dockerfile.web_analytics_dashboard_job .

upload-web-analytics-dashboard:
	docker push us.gcr.io/factors-$(ENV)/web-analytics-dashboard-job:$(TAG)

pack-build-yourstory-add-properties: export IMAGE_NAME=yourstory-add-properties-job
pack-build-yourstory-add-properties: 
	docker build -t us.gcr.io/factors-$(ENV)/yourstory-add-properties-job:$(TAG) -f Dockerfile.yourstory_add_properties_job .

upload-build-yourstory-add-properties: notify-deployment
	docker push us.gcr.io/factors-$(ENV)/yourstory-add-properties-job:$(TAG)

notify-deployment:
	$(GOPATH)/src/factors/scripts/notify_deployment.sh

test:
	go test -failfast -v $(GOPATH)/src/factors/tests && go test -failfast -v $(GOPATH)/src/factors/pattern_server/store
