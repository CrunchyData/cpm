
#!/bin/bash -x

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

#
# script to copy docker server related files to destination
#
GOPATH=/home/jeffmc/devproject
CPMPATH=$GOPATH/src/github.com/crunchydata/crunchy-postgresql-manager
adminserver=localhost
remoteservers=(localhost)

for i in "${remoteservers[@]}"
do
	echo $i
	ssh root@$i "mkdir -p /var/cpm/bin"
	scp $GOPATH/bin/*  \
	$CPMPATH/sbin/*  \
	$CPMPATH/sql/*  \
	root@$i:/var/cpm/bin/
	scp  $CPMPATH/config/cpmserverapi.service  \
	 root@$i:/usr/lib/systemd/system
        ssh root@$i "systemctl enable cpmserverapi.service"
done

# copy all required admin files to the admin server

ssh root@$adminserver "mkdir -p /var/cpm/data/promdash"
ssh root@$adminserver "mkdir -p /var/cpm/bin"
scp $GOPATH/bin/* \
$CPMPATH/sbin/* \
root@$adminserver:/var/cpm/bin

scp $CPMPATH/config/file.sqlite3 root@$adminserver:/var/cpm/data/promdash/
scp $CPMPATH/sbin/*.pem root@$adminserver:/var/cpm/keys
scp $CPMPATH/config/cpmserverapi.service root@$adminserver:/usr/lib/systemd/system

ssh root@$adminserver "systemctl enable cpmserverapi.service"


