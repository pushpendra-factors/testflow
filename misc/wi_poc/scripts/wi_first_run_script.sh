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
ENV_PATH=$REPO_PATH/wi_env
FACTORS_PATH=$REPO_PATH/factors
WI_PATH=$FACTORS_PATH/misc/wi_poc

mkdir -p $REPO_PATH
cd $REPO_PATH

git config --global user.name “Govind Sharma”
git config --global user.email govind@factors.ai
git config --global core.editor nano
git clone https://govind-factors:6298590df2fd6f62c9ecd4a9cfd93837908fb8b0@github.com/Slashbit-Technologies/factors.git

cd $FACTORS_PATH
git checkout automate_weekly_insights

python3 -m venv $ENV_PATH    # Can import wi_env from S3 as well.
source $ENV_PATH/bin/activate

pip3 install -r $WI_PATH/requirements.txt

cd $WI_PATH/../
