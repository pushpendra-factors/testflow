.PHONY: pack upload serve

TAG?=
export TAG

# staging | production
ENV?=staging
export ENV

build:
	docker build -t php-service:latest -f Dockerfile .

serve:
	docker run -p=3000:3000 php-service

pack:
	docker build -t us.gcr.io/factors-$(ENV)/php-service:$(TAG) -f Dockerfile .

upload: export IMAGE_NAME=php-service
upload: notify deployment
	docker push us.gcr.io/factors-$(ENV)/php-service:$(TAG)