.PHONY: pack-hubspot-sync upload-hubspot-sync

TAG?=
export TAG

# staging | production
ENV?=staging
export ENV

# Should be set in ~/.bashrc or ~/.profile or passed every time.
export CHANNEL_TOKEN

pack-hubspot-sync:
	docker build -t us.gcr.io/factors-$(ENV)/hubspot-sync-job:$(TAG) -f Dockerfile.sync .

upload-hubspot-sync: export IMAGE_NAME=hubspot-sync-job
upload-hubspot-sync: notify-deployment
	docker push us.gcr.io/factors-$(ENV)/hubspot-sync-job:$(TAG)

notify-deployment:
	$(GOPATH)/src/factors/scripts/notify_deployment.sh
