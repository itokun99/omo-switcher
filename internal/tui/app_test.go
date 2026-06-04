package tui

import (
	"errors"
	"strings"
	"testing"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/itokun99/omo-switch/internal/domain"
	"github.com/itokun99/omo-switch/internal/infrastructure"
)

func TestNewApp(t *testing.T) {
	app := NewApp(nil)
	if app == nil {
		t.Fatal("NewApp returned nil")
	}
	if !app.loading {
		t.Error("expected loading to be true")
	}
}

func TestAppInit(t *testing.T) {
	app := NewApp(nil)
	cmd := app.Init()
	if cmd == nil {
		t.Error("Init returned nil cmd")
	}
}

func TestDefaultKeyMap(t *testing.T) {
	km := DefaultKeyMap()
	keys := []key.Binding{
		km.Up, km.Down, km.Top, km.Bottom,
		km.Enter, km.Search, km.Help, km.Quit,
		km.Validate, km.Backup, km.Diff, km.Info, km.Reload,
	}
	for _, k := range keys {
		if k.Help().Key == "" {
			t.Error("key binding missing help key")
		}
		if k.Help().Desc == "" {
			t.Error("key binding missing help description")
		}
	}
}

func TestDefaultStyles(t *testing.T) {
	s := DefaultStyles()
	_ = s.Title.Render("test")
	_ = s.Status.Render("test")
	_ = s.Border.Render("test")
}

// Compile-time check that App implements tea.Model.
var _ tea.Model = (*App)(nil)

// Compile-time check that mockConfigService implements configService.
var _ configService = (*mockConfigService)(nil)

// mockConfigService is a test double for configService.
type mockConfigService struct {
	groups       []domain.Group
	active       string
	switchErr    error
	switchCalls  []string
	listErr      error
	activeErr    error
	reloadErr    error
	reloadCalls  int
	showContent  string
	showErr      error
}

func (m *mockConfigService) ListConfigs() ([]domain.Group, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	return m.groups, nil
}

func (m *mockConfigService) ReloadConfigs() ([]domain.Group, error) {
	m.reloadCalls++
	if m.reloadErr != nil {
		return nil, m.reloadErr
	}
	return m.groups, nil
}

func (m *mockConfigService) GetActiveConfig() (string, error) {
	if m.activeErr != nil {
		return "", m.activeErr
	}
	return m.active, nil
}

func (m *mockConfigService) SwitchConfig(alias string) error {
	m.switchCalls = append(m.switchCalls, alias)
	if m.switchErr != nil {
		return m.switchErr
	}
	m.active = alias
	return nil
}

func (m *mockConfigService) ShowConfig(alias string) (string, error) {
	if m.showErr != nil {
		return "", m.showErr
	}
	return m.showContent, nil
}

func (m *mockConfigService) ValidateConfig(alias string) (bool, string, error) {
	for _, g := range m.groups {
		for _, cfg := range g.Configs {
			if cfg.Alias == alias {
				return cfg.IsValid, cfg.Error, nil
			}
		}
	}
	return false, "not found", nil
}

func (m *mockConfigService) ListBackups() ([]infrastructure.BackupInfo, error) {
	return nil, nil
}

func (m *mockConfigService) RestoreBackup(timestamp string) error {
	return nil
}

func (m *mockConfigService) GetConfigPath(alias string) (string, error) {
	return "/test/config/dir/omo-" + alias + ".json", nil
}

func newMockService() *mockConfigService {
	return &mockConfigService{
		groups: []domain.Group{
			{
				Name: "Mono",
				Configs: []domain.Config{
					{Alias: "config-a", FileName: "omo-config-a.json", IsValid: true},
					{Alias: "config-b", FileName: "omo-config-b.json", IsValid: true},
				},
			},
		},
		active: "config-a",
	}
}

func setupLoadedApp(t *testing.T, svc configService) *App {
	t.Helper()
	app := NewApp(svc)
	app.width = 80
	app.height = 24

	groups, err := svc.ListConfigs()
	if err != nil {
		t.Fatalf("list configs: %v", err)
	}
	active, err := svc.GetActiveConfig()
	if err != nil {
		t.Fatalf("get active: %v", err)
	}

	msg := configsLoadedMsg{groups: groups, active: active}
	model, _ := app.Update(msg)
	return model.(*App)
}

