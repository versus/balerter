package manager

import (
	"fmt"
	"github.com/balerter/balerter/internal/alert/alert"
	"github.com/balerter/balerter/internal/alert/message"
	"github.com/balerter/balerter/internal/alert/provider/notify"
	"github.com/balerter/balerter/internal/alert/provider/slack"
	"github.com/balerter/balerter/internal/alert/provider/syslog"
	"github.com/balerter/balerter/internal/alert/provider/telegram"
	"github.com/balerter/balerter/internal/config"
	coreStorage "github.com/balerter/balerter/internal/core_storage"
	"github.com/balerter/balerter/internal/script/script"
	lua "github.com/yuin/gopher-lua"
	"go.uber.org/zap"
)

type alertChannel interface {
	Name() string
	Send(*message.Message) error
}

type Manager struct {
	logger   *zap.Logger
	channels map[string]alertChannel

	engine coreStorage.CoreStorage
}

func New(engine coreStorage.CoreStorage, logger *zap.Logger) *Manager {
	m := &Manager{
		logger:   logger,
		engine:   engine,
		channels: make(map[string]alertChannel),
	}

	return m
}

func (m *Manager) Init(cfg config.Channels) error {
	for _, configWebHook := range cfg.Slack {
		module, err := slack.New(configWebHook, m.logger)
		if err != nil {
			return fmt.Errorf("error init slack channel %s, %w", configWebHook.Name, err)
		}

		m.channels[module.Name()] = module
	}

	for _, cfg := range cfg.Telegram {
		module, err := telegram.New(cfg, m.logger)
		if err != nil {
			return fmt.Errorf("error init telegram channel %s, %w", cfg.Name, err)
		}

		m.channels[module.Name()] = module
	}

	for _, cfg := range cfg.Syslog {
		module, err := syslog.New(cfg, m.logger)
		if err != nil {
			return fmt.Errorf("error init syslog channel %s, %w", cfg.Name, err)
		}

		m.channels[module.Name()] = module
	}

	for _, cfg := range cfg.Notify {
		module, err := notify.New(cfg, m.logger)
		if err != nil {
			return fmt.Errorf("error init syslog channel %s, %w", cfg.Name, err)
		}

		m.channels[module.Name()] = module
	}

	return nil
}

func (m *Manager) Name() string {
	return "alert"
}

func (m *Manager) Stop() error {
	return nil
}

func (m *Manager) GetLoader(script *script.Script) lua.LGFunction {
	return func() lua.LGFunction {
		return func(L *lua.LState) int {
			var exports = map[string]lua.LGFunction{
				"warn":    m.luaCall(script, alert.LevelWarn),
				"warning": m.luaCall(script, alert.LevelWarn),

				"error": m.luaCall(script, alert.LevelError),
				"on":    m.luaCall(script, alert.LevelError),
				"fail":  m.luaCall(script, alert.LevelError),

				"success": m.luaCall(script, alert.LevelSuccess),
				"off":     m.luaCall(script, alert.LevelSuccess),
				"ok":      m.luaCall(script, alert.LevelSuccess),

				"get": m.get(script),
			}

			mod := L.SetFuncs(L.NewTable(), exports)

			L.Push(mod)
			return 1
		}
	}()
}
