package types

import (
	"encoding/json"
)

// Subscriber ...
type Subscriber struct {
	Mail string `json:"mail"`
	Link string `json:"link"`
}

// SubscriberResponse ...
type SubscriberResponse struct {
	Mail string
	Link string
}

var responseOK map[string]string = map[string]string{
	"status_code":    "200",
	"status_message": "OK",
}

// ResponseOK ...
var ResponseOK, _ = json.Marshal(responseOK)
