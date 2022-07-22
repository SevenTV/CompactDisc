package client

import (
	"encoding/json"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Request[T RequestPayload] struct {
	Operation string `json:"op"`
	Data      T      `json:"data"`
}

func ConvertRequest[T RequestPayload](req Request[json.RawMessage]) Request[T] {
	var data T
	_ = json.Unmarshal(req.Data, &data)

	return Request[T]{
		Operation: req.Operation,
		Data:      data,
	}
}

type OperationName string

const (
	OperationNameSyncUser = "SYNC_USER"
)

type RequestPayload interface {
	json.RawMessage | RequestPayloadSyncUser
}

type RequestPayloadSyncUser struct {
	UserID primitive.ObjectID `json:"user_id"`
}
