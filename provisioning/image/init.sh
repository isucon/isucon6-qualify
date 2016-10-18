#!/bin/sh
# init script for isucon6-qualifier

set -ex

export DEBIAN_FRONTEND=noninteractive
apt update
apt install -y --no-install-recommends ansible git aptitude
apt remove -y snapd

mkdir -p -m 700 /root/.ssh
wget -O /root/.ssh/id_rsa http://isucon6q.example.com/id_rsa
chmod 600 /root/.ssh/id_rsa
ssh-keyscan -t rsa github.com >> /root/.ssh/known_hosts
export HOME=/root
git config --global user.name "isucon"
git config --global user.email "isucon@isucon.net"

git clone git@github.com:isucon/isucon6-qualify /tmp/isucon6-qualifier
cd /tmp/isucon6-qualifier/provisioning/image/ansible
PYTHONUNBUFFERED=1 ANSIBLE_FORCE_COLOR=true ansible-playbook -i localhost, *.yml --connection=local -t prod
cd /tmp/isucon6-qualifier/provisioning/image
./db_setup.sh
cd /tmp && rm -rf /tmp/isucon6-qualifier && rm -rf /tmp/ansible-tmp-*
shutdown -r now

