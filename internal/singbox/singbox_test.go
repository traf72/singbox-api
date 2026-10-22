package singbox

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseServiceStatus(t *testing.T) {
	status, err := parseServiceStatus([]byte(`Description=sing-box service
LoadState=loaded
ActiveState=active
SubState=running
UnitFileState=enabled
Result=success
MainPID=1234
ExecMainStatus=0
`))

	require.NoError(t, err)
	assert.Equal(t, "sing-box service", status.Description)
	assert.Equal(t, "loaded", status.LoadState)
	assert.Equal(t, "active", status.ActiveState)
	assert.Equal(t, "running", status.SubState)
	assert.Equal(t, "enabled", status.UnitFileState)
	assert.Equal(t, "success", status.Result)
	assert.Equal(t, 1234, status.MainPID)
	assert.Zero(t, status.ExecMainStatus)
	assert.True(t, status.Active)
}

func TestStatusWhenInteractionDisabled(t *testing.T) {
	t.Setenv("DISABLE_SINGBOX_INTERACTION", "true")

	status, err := Status()

	require.Nil(t, err)
	assert.True(t, status.InteractionDisabled)
	assert.False(t, status.Active)
}
