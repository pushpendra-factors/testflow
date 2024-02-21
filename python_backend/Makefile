.PHONY: pack-python-server upload-python-server pack-adwords-etl upload-adwords-etl pack-chat-data upload-chat-data pack-adwords-etl-debug upload-adwords-etl-debug pack-facebook-etl upload-facebook-etl

TAG?=
export TAG

# staging | production
ENV?=staging
export ENV

# Should be set in ~/.bashrc or ~/.profile or passed every time.
export CHANNEL_TOKEN

pack-python-server:
	docker build -t us.gcr.io/factors-$(ENV)/python-server:$(TAG) -f Dockerfile .

upload-python-server: export IMAGE_NAME=python-server
upload-python-server: notify-deployment
	docker push us.gcr.io/factors-$(ENV)/python-server:$(TAG)

pack-adwords-etl:
	docker build -t us.gcr.io/factors-$(ENV)/adwords-etl:$(TAG) -f Dockerfile.adwords_job .

upload-adwords-etl: export IMAGE_NAME=adwords-etl
upload-adwords-etl: notify-deployment
	docker push us.gcr.io/factors-$(ENV)/adwords-etl:$(TAG)

pack-chat-data:
	docker build -t us.gcr.io/factors-$(ENV)/chat-data:$(TAG) -f Dockerfile.chat_data_job .

upload-chat-data: export IMAGE_NAME=chat-data
upload-chat-data: notify-deployment
	docker push us.gcr.io/factors-$(ENV)/chat-data:$(TAG)

notify-deployment:
	$(GOPATH)/src/factors/scripts/notify_deployment.sh

pack-adwords-etl-debug:
	docker build --no-cache -t us.gcr.io/factors-$(ENV)/adwords-etl-debug:$(TAG) -f Dockerfile.adwords_debug .
upload-adwords-etl-debug: export IMAGE_NAME=adwords-etl-debug
upload-adwords-etl-debug: notify-deployment
	docker push us.gcr.io/factors-$(ENV)/adwords-etl-debug:$(TAG)

pack-facebook-etl:
	docker build -t us.gcr.io/factors-$(ENV)/facebook-etl:$(TAG) -f Dockerfile.facebook_job .
upload-facebook-etl: export IMAGE_NAME=facebook-etl
upload-facebook-etl: notify-deployment
	docker push us.gcr.io/factors-$(ENV)/facebook-etl:$(TAG)

pack-google-search-console-etl:
	docker build -t us.gcr.io/factors-$(ENV)/google-search-console-etl:$(TAG) -f Dockerfile.google_search_console_job .
upload-google-search-console-etl: export IMAGE_NAME=google-search-console-etl
upload-google-search-console-etl: notify-deployment
	docker push us.gcr.io/factors-$(ENV)/google-search-console-etl:$(TAG)

