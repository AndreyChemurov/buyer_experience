package types

import (
	"encoding/json"
)

// Subscriber - данные, которые передаются в запросе
type Subscriber struct {
	Mail string `json:"mail"`
	Link string `json:"link"`
}

// SubscriberResponse - данные, которые используются из запроса
type SubscriberResponse struct {
	Mail string
	Link string
}

var responseOK map[string]string = map[string]string{
	"status_code":    "200",
	"status_message": "OK",
}

// ResponseOK - json 200'го статуса
var ResponseOK, _ = json.Marshal(responseOK)
