.PHONY: build-staging pack-fe upload-fe

# Update tag with the latest release version
# Ex format: v0.09-beta_8bfe4e0
# git rev-list HEAD --max-count=1 --abbrev-commit
TAG?=
export TAG

# staging | production
ENV?=staging
export ENV

#  export FACTORS_REPO=~/repos/factors

build-staging:
	npm run build-staging

build-production:
	npm run build-prod

pack-fe: build-$(ENV)
	docker build -t us.gcr.io/factors-$(ENV)/frontend:$(TAG) -f Dockerfile.frontend .

upload-fe:
	docker push us.gcr.io/factors-$(ENV)/frontend:$(TAG)