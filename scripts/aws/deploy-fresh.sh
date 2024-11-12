#!/usr/bin/env bash

./scripts/aws/delete-all.sh
sleep 40
./scripts/aws/deploy-all.sh