#!/usr/bin/env bash

./scripts/local/delete-all.sh
echo "Waiting for stack deletion to finish"
sleep 45
./scripts/local/deploy-all.sh