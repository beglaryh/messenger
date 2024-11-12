#!/usr/bin/env bash

aws cloudformation deploy \
  --stack-name messenger \
  --template-file template.yml \
  --region us-east-2 \
  --capabilities CAPABILITY_IAM \
  --force-upload
