package database

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/beglaryh/gocommon/collection/list/arraylist"
	"github.com/beglaryh/gocommon/errors"
	"github.com/beglaryh/gocommon/optional"
	"github.com/beglaryh/gocommon/stream"
	"github.com/beglaryh/messenger/domain/message"
	"github.com/beglaryh/messenger/domain/room"
	"github.com/beglaryh/messenger/infrastructure/item/connection"
	"github.com/beglaryh/messenger/infrastructure/item/member"
	messageitem "github.com/beglaryh/messenger/infrastructure/item/message"
	roomitem "github.com/beglaryh/messenger/infrastructure/item/room"
	"github.com/google/uuid"
)

type DB struct {
	client *dynamodb.Client
}

type primaryKey struct {
	pk string
	sk string
}

var (
	table = "messenger"
	gsi1  = "gsi1"
)

func New(config aws.Config) *DB {
	client := dynamodb.NewFromConfig(config)
	if client == nil {
		log.Fatal("client is null")
	}

	return &DB{client}
}

const BATCH_SIZE = 25

var errorCreatingRoom = errors.NewInternal("error creating room")

func (db *DB) SaveRoom(room room.Room) error {
	roomItem, _ := attributevalue.MarshalMap(roomitem.From(room))
	memberItems := member.From(room)
	items := make([]map[string]types.AttributeValue, len(memberItems)+1)

	items[0] = roomItem
	for i, memberItem := range memberItems {
		item, _ := attributevalue.MarshalMap(memberItem)
		items[i+1] = item
	}

	return db.batchInsertion(items)
}

func (db *DB) GetRoomHeader(id uuid.UUID) optional.Optional[room.Room] {
	key := map[string]types.AttributeValue{
		"pk": &types.AttributeValueMemberS{Value: id},
		"sk": &types.AttributeValueMemberS{Value: roomitem.SK},
	}
	o, err := db.client.GetItem(context.TODO(), &dynamodb.GetItemInput{
		TableName: &table,
		Key:       key,
	})
	if err != nil {
		log.Println(err)
		return optional.Empty[room.Room]()
	}
	var item roomitem.RoomItem
	if err := attributevalue.UnmarshalMap(o.Item, &item); err != nil {
		log.Print(err)
		return optional.Empty[room.Room]()
	}

	return optional.With(item.To())
}

func (db *DB) AppendRoomToConnection(cid string, room room.Room) error {
	citem := connection.New(cid, room.Id.String(), room.CreatedBy)
	item, _ := attributevalue.MarshalMap(citem)
	return db.putItem(item)
}

func (db *DB) SaveMessage(message message.Message) error {
	item, _ := attributevalue.MarshalMap(messageitem.From(message))
	return db.putItem(item)
}

func (db *DB) SaveConnection(connectionId, userId string) error {
	roomsIds, err := db.fetchRoomsByUserId(userId)
	if err != nil {
		return err
	}
	items := make([]map[string]types.AttributeValue, len(roomsIds)+1)

	items[0] = map[string]types.AttributeValue{
		"pk":         &types.AttributeValueMemberS{Value: connectionId},
		"sk":         &types.AttributeValueMemberS{Value: userId},
		"entityType": &types.AttributeValueMemberS{Value: "ConnectionHeader"},
	}

	for i, roomId := range roomsIds {
		connectionItem := connection.New(connectionId, roomId, userId)
		item, _ := attributevalue.MarshalMap(connectionItem)
		items[i+1] = item
	}

	return db.batchInsertion(items)
}

