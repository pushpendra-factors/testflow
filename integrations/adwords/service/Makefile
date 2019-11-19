.PHONY: pack upload

TAG?=
export TAG

# staging | production
ENV?=staging
export ENV

pack:
	docker build -t us.gcr.io/factors-$(ENV)/adwords-server:$(TAG) .

upload:
	docker push us.gcr.io/factors-$(ENV)/adwords-server:$(TAG)



