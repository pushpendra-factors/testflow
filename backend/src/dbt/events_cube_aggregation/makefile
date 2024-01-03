.PHONY: pack-website-aggregation-job upload-website-aggregation-job pack-website-aggregation-deploy upload-website-aggregation-deploy

TAG?=
export TAG

ENV?=staging
export ENV

pack-website-aggregation-job:
	docker build -t us.gcr.io/factors-$(ENV)/website-aggregation-job:$(TAG) -f Dockerfile.dbt_website_aggregation_job .

upload-website-aggregation-job: export IMAGE_NAME=website-aggregation-job
upload-website-aggregation-job:
	docker push us.gcr.io/factors-$(ENV)/website-aggregation-job:$(TAG)

pack-website-aggregation-deploy:
	docker build -t us.gcr.io/factors-$(ENV)/website-aggregation-deploy:$(TAG) -f Dockerfile.dbt_website_aggregation_deploy .

upload-website-aggregation-deploy: export IMAGE_NAME=website-aggregation-job
upload-website-aggregation-deploy:
	docker push us.gcr.io/factors-$(ENV)/website-aggregation-deploy:$(TAG)
