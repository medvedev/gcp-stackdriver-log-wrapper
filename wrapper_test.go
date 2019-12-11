package logwrapper

import (
	"context"
	"testing"

	"cloud.google.com/go/functions/metadata"
	"gotest.tools/assert"
)

type contextKey string

var (
	ctx = metadata.NewContext(
		context.Background(),
		&metadata.Metadata{
			EventID: "eventId",
		})
)

func TestMain(m *testing.M) {
	projectID, region, functionName = "projectId", "region", "functionName"
	m.Run()
	projectID, region, functionName = "", "", ""
}

func TestCreateWrapper(t *testing.T) {
	wl, err := FromContext(ctx, "testLogger")
	if err != nil {
		t.Error(err)
	}
	assert.Assert(t, wl != nil)
}

func TestNoMetadata(t *testing.T) {
	wl, err := FromContext(context.Background(), "testLogger")
	if err == nil || wl != nil {
		t.Fatal("Expected error and nil wrapper")
	}
	assert.Equal(t, err.Error(), "Failed to get executionId")
}

func TestNoEnvironment(t *testing.T) {
	projectID = ""
	wl, err := FromContext(ctx, "testLogger")
	if err == nil || wl != nil {
		t.Fatal("Expected error and nil wrapper")
	}
	assert.Equal(t, err.Error(), "GCP_PROJECT not set")
}
