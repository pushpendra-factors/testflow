.PHONY: build deploy

build:
	npm run build

deploy: build
	firebase use factors-production; firebase deploy --only hosting:factors-website