.PHONY: pack-sync upload-sync

TAG?=
export TAG

# staging | production
ENV?=staging
export ENV

pack-sync:
	docker build -t us.gcr.io/factors-$(ENV)/hubspot-sync-job:$(TAG) -f Dockerfile.sync .

upload-sync:
	docker push us.gcr.io/factors-$(ENV)/hubspot-sync-job:$(TAG)
