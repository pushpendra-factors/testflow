.PHONY: build-api serve-api build-ps serve-ps clean pack-api upload-api pack-ps upload-ps pack-debug upload-debug pack-pg-backup upload-pg-backup test

TAG?=$(shell git rev-list HEAD --max-count=1 --abbrev-commit)
export TAG


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
	docker build -t us.gcr.io/factors-staging/app-server:$(TAG) -f Dockerfile.appserver .

upload-api:
	docker push us.gcr.io/factors-staging/app-server:$(TAG)

pack-ps:
	docker build -t us.gcr.io/factors-staging/pattern-server:$(TAG) -f Dockerfile.patternserver .

upload-ps:
	docker push us.gcr.io/factors-staging/pattern-server:$(TAG)

pack-debug:
	docker build --no-cache -t us.gcr.io/factors-staging/debug:$(TAG) -f Dockerfile.debug .

upload-debug:
	docker push us.gcr.io/factors-staging/debug:$(TAG)

pack-pg-backup:
	docker build --no-cache -t us.gcr.io/factors-staging/pg-backup:$(TAG) -f Dockerfile.pg_backup .

upload-pg-backup:
	docker push us.gcr.io/factors-staging/pg-backup:$(TAG)

test: 
	go test -v $(GOPATH)/src/factors/tests && go test -v $(GOPATH)/src/factors/pattern_server/store