.PHONY: pack-adwords-sync upload-adwords-sync

TAG?=
export TAG

# staging | production
ENV?=staging
export ENV

# Should be set in ~/.bashrc or ~/.profile or passed every time.
export CHANNEL_TOKEN

pack-adwords-sync:
	docker build -t us.gcr.io/factors-$(ENV)/adwords-sync:$(TAG) -f Dockerfile.sync .

upload-adwords-sync: export IMAGE_NAME=adwords-sync
upload-adwords-sync: notify-deployment
	docker push us.gcr.io/factors-$(ENV)/adwords-sync:$(TAG)

notify-deployment:
	$(GOPATH)/src/factors/scripts/notify_deployment.sh