func (db *DB) RemoveConnection(cid string) error {
	key := expression.
		Key("pk").
		Equal(expression.Value(cid))
	exp, _ := expression.NewBuilder().WithKeyCondition(key).Build()
	queryInput := dynamodb.QueryInput{
		TableName:                 &table,
		ExpressionAttributeNames:  exp.Names(),
		ExpressionAttributeValues: exp.Values(),
		KeyConditionExpression:    exp.KeyCondition(),
	}

	paginator := dynamodb.NewQueryPaginator(db.client, &queryInput)

	pks := arraylist.New[primaryKey]()
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(context.TODO())
		if err != nil {
			log.Println(err)
			return errors.DefaultInternalError
		}
		members := []member.MemberItem{}
		err = attributevalue.UnmarshalListOfMaps(page.Items, &members)
		if err != nil {
			log.Println(err)
			return errors.DefaultInternalError
		}

		for _, member := range members {
			pks.Add(primaryKey{member.PK, member.SK})
		}
	}

	db.batchDelete(pks.ToArray())
	return nil
}

func (db *DB) putItem(item map[string]types.AttributeValue) error {
	_, err := db.client.PutItem(context.TODO(), &dynamodb.PutItemInput{
		TableName: &table,
		Item:      item,
	})
	if err != nil {
		log.Println(err)
		return errors.NewInternal("error saving item")
	}
	return nil
}

func (db *DB) batchInsertion(items []map[string]types.AttributeValue) error {
	total := len(items)

	for i := 0; i < total; {
		j := i + BATCH_SIZE
		if j > total {
			j = total
		}

		batchItems := items[i:j]
		batch := stream.Map(batchItems, func(item map[string]types.AttributeValue) types.WriteRequest {
			return types.WriteRequest{
				PutRequest: &types.PutRequest{
					Item: item,
				},
			}
		}).Slice()

		_, err := db.client.BatchWriteItem(context.TODO(), &dynamodb.BatchWriteItemInput{
			RequestItems: map[string][]types.WriteRequest{
				table: batch,
			},
		})
		if err != nil {
			log.Println(err)
			return err
		}

		i = j
	}

	return nil
}

func (db *DB) batchDelete(keys []primaryKey) error {
	requests := make([]types.WriteRequest, len(keys))
	for i, e := range keys {
		pk := map[string]types.AttributeValue{
			"pk": &types.AttributeValueMemberS{Value: e.pk},
			"sk": &types.AttributeValueMemberS{Value: e.sk},
		}
		requests[i] = types.WriteRequest{
			DeleteRequest: &types.DeleteRequest{
				Key: pk,
			},
		}
	}
	totalRequests := len(requests)
	for i := 0; i < totalRequests; {
		j := i + BATCH_SIZE
		if j > totalRequests {
			j = len(keys)
		}
		batch := requests[i:j]
		_, err := db.client.BatchWriteItem(context.TODO(), &dynamodb.BatchWriteItemInput{
			RequestItems: map[string][]types.WriteRequest{
				table: batch,
			},
		})
		if err != nil {
			log.Println(err)
			return errors.DefaultInternalError
		}

		i = j
	}
	return nil
}

func (db *DB) fetchRoomsByUserId(userId string) ([]string, error) {
	rooms := arraylist.New[string]()
	keyEx := expression.Key("gsi1pk").Equal(expression.Value(userId))
	exp, _ := expression.NewBuilder().WithKeyCondition(keyEx).Build()

	queryInput := dynamodb.QueryInput{
		TableName:                 &table,
		IndexName:                 &gsi1,
		ExpressionAttributeNames:  exp.Names(),
		ExpressionAttributeValues: exp.Values(),
		KeyConditionExpression:    exp.KeyCondition(),
	}
	queryPaginator := dynamodb.NewQueryPaginator(db.client, &queryInput)
	for queryPaginator.HasMorePages() {
		response, err := queryPaginator.NextPage(context.TODO())
		if err != nil {
			log.Println(err)
			return rooms.ToArray(), errors.DefaultInternalError
		}
		var members []member.MemberItem
		attributevalue.UnmarshalListOfMaps(response.Items, &members)
		for _, m := range members {
			rooms.Add(m.PK)
		}
	}
	return rooms.ToArray(), nil
}
