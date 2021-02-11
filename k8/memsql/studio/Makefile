.PHONY: pack-studio upload-studio

# Update tag with the latest release version
# Ex format: v0.05-beta_d0e9ed9 
# git rev-list HEAD --max-count=1 --abbrev-commit
TAG?=
export TAG

# staging | production
ENV?=staging
export ENV

pack-studio: 
	docker build -t us.gcr.io/factors-$(ENV)/memsql-studio:$(TAG) -f Dockerfile.memsql_studio .

upload-studio:
	docker push us.gcr.io/factors-$(ENV)/memsql-studio:$(TAG)