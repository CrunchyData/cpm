#!/bin/bash
#

# Copyright 2015 Crunchy Data Solutions, Inc.
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
# http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# basic user installation script for CPM
#
# This script assumes a registered RHEL 7 server is the installation host OS.
#

# Exit installation on any unexpected error
set -e

# set the install directory
export INSTALLDIR=/var/cpm
# set the development directory
export DEVDIR=/home/jeffmc/devproject/github.com/crunchydata/crunchy-postgresql-manager

sudo mkdir -p $INSTALLDIR/bin
sudo mkdir -p $INSTALLDIR/config
sudo mkdir -p $INSTALLDIR/www
sudo mkdir -p $INSTALLDIR/keys
sudo mkdir -p $INSTALLDIR/data
sudo mkdir -p $INSTALLDIR/data/promdash
sudo mkdir -p $INSTALLDIR/data/prometheus
sudo mkdir -p $INSTALLDIR/data/etcd
sudo mkdir -p $INSTALLDIR/data/pgsql
export LOGDIR=$INSTALLDIR/logs
sudo mkdir -p $LOGDIR

# prompt for static IP to use
echo "enter static ip to use for this host... "
read STATICIP
# prompt for domain name to use
echo "enter domain name to use... "
read DOMAIN

#Check if current user is member to the wheel group
username= whoami
if groups $username | grep &>/dev/null 'wheel'; then
	echo "Group permissions ok"
else
	echo "You must have sudo privledges to run this install"
	exit
fi

# don't error if packages are already installed
set +e

# install deps
sudo yum -y install docker
sudo rpm -Uvh http://yum.postgresql.org/9.4/redhat/rhel-7-x86_64/pgdg-centos94-9.4-1.noarch.rpm
sudo yum install -y epel-release postgresql94 postgresql94-contrib postgresql94-server libxslt unzip openssh-clients hostname bind-utils net-tools sysstat

set -e

# move the CPM media to the /var/cpm installation directory
sudo cp -r /home/jeffmc/devproject/bin/* $INSTALLDIR/bin
sudo cp -r $DEVDIR/config/* $INSTALLDIR/config
sudo cp -r $DEVDIR/www/* $INSTALLDIR/www

# start up local cpmserverapi
echo "starting cpmserverapi...."
sudo cp $INSTALLDIR/config/cpmserverapi.service /etc/systemd/system
sudo systemctl enable cpmserverapi.service
sudo systemctl start cpmserverapi.service

sed -i "s/crunchy.lab/$DOMAIN/g" ./bu-init-cpm.sh

# generate keys for cpm and cpm-admin
echo "generating keys for cpm and cpm-admin containers..."
sudo $INSTALLDIR/bin/gen-keys.sh
