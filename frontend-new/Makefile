.PHONY: build-staging deploy-staging 

#  export FACTORS_REPO=~/repos/factors

build-staging:
	npm run build-staging

build-prod:
	npm run build-prod


deploy-staging: build-staging
	firebase use factors-staging; firebase deploy --only hosting:staging; firebase deploy --only hosting:flash-staging

deploy-prod: build-prod
	firebase use factors-production; firebase deploy --only hosting:production

