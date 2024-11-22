#!/usr/bin/env bash

aws s3 rm s3://hrach-lambda-dev-build/messenger.onconnect.zip
aws s3 rm s3://hrach-lambda-dev-build/messenger.ondisconnect.zip
aws s3 rm s3://hrach-lambda-dev-build/messenger.request.zip

aws cloudformation delete-stack \
  --region us-east-2 \
  --stack-name messenger
