package internal_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"sse/internal"
	mocks "sse/test/mock"
)

type AuthenticatorTestSuite struct {
	suite.Suite

	mockParser           *mocks.UserInfoFromHttpRequestParser
	mockUserChecker      *mocks.UserExistenceChecker
	mockUserAdderRemover *mocks.ConnectedUserAdderRemover
	mockErrDebugLogger   *mocks.ErrorDebugLogger
}

func (s *AuthenticatorTestSuite) SetupTest() {
	s.mockParser = new(mocks.UserInfoFromHttpRequestParser)
	s.mockUserChecker = new(mocks.UserExistenceChecker)
	s.mockUserAdderRemover = new(mocks.ConnectedUserAdderRemover)
	s.mockErrDebugLogger = new(mocks.ErrorDebugLogger)
}

func (s *AuthenticatorTestSuite) TestValidRequest_AddsConnectedUserToPool() {
	// Given.
	userInfo := &internal.UserInfo{Topic: "topic", UserID: "userid", Nonce: "nonce"}
	a := internal.NewAuthenticator(s.mockUserChecker, s.mockParser, s.mockUserAdderRemover, s.mockErrDebugLogger)
	s.mockParser.
		On("Parse", mock.Anything).
		Return(userInfo, nil).
		Once()
	s.mockUserChecker.
		On("UserExistsInDB", userInfo).
		Return(true, nil).
		Once()
	s.mockErrDebugLogger.
		On("Debug", mock.Anything, mock.Anything).
		Twice()
	s.mockUserAdderRemover.
		On("AddConnectedUser", userInfo).
		Once()
	s.mockUserAdderRemover.
		On("RemoveConnectedUser", userInfo).
		Once()

	var nextHandlerCalled bool
	nextHandler := http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		nextHandlerCalled = true
	})

	// When.
	req := httptest.NewRequest(http.MethodPost, "http://example.com", nil)
	res := httptest.NewRecorder()

	handlerToTest := a.AuthenticateMiddleware(nextHandler)
	handlerToTest.ServeHTTP(res, req)

	// Then.
	const (
		expectedNextHandlerCalled = true
	)

	s.mockParser.AssertExpectations(s.T())
	s.mockUserChecker.AssertExpectations(s.T())
	s.mockUserAdderRemover.AssertExpectations(s.T())
	s.mockErrDebugLogger.AssertExpectations(s.T())
	s.mockErrDebugLogger.AssertNotCalled(s.T(), "Error")
	s.Equal(expectedNextHandlerCalled, nextHandlerCalled)
}

// TestInvalidRequest_WithoutAllParameters_GivesBadRequest at first glance works very well, but unnecessarily checks
// things such as calling Parse() with a specific argument type. These tests should be testing the API, not the inside.
// I would favour just checking that the input given gives us the expected output, as this allows us to change our
// implementation without having to rewrite the working tests.
func (s *AuthenticatorTestSuite) TestInvalidRequest_WithoutAllParameters_GivesBadRequest() {
	// Given.
	a := internal.NewAuthenticator(s.mockUserChecker, s.mockParser, s.mockUserAdderRemover, s.mockErrDebugLogger)
	s.mockParser.On("Parse", mock.Anything).Return(nil, errors.New("failure")).Once()
	s.mockErrDebugLogger.On("Error", mock.Anything, mock.Anything).Once()

	var nextHandlerCalled bool
	nextHandler := http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		nextHandlerCalled = true
	})

	// When.
	req := httptest.NewRequest(http.MethodPost, "http://example.com", nil)
	res := httptest.NewRecorder()

	handlerToTest := a.AuthenticateMiddleware(nextHandler)
	handlerToTest.ServeHTTP(res, req)

	// Then.
	const (
		expectedStatusCode        = 400
		expectedNextHandlerCalled = false
	)

	// Should be called.
	s.mockParser.AssertExpectations(s.T())
	s.mockErrDebugLogger.AssertExpectations(s.T())
	s.mockParser.AssertCalled(s.T(), "Parse", mock.Anything)
	s.mockErrDebugLogger.AssertCalled(s.T(), "Error", mock.Anything, mock.Anything)

	// Should not be called.
	s.mockUserChecker.AssertNotCalled(s.T(), "UserExistsInDB")
	s.mockUserAdderRemover.AssertNotCalled(s.T(), "AddConnectedUser")

	// Should give us this response.
	s.Equal(expectedStatusCode, res.Code)
	s.Equal(expectedNextHandlerCalled, nextHandlerCalled)
}

func (s *AuthenticatorTestSuite) TestValidRequest_WithDBError_GivesInternalServerError() {
	// Given.
	userInfo := &internal.UserInfo{Topic: "topic", UserID: "userid", Nonce: "nonce"}
	a := internal.NewAuthenticator(s.mockUserChecker, s.mockParser, s.mockUserAdderRemover, s.mockErrDebugLogger)
	s.mockParser.
		On("Parse", mock.Anything).
		Return(userInfo, nil).
		Once()
	s.mockUserChecker.
		On("UserExistsInDB", userInfo).
		Return(false, errors.New("some internal db error")).
		Once()
	s.mockErrDebugLogger.
		On("Error", mock.Anything, mock.Anything).
		Once()

	var nextHandlerCalled bool
	nextHandler := http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		nextHandlerCalled = true
	})

	// When.
	req := httptest.NewRequest(http.MethodPost, "http://example.com", nil)
	res := httptest.NewRecorder()

	handlerToTest := a.AuthenticateMiddleware(nextHandler)
	handlerToTest.ServeHTTP(res, req)

	// Then.
	const (
		expectedStatusCode        = 500
		expectedNextHandlerCalled = false
	)

	s.mockParser.AssertExpectations(s.T())
	s.mockErrDebugLogger.AssertExpectations(s.T())
	s.mockUserChecker.AssertExpectations(s.T())
	s.Equal(expectedStatusCode, res.Code)
	s.Equal(expectedNextHandlerCalled, nextHandlerCalled)
}

func (s *AuthenticatorTestSuite) TestValidRequest_WhereUserDoesNotExist_GivesBadRequest() {
	// Given.
	userInfo := &internal.UserInfo{Topic: "topic", UserID: "userid", Nonce: "nonce"}
	a := internal.NewAuthenticator(s.mockUserChecker, s.mockParser, s.mockUserAdderRemover, s.mockErrDebugLogger)
	s.mockParser.
		On("Parse", mock.Anything).
		Return(userInfo, nil).
		Once()
	s.mockUserChecker.
		On("UserExistsInDB", userInfo).
		Return(false, nil).
		Once()
	s.mockErrDebugLogger.
		On("Error", mock.Anything, mock.Anything).
		Once()

	var nextHandlerCalled bool
	nextHandler := http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		nextHandlerCalled = true
	})

	// When.
	req := httptest.NewRequest(http.MethodPost, "http://example.com", nil)
	res := httptest.NewRecorder()

	handlerToTest := a.AuthenticateMiddleware(nextHandler)
	handlerToTest.ServeHTTP(res, req)

	// Then.
	const (
		expectedStatusCode        = 400
		expectedNextHandlerCalled = false
	)

	s.mockParser.AssertExpectations(s.T())
	s.mockErrDebugLogger.AssertExpectations(s.T())
	s.mockUserChecker.AssertExpectations(s.T())
	s.Equal(expectedStatusCode, res.Code)
	s.Equal(expectedNextHandlerCalled, nextHandlerCalled)
}

func TestAuthenticatorTestSuite(t *testing.T) {
	suite.Run(t, new(AuthenticatorTestSuite))
}
