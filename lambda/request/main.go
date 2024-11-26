package main

import (
	"context"
	"encoding/json"
	"log"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/apigatewaymanagementapi"
	"github.com/beglaryh/gocommon/collection/list/arraylist"
	"github.com/beglaryh/gocommon/time/offsetdatetime"
	"github.com/beglaryh/messenger/domain/connection"
	"github.com/beglaryh/messenger/domain/editrequest"
	"github.com/beglaryh/messenger/domain/message"
	"github.com/beglaryh/messenger/domain/reaction"
	"github.com/beglaryh/messenger/domain/room"
	"github.com/beglaryh/messenger/infrastructure/database"
	"github.com/beglaryh/messenger/lambda/common"
	"github.com/beglaryh/messenger/presentation/item/fetchresponse"
	"github.com/beglaryh/messenger/presentation/item/reactionrequestitem"
	"github.com/beglaryh/messenger/presentation/item/requestitem"
	"github.com/google/uuid"
)

var (
	cfg aws.Config   = common.GetConfig()
	db  *database.DB = database.New(cfg)
)

func handler(_ context.Context, r events.APIGatewayWebsocketProxyRequest) (events.APIGatewayProxyResponse, error) {
	var event requestitem.EventItem
	if err := json.Unmarshal([]byte(r.Body), &event); err != nil {
		log.Println(err)
		return events.APIGatewayProxyResponse{StatusCode: 500}, nil
	}

	internalMessage, _ := json.Marshal(event.Message.Message)
	switch event.Message.Action {
	case "create":
		var room room.Room
		if err := json.Unmarshal(internalMessage, &room); err != nil {
			return events.APIGatewayProxyResponse{StatusCode: 400, Body: "invalid request"}, nil
		}
		return createRoom(room, &r)
	case "send":
		var message message.Message
		if err := json.Unmarshal(internalMessage, &message); err != nil {
			return events.APIGatewayProxyResponse{StatusCode: 400, Body: "invalid request"}, nil
		}
		return sendMessage(message, &r)
	case "edit":
		var edit editrequest.EditRequest
		if err := json.Unmarshal(internalMessage, &edit); err != nil {
			return events.APIGatewayProxyResponse{StatusCode: 400, Body: "invalid request"}, nil
		}
		return editMessage(edit, &r)
	case "react":
		var request reactionrequestitem.ReactionRequestItem
		if err := json.Unmarshal(internalMessage, &request); err != nil {
			return events.APIGatewayProxyResponse{StatusCode: 400, Body: "invalid request"}, nil
		}
		return react(request, &r)
	case "fetch":
		cursor := event.Message.Message.(string)
		return fetch(cursor, &r)
	default:
		return events.APIGatewayProxyResponse{StatusCode: 400, Body: "illegal request"}, nil
	}
}

func createRoom(room room.Room, r *events.APIGatewayWebsocketProxyRequest) (events.APIGatewayProxyResponse, error) {
	baseEndpoint := common.BaseEndpoint(r)
	cid := r.RequestContext.ConnectionID
	conn, _ := db.GetConnection(cid)

	room.Id = uuid.New()
	room.CreatedBy = conn.UID
	room.CreatedOn = offsetdatetime.Now()

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

	config := cfg
	config.BaseEndpoint = &baseEndpoint

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

func sendMessage(message message.Message, r *events.APIGatewayWebsocketProxyRequest) (events.APIGatewayProxyResponse, error) {
	cid := r.RequestContext.ConnectionID
	conn, _ := db.GetConnection(cid)

	roomOpt := db.GetRoomHeader(message.RoomId)
	if roomOpt.IsEmpty() {
		log.Println("room not found: " + message.RoomId)
		return events.APIGatewayProxyResponse{StatusCode: 400}, nil
	}
	room, _ := roomOpt.Get()
	message.Members = room.Members
	message.SentBy = conn.UID

	mid, err := uuid.NewV7()
	if err != nil {
		log.Println(err)
		return events.APIGatewayProxyResponse{StatusCode: 500}, nil
	}
	message.Id = mid.String()
	message.CreatedOn = offsetdatetime.Now()

	if err := db.SaveMessage(message); err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
		}, nil
	}

	connections, err := db.GetConnectionsByRoom(message.RoomId)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
		}, nil
	}

	endpoint := common.BaseEndpoint(r)
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

func editMessage(request editrequest.EditRequest, r *events.APIGatewayWebsocketProxyRequest) (events.APIGatewayProxyResponse, error) {
	message, err := db.EditMessage(request)
	if err != nil {
		return events.APIGatewayProxyResponse{StatusCode: 500}, nil
	}

	connections, err := db.GetConnectionsByRoom(message.RoomId)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
		}, nil
	}

	endpoint := common.BaseEndpoint(r)
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

func fetch(lastMessage string, r *events.APIGatewayWebsocketProxyRequest) (events.APIGatewayProxyResponse, error) {
	cid := r.RequestContext.ConnectionID
	conn, _ := db.GetConnection(cid)
	uid := conn.UID

	messages, err := db.GetMessagesAfter(lastMessage, uid)
	if err != nil {
		log.Println(err)
		return events.APIGatewayProxyResponse{StatusCode: 500}, nil
	}

	response := fetchresponse.FetchResponse{Messages: *messages}
	data, _ := json.Marshal(response)

	endpoint := common.BaseEndpoint(r)
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

func react(request reactionrequestitem.ReactionRequestItem, r *events.APIGatewayWebsocketProxyRequest) (events.APIGatewayProxyResponse, error) {
	conn, _ := db.GetConnection(r.RequestContext.ConnectionID)
	msg, err := db.ReactToMessage(request.MID, reaction.Reaction{
		Type: request.Reaction,
		By:   conn.UID,
	})
	if err != nil {
		log.Println(err)
		return events.APIGatewayProxyResponse{StatusCode: 500}, nil
	}
	data, _ := json.Marshal(msg)
	endpoint := common.BaseEndpoint(r)
	config := cfg
	config.BaseEndpoint = &endpoint
	client := apigatewaymanagementapi.NewFromConfig(config)

	_, err = client.PostToConnection(context.TODO(), &apigatewaymanagementapi.PostToConnectionInput{
		Data:         data,
		ConnectionId: &conn.ID,
	})
	if err != nil {
		log.Println(err)
		return events.APIGatewayProxyResponse{StatusCode: 500}, nil
	}
	return events.APIGatewayProxyResponse{StatusCode: 200}, nil
}

func main() {
	lambda.Start(handler)
}
