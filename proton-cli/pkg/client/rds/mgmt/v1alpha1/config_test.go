package v1alpha1

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClientFor(t *testing.T) {
	c, err := ClientFor(&Config{Host: "localhost:8888", Username: "hello", Password: "world"})
	assert.NotNil(t, c)
	assert.NoError(t, err)
}

func TestClientForConfigAndClient(t *testing.T) {
	c, err := ClientForConfigAndHTTPClient(&Config{Host: "localhost:8888", Username: "hello", Password: "world"}, http.DefaultClient)
	assert.NotNil(t, c)
	assert.NoError(t, err)
}
