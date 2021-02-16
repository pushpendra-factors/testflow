.PHONY: pack-sync upload-sync

TAG?=
export TAG

# staging | production
ENV?=staging
export ENV

# Should be set in ~/.bashrc or ~/.profile or passed every time.
export CHANNEL_TOKEN

pack-sync:
	docker build -t us.gcr.io/factors-$(ENV)/linkedin-sync-job:$(TAG) -f Dockerfile.sync .

upload-sync: export IMAGE_NAME=linkedin-sync-job
upload-sync: notify-deployment
	docker push us.gcr.io/factors-$(ENV)/linkedin-sync-job:$(TAG)

notify-deployment:
	$(GOPATH)/src/factors/scripts/notify_deployment.sh