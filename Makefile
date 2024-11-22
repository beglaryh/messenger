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

request:
	GOARCH=amd64 GOOS=linux go build -tags="lambda.norpc" -ldflags="-w -s" -o bin/request/bootstrap lambda/request/main.go
	cp bin/request/bootstrap ./
	zip messenger.request.zip bootstrap
	mv messenger.request.zip bin/request/messenger.request.zip
	rm bootstrap
