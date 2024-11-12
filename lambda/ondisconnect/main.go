package main

import (
	"context"
	"log"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/beglaryh/messenger/infrastructure/database"
)

var db *database.DB = database.New(getConfig())

func handler(_ context.Context, r events.APIGatewayWebsocketProxyRequest) (events.APIGatewayProxyResponse, error) {
	cid := r.RequestContext.ConnectionID

	if err := db.RemoveConnection(cid); err != nil {
		log.Println(err)
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
			Body:       "internal server error",
		}, err
	}

	return events.APIGatewayProxyResponse{
		StatusCode: 200,
	}, nil
}

func main() {
	lambda.Start(handler)
}

func getConfig() aws.Config {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Println(err)
		panic(err)
	}

	return cfg
}
