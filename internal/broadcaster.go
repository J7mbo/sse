package internal

import (
	"fmt"

	"github.com/mroth/sseserver"
)

type connectedUserFinder interface {
	FindConnectedUser(userID string) *UserInfo
}

type Broadcaster struct {
	sseServer *sseserver.Server
	userPool  connectedUserFinder
}

func NewBroadcaster(sseServer *sseserver.Server, userPool connectedUserFinder) *Broadcaster {
	return &Broadcaster{sseServer: sseServer, userPool: userPool}
}

func (b *Broadcaster) Broadcast(userID string, data []byte) {
	if userID == "" {
		return
	}

	user := b.userPool.FindConnectedUser(userID)
	if user == nil {
		return
	}

	b.send(user, data)
}

func (b *Broadcaster) send(ui *UserInfo, data []byte) {
	b.sseServer.Broadcast <- sseserver.SSEMessage{Data: data, Namespace: b.getUrlForBroadcasting(ui)}
}

func (b *Broadcaster) getUrlForBroadcasting(ui *UserInfo) string {
	return fmt.Sprintf("/%s/%s/%s", ui.Topic, ui.UserID, ui.Nonce)
}
