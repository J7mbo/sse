package logger

import "github.com/google/uuid"

type Logger interface {
	/* Usual stuff that can be ignored. */
	Debug(message string, fields ...map[string]interface{})
	/* Stuff that should be seen as we go about. */
	Info(message string, fields ...map[string]interface{})
	/* Something bad happened but the application can continue. */
	Warning(message string, fields ...map[string]interface{})
	/* Everything is majorly fucked. */
	Error(message string, fields ...map[string]interface{})
	/* Every logger uses a UUID - and it can be updated depending on the request coming in, for example. */
	UpdateUuid(uuid uuid.UUID)
}
