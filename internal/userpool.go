package internal

import (
	"context"
	"sync"
)

type UserInfo struct {
	Topic  string
	UserID string
	Nonce  string
}

type UserPool struct {
	connectedUsers map[*UserInfo]context.CancelFunc
	mtx            sync.Mutex
}

func NewUserPool() *UserPool {
	return &UserPool{connectedUsers: make(map[*UserInfo]context.CancelFunc)}
}

func (s *UserPool) AddConnectedUser(ui *UserInfo) {
	s.RemoveConnectedUser(ui)

	_, cancelFunc := context.WithCancel(context.Background())

	s.mtx.Lock()
	s.connectedUsers[ui] = cancelFunc
	s.mtx.Unlock()
}

func (s *UserPool) FindConnectedUser(userID string) *UserInfo {
	if userID == "" {
		return nil
	}

	s.mtx.Lock()
	defer s.mtx.Unlock()

	for cu := range s.connectedUsers {
		if cu.UserID == userID {
			return cu
		}
	}

	return nil
}

func (s *UserPool) CountConnectedUsers() int {
	return len(s.connectedUsers)
}

func (s *UserPool) RemoveConnectedUser(ui *UserInfo) {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	if cancelFunc := s.findConnectedUserCancelFunc(ui.UserID); cancelFunc != nil {
		go cancelFunc()
		delete(s.connectedUsers, ui)
	}
}

func (s *UserPool) findConnectedUserCancelFunc(userID string) context.CancelFunc {
	for cu, cf := range s.connectedUsers {
		if cu.UserID == userID {
			return cf
		}
	}

	return nil
}
