package wshub

import (
	"encoding/json"
)

// Channel is identifier for the channel to use.
type Channel string

// MessageContainer holds the Channel to route the Payload to.
type MessageContainer struct {
	Channel Channel         `json:"channel"`
	Payload json.RawMessage `json:"payload"`
}
