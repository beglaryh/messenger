#!/usr/bin/env bash

awslocal s3 rm s3://hrach-lambda-dev-build/schedule.api.zip
awslocal s3 rm s3://hrach-lambda-dev-build/schedule.cron.zip

awslocal cloudformation delete-stack \
    --region us-east-2 \
    --stack-name schedule    