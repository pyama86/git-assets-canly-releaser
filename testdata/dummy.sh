#!/bin/bash


echo "RELEASE_TAG: $RELEASE_TAG ASSET_FILE: $ASSET_FILE"
if [ $RELEASE_TAG == "latest" ] && [ $ASSET_FILE == "assetfile" ]; then
  exit 0
else
  exit 1
fi
