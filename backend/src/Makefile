.PHONY: build-api serve-api build-ps serve-ps clean pack-api upload-api pack-ps upload-ps pack-debug upload-debug pack-build-seq upload-build-seq test

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

clean:
	rm $(GOPATH)/bin/pattern-app $(GOPATH)/bin/app

pack-api:
	docker build -t us.gcr.io/factors-$(ENV)/app-server:$(TAG) -f Dockerfile.appserver .

upload-api:
	docker push us.gcr.io/factors-$(ENV)/app-server:$(TAG)

pack-ps:
	docker build -t us.gcr.io/factors-$(ENV)/pattern-server:$(TAG) -f Dockerfile.patternserver .

upload-ps:
	docker push us.gcr.io/factors-$(ENV)/pattern-server:$(TAG)

pack-debug:
	docker build --no-cache -t us.gcr.io/factors-$(ENV)/debug:$(TAG) -f Dockerfile.debug .

upload-debug:
	docker push us.gcr.io/factors-$(ENV)/debug:$(TAG)

pack-build-seq:
	docker build -t us.gcr.io/factors-$(ENV)/build-seq-job:$(TAG) -f Dockerfile.build_seq_job .

upload-build-seq:
	docker push us.gcr.io/factors-$(ENV)/build-seq-job:$(TAG)
test: 
	go test -v $(GOPATH)/src/factors/tests && go test -v $(GOPATH)/src/factors/pattern_server/store
