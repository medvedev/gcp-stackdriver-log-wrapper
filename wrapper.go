package logwrapper

import (
	"context"
	"errors"
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

// FromEnvironment creates WrappedLogger from GCP Cloud function runtime (environment)
func FromEnvironment(ctx context.Context, loggerName string) (*WrappedLogger, error) {

	if projectID == "" {
		return nil, errors.New("GCP_PROJECT not set")
	}
	client, err := logging.NewClient(ctx, projectID)
	if err != nil {
		return nil, err
	}
	var wrappedLogger WrappedLogger
	wrappedLogger.logger = client.Logger(loggerName)

	if functionName == "" {
		return nil, errors.New("FUNCTION_NAME not set")
	}

	if functionName == "" {
		return nil, errors.New("FUNCTION_REGION not set")
	}
	wrappedLogger.monRes = &monitoredres.MonitoredResource{
		Type: "cloud_function",
		Labels: map[string]string{
			"function_name": functionName,
			"project_id":    projectID,
			"region":        region,
		},
	}
	ctxMetadata, err := metadata.FromContext(ctx)
	if err != nil {
		return nil, errors.New("Failed to get metadata")
	}
	wrappedLogger.labels = map[string]string{"execution_id": ctxMetadata.EventID}
	return &wrappedLogger, nil
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
	})
}

// Flush (self-explaining)
func (wl *WrappedLogger) Flush() {
	wl.logger.Flush()
}
