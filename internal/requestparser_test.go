package internal_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/suite"

	"sse/internal"
)

type RequestParserTestSuite struct {
	suite.Suite
}

func (s *RequestParserTestSuite) TestValidRequest_IsParsedSuccessfully() {
	// Given.
	req := mux.SetURLVars(
		httptest.NewRequest(http.MethodPost, "http://example.com/subscribe/{topic}/{userID}/{nonce}", nil),
		map[string]string{"topic": "topic", "userID": "userID", "nonce": "nonce"},
	)

	// When.
	ui, err := internal.NewRequestParser().Parse(req)

	// Then.
	s.Nil(err)
	s.Equal(&internal.UserInfo{Topic: "topic", UserID: "userID", Nonce: "nonce"}, ui)
}

func (s *RequestParserTestSuite) TestInvalidRequest_ItNotParsed() {
	// Given.
	req := mux.SetURLVars(
		httptest.NewRequest(http.MethodPost, "http://example.com/subscribe/{topic}/{userID}/{nonce}", nil),
		map[string]string{"topic": "topic", "userID": "userID", "nonce_is_missing": "no_nonce_here"},
	)

	// When.
	ui, err := internal.NewRequestParser().Parse(req)

	// Then.
	s.Nil(ui)
	s.Error(err)
}

func TestRequestParserTestSuite(t *testing.T) {
	suite.Run(t, new(RequestParserTestSuite))
}
