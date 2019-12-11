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
	executionIDFromRequest := func() (string, error) {
		executionID := r.Header["Function-Execution-Id"]
		if len(executionID) == 0 {
			return "", errors.New("Header 'Function-Execution-Id' is missing")
		}
		return executionID[0], nil
	}
	return foo(r.Context(), loggerName, executionIDFromRequest)
}

// FromContext creates WrappedLogger from GCP Cloud function runtime (context + environment)
func FromContext(ctx context.Context, loggerName string) (*WrappedLogger, error) {
	executionIDFromMetadata := func() (string, error) {
		ctxMetadata, err := metadata.FromContext(ctx)
		if err != nil {
			return "", errors.New("Failed to get metadata")
		}
		return ctxMetadata.EventID, nil
	}
	return foo(ctx, loggerName, executionIDFromMetadata)
}

func foo(ctx context.Context, loggerName string, executionIDSupplier func() (string, error)) (*WrappedLogger, error) {
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
	executionId, err := executionIDSupplier()
	if err != nil {
		return nil, errors.New("Failed to get executionId")
	}
	wrappedLogger.labels = map[string]string{"execution_id": executionId}
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
