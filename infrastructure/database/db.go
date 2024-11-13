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
	"github.com/beglaryh/messenger/domain/connection"
	"github.com/beglaryh/messenger/domain/message"
	"github.com/beglaryh/messenger/domain/room"
	"github.com/beglaryh/messenger/infrastructure/item"
	"github.com/beglaryh/messenger/infrastructure/item/connectionheaderitem"
	"github.com/beglaryh/messenger/infrastructure/item/connectionroomitem"
	"github.com/beglaryh/messenger/infrastructure/item/member"
	"github.com/beglaryh/messenger/infrastructure/item/messageroomitem"
	"github.com/beglaryh/messenger/infrastructure/item/messageuseritem"
	roomitem "github.com/beglaryh/messenger/infrastructure/item/room"
)

type DB struct {
	client *dynamodb.Client
}

type primaryKey struct {
	pk string
	sk string
}

func (pk primaryKey) toAVMap() map[string]types.AttributeValue {
	return map[string]types.AttributeValue{
		"pk": &types.AttributeValueMemberS{Value: pk.pk},
		"sk": &types.AttributeValueMemberS{Value: pk.sk},
	}
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

	return db.batchInsertion(&items)
}

func (db *DB) GetRoomHeader(id string) optional.Optional[room.Room] {
	item, err := getItem[roomitem.RoomItem](db.client, primaryKey{id, roomitem.SK})
	if err != nil {
		log.Println(err)
		return optional.Empty[room.Room]()
	}

	return optional.With(item.To())
}

func (db *DB) AppendRoomToConnections(cs []connection.Connection, room room.Room) error {
	items := make([]map[string]types.AttributeValue, len(cs))
	for i, c := range cs {
		item := connectionroomitem.New(c.ID, room.Id.String(), c.UID)
		av, _ := attributevalue.MarshalMap(item)
		items[i] = av
	}
	return db.batchInsertion(&items)
}

func (db *DB) SaveMessage(message message.Message) error {
	roomItem := messageroomitem.From(message)
	userItems := messageuseritem.From(message)

	items := make([]map[string]types.AttributeValue, len(*userItems)+1)

	item, _ := attributevalue.MarshalMap(roomItem)
	items[0] = item
	for i, e := range *userItems {
		item, _ := attributevalue.MarshalMap(e)
		items[i+1] = item
	}

	return db.batchInsertion(&items)
}

func (db *DB) SaveConnection(c connection.Connection) error {
	roomsIds, err := db.fetchRoomsByUserId(c.UID)
	if err != nil {
		return err
	}
	ch := connectionheaderitem.From(c)

	items := make([]map[string]types.AttributeValue, len(roomsIds)+1)

	chItem, _ := attributevalue.MarshalMap(ch)
	items[0] = chItem

	for i, roomId := range roomsIds {
		connectionItem := connectionroomitem.New(c.ID, roomId, c.UID)
		item, _ := attributevalue.MarshalMap(connectionItem)
		items[i+1] = item
	}

	return db.batchInsertion(&items)
}

func (db *DB) GetMessagesAfter(mid, uid string) (*[]message.Message, error) {
	if len(mid) == 0 {
		mid = "0"
	}
	keyEx := expression.Key("gsi1pk").Equal(expression.Value(uid)).
		And(expression.Key("gsi1sk").GreaterThan(expression.Value("M#" + mid)))
	exp, _ := expression.NewBuilder().WithKeyCondition(keyEx).
		WithFilter(expression.Equal(expression.Name("entityType"), expression.Value(item.UserMessage))).
		Build()

	queryInput := dynamodb.QueryInput{
		TableName:                 &table,
		IndexName:                 &gsi1,
		ExpressionAttributeNames:  exp.Names(),
		ExpressionAttributeValues: exp.Values(),
		KeyConditionExpression:    exp.KeyCondition(),
		FilterExpression:          exp.Filter(),
	}
	paginator := dynamodb.NewQueryPaginator(db.client, &queryInput)
	messages := []message.Message{}
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(context.TODO())
		if err != nil {
			log.Println(err)
			return nil, errors.DefaultInternalError
		}
		items := make([]messageuseritem.MessageUserItem, page.Count)
		attributevalue.UnmarshalListOfMaps(page.Items, &items)
		msgs := messageuseritem.To(&items)
		messages = append(messages, msgs...)
	}
	return &messages, nil
}

