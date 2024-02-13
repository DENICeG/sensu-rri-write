#!/bin/bash

if [[ -z "$1" ]]; then
  echo "need tag/version in format v1.x.y"
  exit 1
else
  TAG=$1
fi

if [[ ! -f "./cmd/sensu-rri-write/orderfile" ]]; then
  echo "./cmd/sensu-rri-write/orderfile is missing"
  exit 1
fi

CGO_ENABLED=0 go build -o ./bin/sensu-rri-write ./cmd/sensu-rri-write

tar czf sensu-rri-write_${TAG}_linux_amd64.tar.gz bin/

sha512sum sensu-rri-write_${TAG}_linux_amd64.tar.gz > sensu-rri-write_${TAG}_sha512_checksums.txt
SHA_HASH_ONLY=$(cut -d " " -f 1 sensu-rri-write_${TAG}_sha512_checksums.txt)

sed "s/__TAG__/${TAG}/g" sensu/asset_template.tpl > sensu/asset.yaml
sed -i "s/__SHA__/${SHA_HASH_ONLY}/g" sensu/asset.yaml

mkdir -p artifacts
rm -f artifacts/*
mv sensu-rri-write_${TAG}_linux_amd64.tar.gz sensu-rri-write_${TAG}_sha512_checksums.txt artifacts/

git add .
git commit -m $TAG
git tag $TAG
git push && git push --tags
