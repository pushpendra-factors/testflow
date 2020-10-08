.PHONY: pack-python-server upload-python-server pack-adwords-etl upload-adwords-etl

TAG?=
export TAG

# staging | production
ENV?=staging
export ENV

# Should be set in ~/.bashrc or ~/.profile or passed every time.
export CHANNEL_TOKEN

pack-python-server:
	docker build -t us.gcr.io/factors-$(ENV)/python-server:$(TAG) -f Dockerfile .

upload-python-server: export IMAGE_NAME=python-server
upload-python-server: notify-deployment
	docker push us.gcr.io/factors-$(ENV)/python-server:$(TAG)

pack-adwords-etl:
	docker build -t us.gcr.io/factors-$(ENV)/adwords-etl:$(TAG) -f Dockerfile.adwords_job .

upload-adwords-etl: export IMAGE_NAME=adwords-etl
upload-adwords-etl: notify-deployment
	docker push us.gcr.io/factors-$(ENV)/adwords-etl:$(TAG)

notify-deployment:
	$(GOPATH)/src/factors/scripts/notify_deployment.sh

