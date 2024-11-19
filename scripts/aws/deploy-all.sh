#!/usr/bin/env bash

make onconnect
make ondisconnect
make roomcreate
make sendmessage
make fetch
make edit

aws s3api create-bucket --bucket hrach-lambda-dev-build --region us-east-2 --create-bucket-configuration LocationConstraint=us-east-2
aws s3 cp bin/onconnect/messenger.onconnect.zip s3://hrach-lambda-dev-build
aws s3 cp bin/ondisconnect/messenger.ondisconnect.zip s3://hrach-lambda-dev-build
aws s3 cp bin/roomcreate/messenger.roomcreate.zip s3://hrach-lambda-dev-build
aws s3 cp bin/sendmessage/messenger.sendmessage.zip s3://hrach-lambda-dev-build
aws s3 cp bin/fetch/messenger.fetch.zip s3://hrach-lambda-dev-build
aws s3 cp bin/edit/messenger.edit.zip s3://hrach-lambda-dev-build

./scripts/aws/deploy-stack.sh
