package tasks

import "encoding/json"

const (
	DefaultQueue  string = "default"
	PriorityQueue string = "priority"
)

type Payload[T any] struct {
	CorrelationID string `json:"correlation_id"`
	Data          T      `json:"data"`
}

func ExtractPayload[T any](taskPayloadInBytes []byte) (Payload[T], error) {
	var payload Payload[T]
	err := json.Unmarshal(taskPayloadInBytes, &payload)
	if err != nil {
		return Payload[T]{}, err
	}
	return payload, nil
}
