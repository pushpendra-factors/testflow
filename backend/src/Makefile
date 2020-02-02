.PHONY: build-api serve-api build-ps serve-ps build-ds serve-ds serve-redis clean pack-api upload-api pack-ps upload-ps pack-ds upload-ds pack-redis upload-redis pack-debug upload-debug pack-build-seq upload-build-seq pack-build-reports upload-build-reports pack-hubspot-enrich upload-hubspot-enrich test 

# Update tag with the latest release version
# Ex format: v0.05-beta_d0e9ed9 
# git rev-list HEAD --max-count=1 --abbrev-commit
TAG?=
export TAG

# staging | production
ENV?=staging
export ENV

build-api:
	go build -o $(GOPATH)/bin/app $(GOPATH)/src/factors/app/app.go

serve-api:
	$(GOPATH)/bin/app

build-ps:
	go build -o $(GOPATH)/bin/pattern-app $(GOPATH)/src/factors/pattern_server/cmd/pattern-app.go

serve-ps:
	$(GOPATH)/bin/pattern-app

build-ds:
	go build -o $(GOPATH)/bin/data-service $(GOPATH)/src/factors/data_service/data-service.go

serve-ds:
	$(GOPATH)/bin/data-service

serve-redis:
	redis-server $(GOPATH)/src/redis.conf

clean:
	rm $(GOPATH)/bin/pattern-app $(GOPATH)/bin/app $(GOPATH)/bin/data-service

pack-api:
	docker build -t us.gcr.io/factors-$(ENV)/app-server:$(TAG) -f Dockerfile.appserver .

upload-api:
	docker push us.gcr.io/factors-$(ENV)/app-server:$(TAG)

pack-ps:
	docker build -t us.gcr.io/factors-$(ENV)/pattern-server:$(TAG) -f Dockerfile.patternserver .

upload-ps:
	docker push us.gcr.io/factors-$(ENV)/pattern-server:$(TAG)

pack-ds:
	docker build -t us.gcr.io/factors-$(ENV)/data-server:$(TAG) -f Dockerfile.dataserver .

upload-ds:
	docker push us.gcr.io/factors-$(ENV)/data-server:$(TAG)

pack-redis:
	docker build -t us.gcr.io/factors-$(ENV)/redis:$(TAG) -f Dockerfile.redis .

upload-redis:
	docker push us.gcr.io/factors-$(ENV)/redis:$(TAG)

pack-debug:
	docker build --no-cache -t us.gcr.io/factors-$(ENV)/debug:$(TAG) -f Dockerfile.debug .

upload-debug:
	docker push us.gcr.io/factors-$(ENV)/debug:$(TAG)

pack-build-seq:
	docker build -t us.gcr.io/factors-$(ENV)/build-seq-job:$(TAG) -f Dockerfile.build_seq_job .

upload-build-seq:
	docker push us.gcr.io/factors-$(ENV)/build-seq-job:$(TAG)

pack-build-reports:
	docker build -t us.gcr.io/factors-$(ENV)/build-reports-job:$(TAG) -f Dockerfile.build_reports .

upload-build-reports:
	docker push us.gcr.io/factors-$(ENV)/build-reports-job:$(TAG)

pack-hubspot-enrich:
	docker build -t us.gcr.io/factors-$(ENV)/hubspot-enrich-job:$(TAG) -f Dockerfile.hubspot_enrich_job .

upload-hubspot-enrich:
	docker push us.gcr.io/factors-$(ENV)/hubspot-enrich-job:$(TAG)

test: 
	go test -v $(GOPATH)/src/factors/tests && go test -v $(GOPATH)/src/factors/pattern_server/store
