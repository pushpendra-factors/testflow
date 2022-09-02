Factors Frontend.

1) Follow the first 5 steps of the ```Setup Code Repo``` section from the link https://github.com/Slashbit-Technologies/factors/wiki/Development-Setup
2) Download and install Nodejs. https://nodejs.org/en/download/
3)cd $PATH_TO_FACTORS/factors/frontend-new
4) npm install
5) npm run dev
6) localhost:3000 for frontend. (API assumed to be served from localhost:8080) 

if npm run dev fails and throws an error ```SyntaxError: Unexpected token u in JSON at position 0```. Check your npm version and downgrade it to a version below 7 and repeat from step 4. You can read more about the issue [here](https://github.com/npm/cli/issues/1995). 



# Contribution guidelines
1) -> create a new branch from master
2) -> push branch to remote (origin)
3) -> make Pull Request to master from the repo
4) -> Assign relevant person as reviewer
5) -> after approval Squash merge
6) -> Raise request for deployment in #deploy-requests slack channel (info: change, PR id, commit id, affected area of app)
7) -> After deployment post change log in deployments
