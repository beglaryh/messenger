build:
	go build -o bin/helloworld main.go

onconnect:
	GOARCH=amd64 GOOS=linux go build -tags="lambda.norpc" -ldflags="-w -s" -o bin/onconnect/bootstrap lambda/onconnect/main.go
	cp bin/onconnect/bootstrap ./
	zip messenger.onconnect.zip bootstrap
	mv messenger.onconnect.zip bin/onconnect/messenger.onconnect.zip
	rm bootstrap

ondisconnect:
	GOARCH=amd64 GOOS=linux go build -tags="lambda.norpc" -ldflags="-w -s" -o bin/ondisconnect/bootstrap lambda/ondisconnect/main.go
	cp bin/ondisconnect/bootstrap ./
	zip messenger.ondisconnect.zip bootstrap
	mv messenger.ondisconnect.zip bin/ondisconnect/messenger.ondisconnect.zip
	rm bootstrap

roomcreate:
	GOARCH=amd64 GOOS=linux go build -tags="lambda.norpc" -ldflags="-w -s" -o bin/roomcreate/bootstrap lambda/roomcreate/main.go
	cp bin/roomcreate/bootstrap ./
	zip messenger.roomcreate.zip bootstrap
	mv messenger.roomcreate.zip bin/roomcreate/messenger.roomcreate.zip
	rm bootstrap

sendmessage:
	GOARCH=amd64 GOOS=linux go build -tags="lambda.norpc" -ldflags="-w -s" -o bin/sendmessage/bootstrap lambda/sendmessage/main.go
	cp bin/sendmessage/bootstrap ./
	zip messenger.sendmessage.zip bootstrap
	mv messenger.sendmessage.zip bin/sendmessage/messenger.sendmessage.zip
	rm bootstrap

fetch:
	GOARCH=amd64 GOOS=linux go build -tags="lambda.norpc" -ldflags="-w -s" -o bin/fetch/bootstrap lambda/fetch/main.go
	cp bin/fetch/bootstrap ./
	zip messenger.fetch.zip bootstrap
	mv messenger.fetch.zip bin/fetch/messenger.fetch.zip
	rm bootstrap

