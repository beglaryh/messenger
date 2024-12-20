# Description

This is a simple messenger application utilizing API Gateway Websockets and Lambda Functions.

Current features supported:

- Create Room with n-members
- Send messages
- All connected devices will receive messages live
- Fetch unread messages
- React to messages

Features Pending:

- Undo reactions
- Send media

# Example Payloads

## Create Room

Post Message

```json
{
  "action" : "request",
  "message" : {
  "action" : "create",
    "message" : {
      "name" : "MyRoom",
      "members" : ["u1", "u2"]
    }
  }
}
```

Response
Note: This is a temporary response. Will be enhanced to be more specific for client.

```json
{
  "id" : "b4adcc8e-45e6-4bf9-917a-2724fb98d82b"
}
```

## Send Message

```json
{
  "action" : "request",
  "message" : {
    "action" : "send",
    "message" : {
      "roomId" : "b4adcc8e-45e6-4bf9-917a-2724fb98d82b"
      "message" : "Hello, World!",
    }
  }
}
```

Response

```json
{
  "id" : "e36faf3c-3887-401b-91ec-8b4e634ad57d",
  "roomId" : "b4adcc8e-45e6-4bf9-917a-2724fb98d82b",
  "message" : "Hello, World!",
  "sentBy" : "u1",
  "isEdited" : false,
  "createdOn" : "2024-01-01T08:01:00Z",
}
```

## Fetch Unread messages

This is subject to change. For now it is a simple model where the message will be
the last message that has been consumed by the client. If the client has no messages, then the `message`
field can be left blank. Current response is not paginated and will return all unread messages.

Future model will support pagination and basic filters.

Request

```json
{
  "action" : "request",
  "message" : {
    "action" : "fetch",
    "message" : "ec6b729d-2964-46c3-b8a9-a1f1a5a1cb25"
  }
}
```

Response

```json
{
  "messages" : [
    {
      "id" : "ceb14139-682a-4c16-b502-d9f1e452d4fa",
      "roomId" : "f8f9f7a1-85ef-4937-9911-7df5b950d8b9",
      "message" : "Hello, World!",
      "sentBy" : "u1",
      "sentOn" : "2024-04-01T08:01:00Z",
      "isEdited" : false
    }
  ]
}
```

## Edit Message

Request

```json
{
  "action" : "request",
  "message" : {
    "action" : "edit",
    "message" : {
      "id" : "dc8be21d-60fe-4414-ba01-dfd43764bb05",
      "message" : "Typo!"
    }
  }
}
```

## React

Request

```json
{
  "action" : "request",
  "message" : {
    "action" : "react",
    "message" : {
      "id" : "7e0dc40d-b699-460f-9985-908182d82fd8",
      "reaction" : "Heart"
    }
  }
}
```
