package logwrapper

import (
	"context"
	"errors"
	"net/http"
	"os"

	"cloud.google.com/go/logging"

	"cloud.google.com/go/functions/metadata"
	"google.golang.org/genproto/googleapis/api/monitoredres"
)

var (
	projectID    = os.Getenv("GCP_PROJECT")
	functionName = os.Getenv("FUNCTION_NAME")
	region       = os.Getenv("FUNCTION_REGION")
)

// WrappedLogger struct holding stackdriver logger and monitored resource data
type WrappedLogger struct {
	monRes *monitoredres.MonitoredResource
	logger *logging.Logger
	labels map[string]string
}

// FromRequest creates WrappedLogger from GCP Cloud function runtime (request)
func FromRequest(r *http.Request, loggerName string) (*WrappedLogger, error) {
	logger, err := createLogger(r.Context(), loggerName)
	if err == nil {
		executionID := r.Header["Function-Execution-Id"]
		if len(executionID) == 0 {
			err = errors.New("Header 'Function-Execution-Id' is missing")
		} else {
			logger.labels = map[string]string{"execution_id": executionID[0]}
		}
	}
	return logger, err
}

// FromContext creates WrappedLogger from GCP Cloud function runtime (context + environment)
func FromContext(ctx context.Context, loggerName string) (*WrappedLogger, error) {
	logger, err := createLogger(ctx, loggerName)
	if err == nil {
		ctxMetadata, err := metadata.FromContext(ctx)
		if err != nil {
			err = errors.New("Failed to get metadata")
		} else {
			logger.labels = map[string]string{"event_id": ctxMetadata.EventID}
		}
	}
	return logger, err
}

func createLogger(ctx context.Context, loggerName string) (*WrappedLogger, error) {
	var wrappedLogger WrappedLogger
	monres, err := getMonitoredResourceFromEnvironment()
	if err != nil {
		return nil, err
	}
	wrappedLogger.monRes = monres
	client, err := logging.NewClient(ctx, projectID)
	if err != nil {
		return nil, err
	}
	wrappedLogger.logger = client.Logger(loggerName)
	return &wrappedLogger, nil
}

func getMonitoredResourceFromEnvironment() (*monitoredres.MonitoredResource, error) {
	if projectID == "" {
		return nil, errors.New("GCP_PROJECT not set")
	}
	if functionName == "" {
		return nil, errors.New("FUNCTION_NAME not set")
	}
	if functionName == "" {
		return nil, errors.New("FUNCTION_REGION not set")
	}
	return &monitoredres.MonitoredResource{
		Type: "cloud_function",
		Labels: map[string]string{
			"function_name": functionName,
			"project_id":    projectID,
			"region":        region,
		},
	}, nil
}

// Debug (self-explaining)
func (wl *WrappedLogger) Debug(payload interface{}) {
	wl.Log(logging.Debug, payload)
}

// Info (self-explaining)
func (wl *WrappedLogger) Info(payload interface{}) {
	wl.Log(logging.Info, payload)
}

// Error (self-explaining)
func (wl *WrappedLogger) Error(payload interface{}) {
	wl.Log(logging.Error, payload)
}

// Log (self-explaining)
func (wl *WrappedLogger) Log(severity logging.Severity, payload interface{}) {
	wl.logger.Log(logging.Entry{
		Severity: severity,
		Resource: wl.monRes,
		Payload:  payload,
		Labels:   wl.labels,
	})
}

// Flush (self-explaining)
func (wl *WrappedLogger) Flush() {
	wl.logger.Flush()
}
