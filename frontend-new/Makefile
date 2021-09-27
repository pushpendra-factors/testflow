.PHONY: build-staging deploy-staging 

#  export FACTORS_REPO=~/repos/factors

build-staging:
	npm run build-staging

build-prod:
	npm run build-prod


deploy-staging: build-staging
	firebase use factors-staging; firebase deploy --only hosting:staging; firebase deploy --only hosting:sloth-staging; node sdk-monitor.js stage 2

deploy-prod: notify-deployment
deploy-prod: build-prod
	firebase use factors-production; firebase deploy --only hosting:production; firebase deploy --only hosting:flash-production; node sdk-monitor.js prod 4

notify-deployment:
	$(FACTORS_REPO)/frontend-new/notify_deployment.sh
