.PHONY: build-staging build-prod pack-fe-staging pack-fe-prod upload-fe-staging upload-fe-prod

# Update tag with the latest release version
# Ex format: v0.09-beta_8bfe4e0
# git rev-list HEAD --max-count=1 --abbrev-commit
TAG?=
export TAG

#  export FACTORS_REPO=~/repos/factors

build-staging:
	npm run build-staging

build-prod:
	npm run build-prod

pack-fe-staging: build-staging
	docker build -t us.gcr.io/factors-staging/frontend:$(TAG) -f Dockerfile.staging .

pack-fe-prod: build-prod
	docker build -t us.gcr.io/factors-production/frontend:$(TAG) -f Dockerfile.production .

upload-fe-prod:
	docker push us.gcr.io/factors-production/frontend:$(TAG)

upload-fe-staging:
	docker push us.gcr.io/factors-staging/frontend:$(TAG)