package manager

import (
	"github.com/balerter/balerter/internal/config"
	"github.com/balerter/balerter/internal/script/script"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	lua "github.com/yuin/gopher-lua"
	"go.uber.org/zap"
	"testing"
)

func TestManager_Init(t *testing.T) {
	m := New(nil, zap.NewNop())

	cfg := config.Channels{
		Slack: []config.ChannelSlack{
			{
				Name:    "slack1",
				Token:   "token",
				Channel: "channel",
			},
		},
	}

	err := m.Init(cfg)
	require.NoError(t, err)
	require.Equal(t, 1, len(m.channels))

	c, ok := m.channels["slack1"]
	require.True(t, ok)
	assert.Equal(t, "slack1", c.Name())
}

func TestManager_Loader(t *testing.T) {
	m := New(nil, zap.NewNop())

	L := lua.NewState()

	f := m.GetLoader(&script.Script{})
	c := f(L)
	assert.Equal(t, 1, c)

	v := L.Get(1).(*lua.LTable)

	assert.IsType(t, &lua.LNilType{}, v.RawGet(lua.LString("wrong-name")))

	assert.IsType(t, &lua.LFunction{}, v.RawGet(lua.LString("warn")))
	assert.IsType(t, &lua.LFunction{}, v.RawGet(lua.LString("on")))
	assert.IsType(t, &lua.LFunction{}, v.RawGet(lua.LString("error")))
	assert.IsType(t, &lua.LFunction{}, v.RawGet(lua.LString("fail")))
	assert.IsType(t, &lua.LFunction{}, v.RawGet(lua.LString("success")))
	assert.IsType(t, &lua.LFunction{}, v.RawGet(lua.LString("off")))
	assert.IsType(t, &lua.LFunction{}, v.RawGet(lua.LString("ok")))
}

func TestManager_Name(t *testing.T) {
	m := &Manager{}

	assert.Equal(t, "alert", m.Name())
}

func TestManager_Stop(t *testing.T) {
	m := &Manager{}

	assert.NoError(t, m.Stop())
}
