package logger

import (
	"github.com/google/uuid"
)

/* RequiredLogFields are fields to be always present in the logs. They are merged in with LogFields when logged. */
type RequiredLogFields struct {
	Env           string
	CorrelationId uuid.UUID
	Index         string
}

/* toMap converts the struct for logging; the key and value pairs in the struct tag become the KvPs in the map. */
func (lf *RequiredLogFields) toMap() map[string]interface{} {
	return map[string]interface{}{
		"env":            lf.Env,
		"correlation_id": lf.CorrelationId.String(),
		"index":          lf.Index,
	}
}
