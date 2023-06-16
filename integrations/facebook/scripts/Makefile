.PHONY: pack-facebook-sync upload-facebook-sync

TAG?=
export TAG

# staging | production
ENV?=staging
export ENV

# Should be set in ~/.bashrc or ~/.profile or passed every time.
export CHANNEL_TOKEN

pack-facebook-sync:
	docker build -t us.gcr.io/factors-$(ENV)/facebook-sync-job:$(TAG) -f Dockerfile.sync .

upload-facebook-sync: export IMAGE_NAME=facebook-sync-job
upload-facebook-sync: notify-deployment
	docker push us.gcr.io/factors-$(ENV)/facebook-sync-job:$(TAG)

notify-deployment:
	$(GOPATH)/src/factors/scripts/notify_deployment.sh
