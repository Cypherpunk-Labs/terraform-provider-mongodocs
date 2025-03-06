#!/bin/bash
VERSION="0.1.7"
mkdir release
env GOOS=linux GOARCH=amd64 go build -o release/terraform-provider-mongodocs_v${VERSION}
cd release
zip terraform-provider-mongodocs_v${VERSION}_linux_amd64.zip terraform-provider-mongodocs_v${VERSION}
rm terraform-provider-mongodocs_v${VERSION}
cd ..
env GOOS=darwin GOARCH=arm64 go build -o release/terraform-provider-mongodocs_v${VERSION}
cd release
zip terraform-provider-mongodocs_v${VERSION}_darwin_arm64.zip terraform-provider-mongodocs_v${VERSION}
rm terraform-provider-mongodocs_v${VERSION}
cp ../terraform-registry-manifest.json terraform-provider-mongodocs_v${VERSION}_manifest.json
shasum -a 256 *.zip > terraform-provider-mongodocs_v${VERSION}_SHA256SUMS
shasum -a 256 *.json >> terraform-provider-mongodocs_v${VERSION}_SHA256SUMS
gpg --detach-sign terraform-provider-mongodocs_v${VERSION}_SHA256SUMS
echo "Finished Release Build"
