package handler

import (
	"encoding/json"
	"fmt"

	"github.com/streadway/amqp"

	"sse/internal/logger"
	"sse/internal/rabbitmq/message"
)

type broadcaster interface {
	Broadcast(userID string, data []byte)
}

type FoundContract struct {
	broadcaster broadcaster
	logger      logger.Logger
}

func NewFoundContract(broadcaster broadcaster, lgr logger.Logger) *FoundContract {
	return &FoundContract{broadcaster: broadcaster, logger: lgr}
}

func (f *FoundContract) Handle(msg *amqp.Delivery) {
	var foundContract message.FoundContract

	err := json.Unmarshal(msg.Body, &foundContract)
	if err != nil {
		f.logger.Warning(fmt.Sprintf("unable to decode json for 'foundcontract' message, error: %s", err.Error()))
		return
	}

	// Broadcast to the relevant user, if they're connected... If not, who cares?
	f.logger.Debug(fmt.Sprintf("broadcasting 'foundcontract' message to user: %s", foundContract.UserID))
	f.broadcaster.Broadcast(foundContract.UserID, msg.Body)
}

func (*FoundContract) HandlesMessageType() string {
	return "foundcontract"
}
