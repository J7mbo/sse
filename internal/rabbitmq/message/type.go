package message

//
// Type is used first on de-serialised messages from the queue to figure out which message handler is required to handle
// the message.
//
type Type struct {
	MessageType string `json:"type"`
}
