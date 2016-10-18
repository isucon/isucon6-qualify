#!/bin/sh
# init script for isucon6-qualifier

set -ex

export DEBIAN_FRONTEND=noninteractive
apt update
apt install -y ansible git aptitude
apt remove -y snapd

mkdir -p -m 700 /root/.ssh
wget -O /root/.ssh/id_rsa http://isucon6q.example.com/id_rsa
chmod 600 /root/.ssh/id_rsa
ssh-keyscan -t rsa github.com >> /root/.ssh/known_hosts
export HOME=/root
git config --global user.name "isucon"
git config --global user.email "isucon@isucon.net"

git clone git@github.com:isucon/isucon6-qualify /tmp/isucon6-qualifier
cd /tmp/isucon6-qualifier/provisioning/bench
PYTHONUNBUFFERED=1 ANSIBLE_FORCE_COLOR=true ansible-playbook -i localhost, ansible/*.yml --connection=local
cd /tmp && rm -rf /tmp/isucon6-qualifier
curl https://github.com/{Songmu,motemen,tatsuru,edvakf,catatsuy,walf443,st-cyrill,myfinder,aereal,tarao,yuuki}.keys >> /home/isucon/.ssh/authorized_keys

