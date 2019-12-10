package logwrapper

import (
	"context"
	"testing"

	"gotest.tools/assert"
)

var ctx = context.Background()

func TestMain(m *testing.M) {
	projectID, region, functionName = "projectId", "region", "functionName"
	m.Run()
	projectID, region, functionName = "", "", ""
}

func TestCreateWrapper(t *testing.T) {
	wl, err := FromEnvironment(ctx, "testLogger")
	if err != nil {
		t.Error(err)
	}
	assert.Assert(t, wl != nil)
}

func TestNoEnvironment(t *testing.T) {
	projectID = ""
	wl, err := FromEnvironment(ctx, "testLogger")
	if err == nil || wl != nil {
		t.Fatal("Expected error and nil wrapper")
	}
	assert.Equal(t, err.Error(), "GCP_PROJECT not set")
}
