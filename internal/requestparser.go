package internal

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

type RequestParser struct{}

func NewRequestParser() *RequestParser {
	return &RequestParser{}
}

func (*RequestParser) Parse(r *http.Request) (ui *UserInfo, err error) {
	vars := mux.Vars(r)

	topic, foundTopic := vars["topic"]
	userID, foundUserID := vars["userID"]
	nonce, foundNonce := vars["nonce"]

	if !foundTopic || !foundUserID || !foundNonce {
		return nil, fmt.Errorf("invalid url, expected /subscribe/<topic>/<userid>/<nonce>, got: %s", r.RequestURI)
	}

	return &UserInfo{topic, userID, nonce}, nil
}
