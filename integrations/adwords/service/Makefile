.PHONY: pack upload

TAG?=
export TAG

# staging | production
ENV?=staging
export ENV

# Should be set in ~/.bashrc or ~/.profile or passed every time.
export CHANNEL_TOKEN

pack:
	docker build -t us.gcr.io/factors-$(ENV)/adwords-server:$(TAG) .

upload: export IMAGE_NAME=adwords-server
upload: notify-deployment
	docker push us.gcr.io/factors-$(ENV)/adwords-server:$(TAG)

notify-deployment:
	$(GOPATH)/src/factors/scripts/notify_deployment.sh

