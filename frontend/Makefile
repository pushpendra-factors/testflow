.PHONY: build-staging build-prod deploy-staging deploy-prod notify-deployment

#  export FACTORS_REPO=~/repos/factors

build-staging:
	npm run build-staging

build-prod:
	npm run build-prod

deploy-staging: build-staging
	firebase use factors-staging; firebase deploy --only hosting:staging

deploy-prod: notify-deployment
deploy-prod: build-prod
	firebase use factors-production; firebase deploy --only hosting:production

notify-deployment:
	$(FACTORS_REPO)/frontend/notify_deployment.sh
