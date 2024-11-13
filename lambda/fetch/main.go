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
	"github.com/beglaryh/messenger/presentation/item/fetchrequest"
	"github.com/beglaryh/messenger/presentation/item/fetchresponse"
)

var (
	cfg aws.Config   = common.GetConfig()
	db  *database.DB = database.New(cfg)
)

func handler(_ context.Context, r events.APIGatewayWebsocketProxyRequest) (events.APIGatewayProxyResponse, error) {
	cid := r.RequestContext.ConnectionID
	conn, _ := db.GetConnection(cid)
	uid := conn.UID
	var request fetchrequest.FetchRequest

	if err := json.Unmarshal([]byte(r.Body), &request); err != nil {
		log.Println(err)
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
		}, nil
	}

	lastMessage := request.Message
	messages, err := db.GetMessagesAfter(lastMessage, uid)
	if err != nil {
		log.Println(err)
		return events.APIGatewayProxyResponse{StatusCode: 500}, nil
	}

	response := fetchresponse.FetchResponse{Messages: *messages}
	data, _ := json.Marshal(response)

	endpoint := common.BaseEndpoint(&r)
	config := cfg
	config.BaseEndpoint = &endpoint

	client := apigatewaymanagementapi.NewFromConfig(config)

	_, err = client.PostToConnection(context.TODO(), &apigatewaymanagementapi.PostToConnectionInput{
		ConnectionId: &cid,
		Data:         data,
	})
	if err != nil {
		log.Println(err)
		return events.APIGatewayProxyResponse{StatusCode: 500}, nil
	}

	return events.APIGatewayProxyResponse{
		StatusCode: 200,
	}, nil
}

func main() {
	lambda.Start(handler)
}
