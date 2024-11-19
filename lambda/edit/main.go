package main

import (
	"context"
	"encoding/json"
	"log"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/apigatewaymanagementapi"
	"github.com/beglaryh/messenger/infrastructure/database"
	"github.com/beglaryh/messenger/lambda/common"
	"github.com/beglaryh/messenger/presentation/item/editrequestitem"
)

var (
	cfg aws.Config   = common.GetConfig()
	db  *database.DB = database.New(cfg)
)

func handler(_ context.Context, r events.APIGatewayWebsocketProxyRequest) (events.APIGatewayProxyResponse, error) {
	var item editrequestitem.EditRequest

	if err := json.Unmarshal([]byte(r.Body), &item); err != nil {
		log.Println(err)
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
		}, nil
	}

	message, err := db.EditMessage(item.Message.To())
	if err != nil {
		return events.APIGatewayProxyResponse{StatusCode: 500}, nil
	}

	connections, err := db.GetConnectionsByRoom(message.RoomId)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
		}, nil
	}

	endpoint := common.BaseEndpoint(&r)
	config := cfg
	config.BaseEndpoint = &endpoint
	client := apigatewaymanagementapi.NewFromConfig(config)

	data, err := json.Marshal(message)
	if err != nil {
		log.Println(err)
		return events.APIGatewayProxyResponse{StatusCode: 500}, nil
	}

	for _, connection := range connections {
		_, err := client.PostToConnection(context.TODO(), &apigatewaymanagementapi.PostToConnectionInput{
			ConnectionId: &connection.PK,
			Data:         data,
		})
		if err != nil {
			log.Println("error sending message")
			log.Println(err)
			return events.APIGatewayProxyResponse{StatusCode: 500}, nil
		}
	}

	return events.APIGatewayProxyResponse{
		StatusCode: 200,
	}, nil
}

func main() {
	lambda.Start(handler)
}
