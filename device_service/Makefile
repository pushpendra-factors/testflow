.PHONY: pack-device-service upload-device-service serve-device-service

TAG?=
export TAG

# staging | production
ENV?=staging
export ENV

build-device-server:
	docker build -t device-server:latest -f Dockerfile.device_service .

serve-device-server:
	docker run -p=3000:3000 device-server

pack-device-server:
	docker build -t us.gcr.io/factors-$(ENV)/device-server:$(TAG) -f Dockerfile.device_service .

upload-device-server: export IMAGE_NAME=device-server
upload-device-server: notify-deployment
	docker push us.gcr.io/factors-$(ENV)/device-server:$(TAG)
	
notify-deployment:
	$(GOPATH)/src/factors/scripts/notify_deployment.sh
