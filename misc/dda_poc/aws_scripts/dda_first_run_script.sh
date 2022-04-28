# Meant to be stored at S3 and to be run whenever a new cluster is launched.

source /etc/profile.d/lang.sh
export LANG=en_US.UTF-8
export LANGUAGE=en_US.UTF-8
export LC_COLLATE=C
export LC_CTYPE=en_US.UTF-8

sudo yum -y update
sudo yum -y install git-core
sudo yum -y install htop

REPO_PATH=~/repos
FACTORS_PATH=$REPO_PATH/factors
DDA_PATH=$FACTORS_PATH/misc/dda_poc

mkdir -p $REPO_PATH
cd $REPO_PATH

git config --global user.name “Govind Sharma”
git config --global user.email govind@factors.ai
git config --global core.editor nano
git clone https://govind-factors:6298590df2fd6f62c9ecd4a9cfd93837908fb8b0@github.com/Slashbit-Technologies/factors.git

cd $FACTORS_PATH
git checkout poc-data-attribution

cd $DDA_PATH/../

sudo /emr/notebook-env/bin/pip install tabulate
