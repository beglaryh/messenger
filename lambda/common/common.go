package common

import (
	"context"
	"log"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
)

func GetConfig() aws.Config {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Println(err)
		panic(err)
	}

	return cfg
}

func BaseEndpoint(r *events.APIGatewayWebsocketProxyRequest) string {
	return "https://" + r.RequestContext.DomainName + "/" + r.RequestContext.Stage
}
