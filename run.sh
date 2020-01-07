#!/bin/bash

helm template   \
    -f ./tmp/name_space.yaml \
    -f ./tmp/containers.yaml \
    -f ./tmp/manifest.yaml \
    -f ./tmp/vault.yaml \
    -f ./tmp/secrets.yaml \
    -f ./tmp/cmdb-0.0.4/repo/cmdb/cmdb.yaml \
    -n $(cat ./tmp/release_name.txt) \
  ./tmp/cmdb-0.0.4/repo/cmdb