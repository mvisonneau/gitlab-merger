package cli

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewApp(t *testing.T) {
	app := NewApp("0.0.0", time.Now())
	assert.Equal(t, "gitlab-merger", app.Name)
	assert.Equal(t, "0.0.0", app.Version)
}
