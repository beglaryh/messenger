#!/usr/bin/env bash

aws --endpoint-url=http://localhost:4566 cloudformation deploy \
    --stack-name schedule \
    --template-file template.yml \
    --region us-east-2 \
    --capabilities CAPABILITY_IAM \
    --force-upload