.PHONY: build-staging pack-fe upload-fe

# Update tag with the latest release version
# Ex format: v0.09-beta_8bfe4e0
# git rev-list HEAD --max-count=1 --abbrev-commit
TAG?=
export TAG


build-staging:
	npm run build-staging

pack-fe: build-staging
	docker build -t us.gcr.io/factors-staging/frontend:$(TAG) -f Dockerfile.frontend .

upload-fe:
	docker push us.gcr.io/factors-staging/frontend:$(TAG)