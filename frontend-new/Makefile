.PHONY: build-staging deploy-staging 

#  export FACTORS_REPO=~/repos/factors

build-staging:
	npm run build-staging

build-prod:
	npm run build-prod


deploy-staging: build-staging
	firebase use factors-staging; firebase deploy --only hosting:staging; node sdk-monitor.js stage 2; node sdk-monitor.js staging 2 https://b2mxnuvcer.firebaseapp.com/factors.v1.js

deploy-prod: notify-deployment
deploy-prod: build-prod
	firebase use factors-production; firebase deploy --only hosting:production; node sdk-monitor.js prod 4; node sdk-monitor.js prod 4 https://b3mxnuvcer.firebaseapp.com/factors.v1.js; https://b3mxnuvcer.firebaseapp.com/b3mxnuvcer.js; https://dyh8ken8pc.firebaseapp.com/dyh8ken8pc.js

deploy-violet: build-staging
	firebase use factors-staging; firebase deploy --only hosting:violet

deploy-blue: build-staging
	firebase use factors-staging; firebase deploy --only hosting:blue	

notify-deployment:
	$(FACTORS_REPO)/frontend-new/notify_deployment.sh
