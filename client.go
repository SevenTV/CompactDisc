package compactdisc

import (
	"bytes"
	"encoding/json"
	"net/http"

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

func (req Request[T]) ToRaw() Request[json.RawMessage] {
	b, _ := json.Marshal(req.Data)

	return Request[json.RawMessage]{
		Operation: req.Operation,
		Data:      b,
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
	Revoke bool               `json:"revoke"`
}

type Instance interface {
	SyncUser(userID primitive.ObjectID) (*http.Response, error)
	RevokeUser(userID primitive.ObjectID) (*http.Response, error)
}

type cdInst struct {
	addr string

	httpClient *http.Client
}

func New(addr string) Instance {
	cl := http.Client{}

	return &cdInst{
		addr:       addr,
		httpClient: &cl,
	}
}

func (inst *cdInst) request(r Request[json.RawMessage]) (*http.Response, error) {
	b, err := json.Marshal(&r)
	if err != nil {
		return nil, err
	}

	return inst.httpClient.Post(inst.addr, "application/json", bytes.NewBuffer(b))
}

func (inst *cdInst) SyncUser(userID primitive.ObjectID) (*http.Response, error) {
	return inst.request(Request[RequestPayloadSyncUser]{
		Operation: OperationNameSyncUser,
		Data: RequestPayloadSyncUser{
			UserID: userID,
		},
	}.ToRaw())
}

func (inst *cdInst) RevokeUser(userID primitive.ObjectID) (*http.Response, error) {
	return inst.request(Request[RequestPayloadSyncUser]{
		Operation: OperationNameSyncUser,
		Data: RequestPayloadSyncUser{
			UserID: userID,
			Revoke: true,
		},
	}.ToRaw())
}
