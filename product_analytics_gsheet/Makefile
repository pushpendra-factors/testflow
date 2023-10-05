.PHONY: pack-product-analytics-job upload-product-analytics-job serve-product-analytics-job

TAG?=
export TAG

# staging | production
ENV?=staging
export ENV

build-product-analytics-job:
	docker build -t product-analytics-job:latest -f Dockerfile .

serve-product-analytics-job:
	docker run --name product-analytics product-analytics-job

pack-product-analytics-job:
	docker build -t us.gcr.io/factors-$(ENV)/product-analytics-job:$(TAG) -f Dockerfile .

upload-product-analytics-job: export IMAGE_NAME=product-analytics-job
upload-product-analytics-job: notify-deployment
	docker push us.gcr.io/factors-$(ENV)/product-analytics-job:$(TAG)
	
notify-deployment:
	$(GOPATH)/src/factors/scripts/notify_deployment.sh
