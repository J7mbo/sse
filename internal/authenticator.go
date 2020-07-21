package internal

//go:generate mockery --all -case underscore -output ../test/mock
//go:generate gofmt -r "userExistenceChecker -> UserExistenceChecker" -w ../test/mock/user_existence_checker.go
//go:generate gofmt -r "connectedUserAdderRemover -> ConnectedUserAdderRemover" -w ../test/mock/connected_user_adder_remover.go
//go:generate gofmt -r "errorDebugLogger -> ErrorDebugLogger" -w ../test/mock/error_debug_logger.go
//go:generate gofmt -r "userInfoFromHttpRequestParser -> UserInfoFromHttpRequestParser" -w ../test/mock/user_info_from_http_request_parser.go
import (
	"fmt"
	"net/http"

	_ "github.com/lib/pq"
)

type userAndNonceChecker interface {
	UserExistsInDB(ui *UserInfo) (bool, error)
}

type connectedUserAdderRemover interface {
	AddConnectedUser(ui *UserInfo)
	RemoveConnectedUser(ui *UserInfo)
}

type errorDebugLogger interface {
	Error(message string, fields ...map[string]interface{})
	Debug(message string, fields ...map[string]interface{})
}

type userInfoFromHttpRequestParser interface {
	Parse(r *http.Request) (ui *UserInfo, err error)
}

type Authenticator struct {
	parser               userInfoFromHttpRequestParser
	userExistenceChecker userAndNonceChecker
	userAdderRemover     connectedUserAdderRemover
	logger               errorDebugLogger
}

func NewAuthenticator(
	uec userAndNonceChecker, psr userInfoFromHttpRequestParser, uar connectedUserAdderRemover, lgr errorDebugLogger,
) *Authenticator {
	return &Authenticator{userExistenceChecker: uec, parser: psr, userAdderRemover: uar, logger: lgr}
}

func (a *Authenticator) AuthenticateMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userInfo, err := a.parser.Parse(r)
		if err != nil {
			a.logger.Error(fmt.Sprintf("error parsing user info from request: %s", err.Error()))
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		userExists, err := a.userExistenceChecker.UserExistsInDB(userInfo)
		if err != nil {
			a.logger.Error(fmt.Sprintf("error looking up user in userAndNonceChecker: %s", err.Error()))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if !userExists {
			a.logger.Error(fmt.Sprintf("service called with nonexistent user: %s", userInfo.UserID))
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// User Connected.
		a.logger.Debug(fmt.Sprintf("CONNECT: userID: %s, topic: %s", userInfo.UserID, userInfo.Topic))
		a.userAdderRemover.AddConnectedUser(userInfo)

		next.ServeHTTP(w, r)

		// User Disconnected.
		a.logger.Debug(fmt.Sprintf("DISCONNECT: userID: %s, topic: %s", userInfo.UserID, userInfo.Topic))
		a.userAdderRemover.RemoveConnectedUser(userInfo)
	})
}
