.PHONY: build pack upload

# Update tag with the latest release version
# Ex format: v0.09-beta_8bfe4e0
# git rev-list HEAD --max-count=1 --abbrev-commit
TAG?=
export TAG

#  export FACTORS_REPO=~/repos/factors

build:
	npm run build

pack: build
	docker build -t us.gcr.io/factors-production/website:$(TAG) .

upload:
	docker push us.gcr.io/factors-production/website:$(TAG)