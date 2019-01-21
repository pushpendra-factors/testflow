.PHONY: build-api build-ps clean pack-api upload-api pack-ps upload-ps

TAG?=$(shell git rev-list HEAD --max-count=1 --abbrev-commit)
export TAG


build-api:
	go build -o $(GOPATH)/bin/app $(GOPATH)/src/factors/app/app.go

build-ps:
	go build -o $(GOPATH)/bin/pattern-app $(GOPATH)/src/factors/pattern_server/cmd/pattern-app.go

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
