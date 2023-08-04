.PHONY: build-prod deploy-prod

build-prod:
	npm run build-prod

deploy-staging: build-prod
	firebase use factors-staging; firebase deploy --only hosting:staging,hosting:staging-dyh7ken7pc

deploy-prod: build-prod
	firebase use factors-production; firebase deploy --only hosting:production,hosting:production-dyh8ken8pc