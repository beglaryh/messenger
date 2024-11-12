package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/beglaryh/gocommon/time/offsetdatetime"
	"github.com/beglaryh/messenger/domain/room"
	"github.com/beglaryh/messenger/infrastructure/database"
	"github.com/google/uuid"
)

var db *database.DB = database.New(getConfig())

func handler(_ context.Context, r events.APIGatewayWebsocketProxyRequest) (events.APIGatewayProxyResponse, error) {
	cid := r.RequestContext.ConnectionID
	uid, ok := r.Headers["uid"]

	if !ok {
		log.Println("missing uid")
		return events.APIGatewayProxyResponse{
			StatusCode: 400,
			Body:       "bad request",
		}, errors.New("bad request")
	}

	var room room.Room
	var body map[string]string

	_ = json.Unmarshal([]byte(r.Body), &body)

	if err := json.Unmarshal([]byte(body["message"]), &room); err != nil {
		log.Println(err)
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
		}, nil
	}
	room.Id = uuid.New()
	room.CreatedOn = offsetdatetime.Now()
	room.CreatedBy = uid
	if err := db.SaveRoom(room); err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
		}, nil
	}
	if err := db.AppendRoomToConnection(cid, room); err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
		}, nil
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
