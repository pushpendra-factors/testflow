.PHONY: pack-linkedin-sync upload-linkedin-sync pack-linkedin-company-sync upload-linkedin-company-sync

TAG?=
export TAG

# staging | production
ENV?=staging
export ENV

# Should be set in ~/.bashrc or ~/.profile or passed every time.
export CHANNEL_TOKEN

pack-linkedin-sync-dup:
	docker build -t us.gcr.io/factors-$(ENV)/linkedin-sync-job-dup:$(TAG) -f Dockerfile.sync .

upload-linkedin-sync-dup: export IMAGE_NAME=linkedin-sync-job-dup
upload-linkedin-sync-dup: notify-deployment
	docker push us.gcr.io/factors-$(ENV)/linkedin-sync-job-dup:$(TAG)

pack-linkedin-company-sync:
	docker build -t us.gcr.io/factors-$(ENV)/linkedin-company-sync-job:$(TAG) -f Dockerfile.company_sync .

upload-linkedin-company-sync: export IMAGE_NAME=linkedin-company-sync-job
upload-linkedin-company-sync: notify-deployment
	docker push us.gcr.io/factors-$(ENV)/linkedin-company-sync-job:$(TAG)

notify-deployment:
	$(GOPATH)/src/factors/scripts/notify_deployment.sh