func TestSwitchSuccess(t *testing.T) {
	svc := newMockService()
	app := setupLoadedApp(t, svc)

	// Move cursor to "config-b" (second config, skip group header).
	app.list.MoveDown()

	// Press Enter.
	model, cmd := app.Update(tea.KeyMsg{Type: tea.KeyEnter})
	app = model.(*App)
	if cmd == nil {
		t.Fatal("expected switch command after Enter")
	}

	// Execute the switch command to get the result message.
	msg := cmd()
	switchMsg, ok := msg.(switchCompleteMsg)
	if !ok {
		t.Fatalf("expected switchCompleteMsg, got %T", msg)
	}

	if switchMsg.err != nil {
		t.Fatalf("unexpected switch error: %v", switchMsg.err)
	}

	// Feed switchCompleteMsg to Update.
	model, batchCmd := app.Update(switchMsg)
	app = model.(*App)

	if batchCmd == nil {
		t.Fatal("expected batch command after switch success")
	}

	if !strings.Contains(app.status.Message(), "Switched to: config-b") {
		t.Errorf("expected success message, got: %q", app.status.Message())
	}

	if app.active != "config-b" {
		t.Errorf("expected active = config-b, got: %q", app.active)
	}

	if len(svc.switchCalls) != 1 || svc.switchCalls[0] != "config-b" {
		t.Errorf("expected SwitchConfig called with config-b, got: %v", svc.switchCalls)
	}
}

func TestSwitchError(t *testing.T) {
	svc := newMockService()
	svc.switchErr = errors.New("backup failed")
	app := setupLoadedApp(t, svc)

	// Move cursor to "config-a" (first config).
	// Press Enter.
	model, cmd := app.Update(tea.KeyMsg{Type: tea.KeyEnter})
	app = model.(*App)
	if cmd == nil {
		t.Fatal("expected switch command after Enter")
	}

	msg := cmd()
	switchMsg, ok := msg.(switchCompleteMsg)
	if !ok {
		t.Fatalf("expected switchCompleteMsg, got %T", msg)
	}

	if switchMsg.err == nil {
		t.Fatal("expected switch error")
	}

	// Feed error switchCompleteMsg to Update.
	model, tickCmd := app.Update(switchMsg)
	app = model.(*App)

	if tickCmd == nil {
		t.Fatal("expected tick command for error auto-clear")
	}

	if !strings.Contains(app.status.Message(), "Error:") {
		t.Errorf("expected error message, got: %q", app.status.Message())
	}

	if !strings.Contains(app.status.Message(), "backup failed") {
		t.Errorf("expected error message to contain 'backup failed', got: %q", app.status.Message())
	}

	// Active should NOT change on error.
	if app.active != "config-a" {
		t.Errorf("expected active to remain config-a after error, got: %q", app.active)
	}
}

func TestSwitchUpdatesActive(t *testing.T) {
	svc := newMockService()
	app := setupLoadedApp(t, svc)

	if app.active != "config-a" {
		t.Fatalf("expected initial active = config-a, got: %q", app.active)
	}

	// Move cursor to "config-b".
	app.list.MoveDown()

	// Press Enter -> start switch.
	model, cmd := app.Update(tea.KeyMsg{Type: tea.KeyEnter})
	app = model.(*App)
	if cmd == nil {
		t.Fatal("expected switch command")
	}

	// Switch returns success.
	msg := cmd()
	switchMsg := msg.(switchCompleteMsg)

	// Before Update, active should still be old value.
	// Feed switchCompleteMsg.
	model, _ = app.Update(switchMsg)
	app = model.(*App)

	// Active should have updated immediately in the handler.
	if app.active != "config-b" {
		t.Errorf("expected active = config-b after switch, got: %q", app.active)
	}

	// Verify mock service state.
	if svc.active != "config-b" {
		t.Errorf("expected mock active = config-b, got: %q", svc.active)
	}
}

func TestEmptyState(t *testing.T) {
	app := NewApp(nil)
	app.width = 80
	app.height = 24

	msg := configsLoadedMsg{
		groups: []domain.Group{},
		active: "",
	}
	model, _ := app.Update(msg)
	updated := model.(*App)

	if !updated.isEmpty {
		t.Error("expected isEmpty to be true when no groups")
	}

	view := updated.View()
	if !strings.Contains(view, "No configs found") {
		t.Errorf("expected empty state message in view, got: %s", view)
	}
}

func TestEmptyStateAllGroupsEmpty(t *testing.T) {
	app := NewApp(nil)
	app.width = 80
	app.height = 24

	msg := configsLoadedMsg{
		groups: []domain.Group{
			{Name: "Mono", Configs: []domain.Config{}},
			{Name: "Optimized", Configs: []domain.Config{}},
		},
		active: "",
	}
	model, _ := app.Update(msg)
	updated := model.(*App)

	if !updated.isEmpty {
		t.Error("expected isEmpty to be true when all groups have no configs")
	}
}

