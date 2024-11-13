package main

import (
	"context"
	"encoding/json"
	"log"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/apigatewaymanagementapi"
	"github.com/beglaryh/gocommon/collection/list/arraylist"
	"github.com/beglaryh/gocommon/time/offsetdatetime"
	"github.com/beglaryh/messenger/domain/connection"
	"github.com/beglaryh/messenger/infrastructure/database"
	"github.com/beglaryh/messenger/lambda/common"
	"github.com/beglaryh/messenger/presentation/item/createroom"
	"github.com/google/uuid"
)

var (
	cfg aws.Config   = common.GetConfig()
	db  *database.DB = database.New(cfg)
)

func handler(_ context.Context, r events.APIGatewayWebsocketProxyRequest) (events.APIGatewayProxyResponse, error) {
	cid := r.RequestContext.ConnectionID
	conn, _ := db.GetConnection(cid)
	uid := conn.UID
	var item createroom.CreateRoomItem

	if err := json.Unmarshal([]byte(r.Body), &item); err != nil {
		log.Println(err)
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
		}, nil
	}
	room := item.Message
	room.Id = uuid.New()
	room.CreatedOn = offsetdatetime.Now()
	room.CreatedBy = uid

	if err := db.SaveRoom(room); err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
		}, nil
	}

	connections := arraylist.New[connection.Connection]()
	for _, user := range room.Members {
		userConnections, err := db.GetConnectionsByUserId(user)
		if err != nil {
			return events.APIGatewayProxyResponse{StatusCode: 500}, nil
		}
		connections.Add(userConnections...)
	}

	if err := db.AppendRoomToConnections(connections.ToArray(), room); err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
		}, nil
	}

	response := map[string]string{"id": room.Id.String()}
	data, _ := json.Marshal(response)

	endpoint := common.BaseEndpoint(&r)
	config := cfg
	config.BaseEndpoint = &endpoint

	client := apigatewaymanagementapi.NewFromConfig(config)

	_, err := client.PostToConnection(context.TODO(), &apigatewaymanagementapi.PostToConnectionInput{
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

func getConfig() aws.Config {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Println(err)
		panic(err)
	}

	return cfg
}
