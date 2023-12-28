#!/bin/bash


if [ $RELEASE_TAG == "latest" ] && [ $ASSET_FILE == "assetfile" ]; then
  exit 0
else
  exit 1
fi
