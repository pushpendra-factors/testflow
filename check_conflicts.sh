#!/bin/sh

#constants
master="origin/master"
staging="origin/staging"
release="origin/release"
merge_conflict_msg="Automatic merge failed; fix conflicts"

# turning off hints
git config --global advice.resolveConflict false
git config --global advice.commitBeforeMerge false

# # checking out current branch
git config --global user.name "Github Actions"
git config --global user.email "actions@github.com"
git fetch --unshallow -q
curbranch="origin/${GITHUB_HEAD_REF#refs/heads/}"
basebranch="origin/${GITHUB_BASE_REF#refs/heads/}"
git checkout -q $curbranch

# checks for merge conflict
CheckMergeConflict(){
 if git merge --no-commit --no-ff -q "$1" | grep -q "$merge_conflict_msg"; then
    printf "$curbranch has conflicts with $1 "
    git merge --abort
    exit 1
 else
    printf "$curbranch has no conflicts with $1 \n "
    git merge --abort
    return
 fi   
}

# check for release branch 
if [ "$curbranch" == "$release" ]; then
    exit
else

# To check conflicts when base branch is master
    if [ "$basebranch" == "$master" ]; then
        CheckMergeConflict $staging
        CheckMergeConflict $release

# To check conflicts when base branch is staging
    elif [ "$basebranch" == "$staging" ]; then
         CheckMergeConflict $release
    fi
fi