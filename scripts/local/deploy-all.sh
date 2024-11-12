#!/usr/bin/env bash

make api
make cron

awslocal s3api create-bucket --bucket hrach-lambda-dev-build --region us-east-2 --create-bucket-configuration LocationConstraint=us-east-2

awslocal s3 cp bin/api/schedule.api.zip s3://hrach-lambda-dev-build
awslocal s3 cp bin/cron/schedule.cron.zip s3://hrach-lambda-dev-build

./scripts/local/deploy-stack.sh