#!/bin/env sh
tfplugingen-framework generate all --input schema.json --output generated
cp generated/provider_*/*.go provider/
for res in $(jq -r .resources[].name schema.json); do
  test -d ${res} || mkdir ${res}
  cp generated/resource_${res}/*.go ${res}/
done
rm -rf generated
