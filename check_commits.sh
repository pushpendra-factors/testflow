#!/bin/sh

#constants
master="origin/master"
staging="origin/staging"
release="origin/release"
curbranch="origin/${GITHUB_HEAD_REF#refs/heads/}"
basebranch="origin/${GITHUB_BASE_REF#refs/heads/}"


# turning off hints
git config --global advice.resolveConflict false
git config --global advice.commitBeforeMerge false

# check out current branch
git config --global user.name "Github Actions"
git config --global user.email "actions@github.com"
echo Hello
git fetch --unshallow -q
git checkout -q $curbranch


# Count number of commits on current branch
total_commits=$(git rev-list --count HEAD ^$basebranch)

# To check conflicts when number of commits more than 1 
if [[ "$curbranch" != "$master" && "$curbranch" != "$release" && "$curbranch" != "$staging" ]]; then
    if [[ "$total_commits" -gt 1 ]]; then
        printf "Merge failed! Your $curbranch branch has more than 1 commit; please squash all commits.\n"
        exit -1
    fi
fi
