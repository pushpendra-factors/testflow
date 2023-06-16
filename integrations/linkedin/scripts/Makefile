.PHONY: pack-linkedin-sync upload-linkedin-sync

TAG?=
export TAG

# staging | production
ENV?=staging
export ENV

# Should be set in ~/.bashrc or ~/.profile or passed every time.
export CHANNEL_TOKEN

pack-linkedin-sync:
	docker build -t us.gcr.io/factors-$(ENV)/linkedin-sync-job:$(TAG) -f Dockerfile.sync .

upload-linkedin-sync: export IMAGE_NAME=linkedin-sync-job
upload-linkedin-sync: notify-deployment
	docker push us.gcr.io/factors-$(ENV)/linkedin-sync-job:$(TAG)

notify-deployment:
	$(GOPATH)/src/factors/scripts/notify_deployment.sh