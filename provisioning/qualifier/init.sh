#!/bin/sh
# init script for isucon6-qualifier

set -ex

export DEBIAN_FRONTEND=noninteractive
apt update
apt install -y --no-install-recommends ansible git aptitude
apt remove -y snapd

cd /tmp &&  wget -O- https://isucon6qimage.example.com/isucon6q/isucon6-qualifier.tar.gz | tar zxvf -
cd /tmp/isucon6-qualifier/provisioning/image/ansible
PYTHONUNBUFFERED=1 ANSIBLE_FORCE_COLOR=true ansible-playbook -i localhost, *.yml --connection=local -t prod
cd /tmp/isucon6-qualifier/provisioning/image
./db_setup.sh
cd /tmp && rm -rf /tmp/isucon6-qualifier && rm -rf /tmp/ansible-tmp-*
shutdown -r now

