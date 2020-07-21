package logger

import (
	"io"
	"os"

	"github.com/google/uuid"
	logrusLogger "github.com/sirupsen/logrus"
)

type FileLogger struct {
	logrusLogger   *logrusLogger.Logger
	requiredFields *RequiredLogFields
}

func New(lf io.Writer, lgr *logrusLogger.Logger, rf *RequiredLogFields) *FileLogger {
	lgr.SetFormatter(&logrusLogger.JSONFormatter{
		DisableTimestamp: true,
		FieldMap: logrusLogger.FieldMap{
			logrusLogger.FieldKeyTime:  "@timestamp",
			logrusLogger.FieldKeyLevel: "@level",
			logrusLogger.FieldKeyMsg:   "message",
		},
	})

	lgr.SetOutput(io.MultiWriter(os.Stdout, lf))

	return &FileLogger{
		logrusLogger:   lgr,
		requiredFields: rf,
	}
}

func (l *FileLogger) UpdateUuid(uuid uuid.UUID) {
	l.requiredFields.CorrelationId = uuid
}

func (l *FileLogger) Debug(message string, fields ...map[string]interface{}) {
	l.logrusLogger.WithFields(l.mergeFields(fields)).Debug(message)
}

func (l *FileLogger) Info(message string, fields ...map[string]interface{}) {
	l.logrusLogger.WithFields(l.mergeFields(fields)).Info(message)
}

func (l *FileLogger) Warning(message string, fields ...map[string]interface{}) {
	l.logrusLogger.WithFields(l.mergeFields(fields)).Warning(message)
}

func (l *FileLogger) Error(message string, fields ...map[string]interface{}) {
	l.logrusLogger.WithFields(l.mergeFields(fields)).Error(message)
}

/* mergeFields merges the required fields with the provided fields for logging to elasticsearch. */
func (l *FileLogger) mergeFields(fields []map[string]interface{}) logrusLogger.Fields {
	finalFields := l.requiredFields.toMap()

	if len(fields) > 0 {
		for _, field := range fields {
			finalFields = l.mergeMaps(finalFields, field)
		}
	}

	return finalFields
}

func (l *FileLogger) mergeMaps(maps ...map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})

	for _, m := range maps {
		for k, v := range m {
			result[k] = v
		}
	}

	return result
}
