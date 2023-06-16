.PHONY: pack-device-service upload-device-service serve-device-service

TAG?=
export TAG

# staging | production
ENV?=staging
export ENV

build-device-service:
	docker build -t php-service:latest -f Dockerfile .

serve-device-service:
	docker run -p=3000:3000 php-service

pack-device-service:
	docker build -t us.gcr.io/factors-$(ENV)/php-service:$(TAG) -f Dockerfile .

upload-device-service: export IMAGE_NAME=php-service
upload-device-service: notify-deployment
	docker push us.gcr.io/factors-$(ENV)/php-service:$(TAG)
	
notify-deployment:
	$(GOPATH)/src/factors/scripts/notify_deployment.sh