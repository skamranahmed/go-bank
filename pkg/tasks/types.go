package tasks

const (
	DefaultQueue  string = "default"
	PriorityQueue string = "priority"
)

type Payload[T any] struct {
	CorrelationID string `json:"correlation_id"`
	Data          T      `json:"data"`
}