func (db *DB) GetConnection(cid string) (connection.Connection, error) {
	item, err := getItem[connectionheaderitem.ConnectionHeaderItem](db.client, primaryKey{cid, connectionheaderitem.CH_SK})
	if err != nil {
		return connection.Connection{}, err
	}
	return item.To(), nil
}

func (db *DB) GetConnectionsByUserId(uid string) ([]connection.Connection, error) {
	keyEx := expression.Key("gsi1pk").Equal(expression.Value(uid)).
		And(expression.Key("gsi1sk").BeginsWith(connectionheaderitem.CH_GSI1SK_PREXIF))
	exp, _ := expression.NewBuilder().WithKeyCondition(keyEx).Build()
	input := dynamodb.QueryInput{
		TableName:                 &table,
		IndexName:                 &gsi1,
		ExpressionAttributeNames:  exp.Names(),
		ExpressionAttributeValues: exp.Values(),
		KeyConditionExpression:    exp.KeyCondition(),
	}

	response, err := db.client.Query(context.TODO(), &input)
	if err != nil {
		log.Println(err)
		return make([]connection.Connection, 0), errors.DefaultInternalError
	}

	connections := make([]connection.Connection, len(response.Items))
	for i, item := range response.Items {
		var ch connectionheaderitem.ConnectionHeaderItem
		err := attributevalue.UnmarshalMap(item, &ch)
		if err != nil {
			log.Println(err)
			return make([]connection.Connection, 0), errors.DefaultInternalError
		}
		connections[i] = ch.To()
	}
	return connections, nil
}

func (db *DB) GetConnectionsByRoom(rid string) ([]connectionroomitem.ConnectionRoomItem, error) {
	keyEx := expression.
		Key("gsi1pk").Equal(expression.Value(rid)).
		And(expression.KeyBeginsWith(expression.Key("gsi1sk"), "U#"))

	exp, _ := expression.NewBuilder().WithKeyCondition(keyEx).Build()

	queryInput := dynamodb.QueryInput{
		TableName:                 &table,
		IndexName:                 &gsi1,
		ExpressionAttributeNames:  exp.Names(),
		ExpressionAttributeValues: exp.Values(),
		KeyConditionExpression:    exp.KeyCondition(),
	}
	queryPaginator := dynamodb.NewQueryPaginator(db.client, &queryInput)
	connections := arraylist.New[connectionroomitem.ConnectionRoomItem]()
	for queryPaginator.HasMorePages() {
		response, err := queryPaginator.NextPage(context.TODO())
		if err != nil {
			log.Println(err)
			return connections.ToArray(), errors.DefaultInternalError
		}
		var items []connectionroomitem.ConnectionRoomItem
		attributevalue.UnmarshalListOfMaps(response.Items, &items)
		for _, m := range items {
			connections.Add(m)
		}
	}

	return connections.ToArray(), nil
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

func (db *DB) batchInsertion(items *[]map[string]types.AttributeValue) error {
	total := len(*items)

	for i := 0; i < total; {
		j := i + BATCH_SIZE
		if j > total {
			j = total
		}

		batchItems := (*items)[i:j]
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

func getItem[E any](client *dynamodb.Client, primaryKey primaryKey) (E, error) {
	o, err := client.GetItem(context.TODO(), &dynamodb.GetItemInput{
		TableName: &table,
		Key:       primaryKey.toAVMap(),
	})

	var item E
	if err != nil {
		log.Println(err)
		return item, errors.DefaultInternalError
	}
	if err := attributevalue.UnmarshalMap(o.Item, &item); err != nil {
		log.Print(err)
		return item, errors.DefaultInternalError
	}

	return item, nil
}
