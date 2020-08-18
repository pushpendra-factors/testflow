.PHONY: build-staging deploy-staging 

#  export FACTORS_REPO=~/repos/factors

build-staging:
	npm run build-staging


deploy-staging: build-staging
	firebase use factors-staging; firebase deploy --only hosting:tufte-staging