func TestTerminalTooSmall(t *testing.T) {
	app := NewApp(nil)
	app.width = 40
	app.height = 10

	view := app.View()
	if !strings.Contains(view, "Terminal too small") {
		t.Errorf("expected terminal too small message, got: %s", view)
	}
	if !strings.Contains(view, "80x24") {
		t.Errorf("expected minimum dimensions in message, got: %s", view)
	}
}

func TestTerminalSizeBoundary(t *testing.T) {
	app := NewApp(nil)
	app.loading = false
	app.width = MinWidth
	app.height = MinHeight

	view := app.View()
	if strings.Contains(view, "Terminal too small") {
		t.Error("should not show resize message at exact minimum dimensions")
	}
}

func TestErrorMessage(t *testing.T) {
	app := NewApp(nil)
	app.width = 80
	app.height = 24

	testErr := errors.New("config file not found")
	msg := errMsg{err: testErr}
	model, cmd := app.Update(msg)
	updated := model.(*App)

	if updated.err == nil {
		t.Error("expected error to be set")
	}
	if updated.status.Message() == "" {
		t.Error("expected status message to contain error")
	}
	if cmd == nil {
		t.Error("expected auto-clear tick command")
	}
}

func TestErrorClear(t *testing.T) {
	app := NewApp(nil)
	app.width = 80
	app.height = 24

	app.err = errors.New("test error")
	app.status.SetMessage("Error: test error")

	model, _ := app.Update(clearErrMsg{})
	updated := model.(*App)

	if updated.status.Message() != "" {
		t.Errorf("expected status message cleared, got: %q", updated.status.Message())
	}
}

func TestErrorRecovery(t *testing.T) {
	app := NewApp(nil)
	app.width = 80
	app.height = 24
	app.err = errors.New("load failed")
	app.loading = false

	reloadKey := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}}
	model, cmd := app.Update(reloadKey)
	updated := model.(*App)

	if !updated.loading {
		t.Error("expected loading to be true after r key")
	}
	if updated.err != nil {
		t.Error("expected error to be cleared after r key")
	}
	if cmd == nil {
		t.Error("expected reload command")
	}
}

func TestNavigationDisabledWhenEmpty(t *testing.T) {
	app := NewApp(nil)
	app.width = 80
	app.height = 24
	app.isEmpty = true

	downKey := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}}
	model, _ := app.Update(downKey)
	updated := model.(*App)

	if !updated.isEmpty {
		t.Error("expected isEmpty to remain true")
	}
}

func TestQuitAlwaysWorks(t *testing.T) {
	app := NewApp(nil)
	app.width = 80
	app.height = 24
	app.isEmpty = true

	_, cmd := app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	if cmd == nil {
		t.Fatal("expected quit command")
	}
}

func TestReload(t *testing.T) {
	svc := newMockService()
	app := setupLoadedApp(t, svc)

	// Press r key.
	model, cmd := app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
	app = model.(*App)

	if !app.loading {
		t.Error("expected loading to be true after r key")
	}
	if app.err != nil {
		t.Error("expected err to be cleared after r key")
	}
	if cmd == nil {
		t.Fatal("expected reload command after r key")
	}

	// Execute the command to get the result message.
	msg := cmd()
	loadedMsg, ok := msg.(configsLoadedMsg)
	if !ok {
		t.Fatalf("expected configsLoadedMsg, got %T", msg)
	}
	if !loadedMsg.reloaded {
		t.Error("expected reloaded to be true in configsLoadedMsg")
	}

	// Feed configsLoadedMsg to Update.
	model, tickCmd := app.Update(loadedMsg)
	app = model.(*App)

	if app.loading {
		t.Error("expected loading to be false after configs loaded")
	}
	if len(app.groups) != 1 {
		t.Errorf("expected 1 group, got %d", len(app.groups))
	}
	if app.active != "config-a" {
		t.Errorf("expected active to be config-a, got %q", app.active)
	}
	// Should have auto-clear tick after reload message.
	if tickCmd == nil {
		t.Error("expected tick command for auto-clear after reload")
	}

	// Verify ReloadConfigs was called on the service.
	if svc.reloadCalls != 1 {
		t.Errorf("expected ReloadConfigs called 1 time, got %d", svc.reloadCalls)
	}
}

func TestReloadShowsMessage(t *testing.T) {
	svc := newMockService()
	app := setupLoadedApp(t, svc)

	// Press r key and execute the reload command.
	_, cmd := app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
	msg := cmd()
	app.Update(msg)

	if !strings.Contains(app.status.Message(), "Configs reloaded") {
		t.Errorf("expected 'Configs reloaded' status message, got: %q", app.status.Message())
	}
}
