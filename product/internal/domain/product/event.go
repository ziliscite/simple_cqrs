package product

type Message interface {
	*Product | ID
}

type EventAction string

const (
	Create EventAction = "create"
	Update EventAction = "update"
	Delete EventAction = "delete"
)

type Event[T Message] struct {
	Action  EventAction
	Message T
}

func NewEvent[T Message](action EventAction, message T) Event[T] {
	return Event[T]{
		Action:  action,
		Message: message,
	}
}
