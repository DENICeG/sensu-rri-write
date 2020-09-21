#!/bin/bash

if [[ -z "$1" ]]; then 
  echo "need tag/version in format v1.x.y"
  exit 1
else
  TAG=$1
fi

if ! which packr2 >/dev/null; then 
  echo packr2 binary is not installed
  echo "(go get github.com/gobuffalo/packr/v2/packr2)"
  exit 1
fi

cd cmd/sensu-rri-write
packr2
CGO_ENABLED=0 go build -o ../../bin/sensu-rri-write main.go
packr2 clean
cd ../..

tar czf sensu-rri-write_${TAG}_linux_amd64.tar.gz bin/

sha512sum sensu-rri-write_${TAG}_linux_amd64.tar.gz > sensu-rri-write_${TAG}_sha512_checksums.txt
SHA_HASH_ONLY=$(cut -d " " -f 1 sensu-rri-write_${TAG}_sha512_checksums.txt)

sed "s/__TAG__/${TAG}/g" sensu/asset_template.tpl > sensu/asset.yaml
sed -i "s/__SHA__/${SHA_HASH_ONLY}/g" sensu/asset.yaml

mkdir -p artifacts
rm -f artifacts/*
mv sensu-rri-write_${TAG}_linux_amd64.tar.gz sensu-rri-write_${TAG}_sha512_checksums.txt artifacts/

git add .
git commit
git tag $TAG
git push && git push --tags
