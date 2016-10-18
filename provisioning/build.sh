#!/bin/sh

workdir=$(cd $(dirname $0) && pwd)
cd $workdir
for app in portal bench qualifier; do
    cat $app/deploy.json.template \
        | jq --arg base64 "$(cat ${app}/init.sh | base64 -w0)" '(.resources | .[] | select( .type == "Microsoft.Compute/virtualMachines") | .properties.osProfile.customData) |= $base64'  \
             > $app/deploy.json
done

cd $workdir/../../ && tar czvf $workdir/isucon6-qualifier.tar.gz isucon6-qualifier/db  isucon6-qualifier/bin  isucon6-qualifier/provisioning/image/ansible isucon6-qualifier/provisioning/image/files isucon6-qualifier/provisioning/image/db_setup.sh

