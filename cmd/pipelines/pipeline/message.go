package pipeline

type Message[T any] struct {
	ID      int64
	Payload T
}
