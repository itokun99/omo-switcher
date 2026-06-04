package tui

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/lipgloss"

	"github.com/itokun99/omo-switch/internal/domain"
	"github.com/itokun99/omo-switch/internal/infrastructure"
	"github.com/itokun99/omo-switch/internal/tui/components"
)

// configService defines the subset of ConfigService methods used by the TUI.
// This interface allows testing with mock implementations.
type configService interface {
	ListConfigs() ([]domain.Group, error)
	GetActiveConfig() (string, error)
	ReloadConfigs() ([]domain.Group, error)
	SwitchConfig(alias string) error
	ShowConfig(alias string) (string, error)
	ValidateConfig(alias string) (bool, string, error)
	ListBackups() ([]infrastructure.BackupInfo, error)
	RestoreBackup(timestamp string) error
	GetConfigPath(alias string) (string, error)
}

// Minimum terminal dimensions.
const (
	MinWidth  = 80
	MinHeight = 24
)

// Error auto-clear duration.
const errorClearDuration = 3 * time.Second

// ViewMode represents the current TUI view.
type ViewMode int

const (
	ViewList ViewMode = iota
	ViewDetail
	ViewSearch
	ViewHelp
	ViewBackup
	ViewDiff
	ViewValidate
	ViewInfo
)

// App is the main TUI model.
type App struct {
	service configService
	keys    KeyMap
	styles  Styles

	// State
	mode    ViewMode
	groups  []domain.Group
	active  string
	cursor  int // current item index (flattened across groups)
	loading bool
	isEmpty bool
	err     error
	width   int
	height  int

	list     components.ListModel
	status   components.StatusModel
	search   components.SearchModel
	detail   components.DetailModel
	help     components.HelpModel
	validate components.ValidateModel
	backup   components.BackupModel
	diff     components.DiffModel
	info     components.InfoModel
}

// NewApp creates a new TUI App.
func NewApp(service configService) *App {
	return &App{
		service:  service,
		keys:     DefaultKeyMap(),
		styles:   DefaultStyles(),
		loading:  true,
		status:   components.NewStatusModel(),
		search:   components.NewSearchModel(),
		detail:   components.NewDetailModel(),
		help:     components.NewHelpModel(),
		validate: components.NewValidateModel(),
		backup:   components.NewBackupModel(),
		diff:     components.NewDiffModel(),
		info:     components.NewInfoModel(),
	}
}

// Init implements tea.Model.
func (a *App) Init() tea.Cmd {
	return tea.Batch(
		loadConfigs(a.service),
		tea.EnterAltScreen,
	)
}

// Update implements tea.Model.
func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		return a, nil

	case tea.KeyMsg:
		return a.handleKey(msg)

	case configsLoadedMsg:
		a.groups = msg.groups
		a.active = msg.active
		a.loading = false
		a.isEmpty = len(msg.groups) == 0 || allGroupsEmpty(msg.groups)

		if msg.reloaded {
			a.status.SetMessage("Configs reloaded")
			return a, tea.Tick(errorClearDuration, func(time.Time) tea.Msg {
				return clearMessageMsg{}
			})
		}

		if a.isEmpty {
			a.status.SetMessage("No configs found — add omo-*.json files")
			return a, nil
		}
		group := domain.GetGroupForAlias(msg.active)
		a.status.SetActive(msg.active, group)
		a.list = components.NewListModel(msg.groups, msg.active, a.height-4)
		return a, nil

	case errMsg:
		a.err = msg.err
		a.loading = false
		a.status.SetMessage(fmt.Sprintf("Error: %v", msg.err))
		return a, tea.Tick(errorClearDuration, func(time.Time) tea.Msg {
			return clearErrMsg{}
		})

	case clearErrMsg:
		a.status.ClearMessage()
		return a, nil

	case switchCompleteMsg:
		if msg.err != nil {
			a.status.SetMessage("Error: " + msg.err.Error())
			return a, tea.Tick(errorClearDuration, func(time.Time) tea.Msg {
				return clearMessageMsg{}
			})
		}
		a.status.SetMessage("Switched to: " + msg.alias)
		a.active = msg.alias
		return a, tea.Batch(
			loadConfigs(a.service),
			clearMessageAfter(errorClearDuration),
		)

	case clearMessageMsg:
		a.status.ClearMessage()
		return a, nil

	case validateCompleteMsg:
		a.validate.Show(msg.results, a.height)
		return a, nil

	case backupsLoadedMsg:
		if msg.err != nil {
			a.status.SetMessage("Error loading backups: " + msg.err.Error())
			return a, tea.Tick(errorClearDuration, func(time.Time) tea.Msg {
				return clearMessageMsg{}
			})
		}
		a.backup.Show(msg.backups, a.height)
		return a, nil

	case restoreCompleteMsg:
		if msg.err != nil {
			a.status.SetMessage("Restore failed: " + msg.err.Error())
			a.backup.CancelConfirm()
			return a, tea.Tick(errorClearDuration, func(time.Time) tea.Msg {
				return clearMessageMsg{}
			})
		}
		a.backup.Hide()
		a.status.SetMessage("Backup restored successfully")
		return a, tea.Batch(
			loadConfigs(a.service),
			clearMessageAfter(errorClearDuration),
		)
	}

	return a, nil
}

// View implements tea.Model.
func (a *App) View() string {
	if a.width < MinWidth || a.height < MinHeight {
		return fmt.Sprintf(
			"Terminal too small (need %dx%d, got %dx%d)\n\nPlease resize your terminal.",
			MinWidth, MinHeight, a.width, a.height,
		)
	}

	if a.loading {
		return "Loading configs..."
	}

	if a.isEmpty {
		return "No configs found in ~/.config/opencode/omo_configs/\n\nAdd omo-*.json files to get started."
	}

	if a.err != nil && a.isEmpty {
		return fmt.Sprintf("Error: %v\n\nPress r to retry, q to quit.", a.err)
	}

	listStyles := components.ListStyles{
		GroupHeader: a.styles.Subtitle,
		Active:      a.styles.Active,
		Inactive:    a.styles.Inactive,
		Cursor:      a.styles.Active,
	}

	var content string
	if a.backup.IsActive() {
		backupStyles := components.BackupStyles{
			Selected: a.styles.Active,
			Normal:   a.styles.Inactive,
			Confirm:  a.styles.ErrorMsg,
		}
		content = a.backup.Render(backupStyles, a.width)
	} else if a.diff.IsActive() {
		diffStyles := components.DiffStyles{
			Header:    a.styles.Title,
			Added:     lipgloss.NewStyle().Foreground(lipgloss.Color("42")),
			Removed:   lipgloss.NewStyle().Foreground(lipgloss.Color("196")),
			Unchanged: lipgloss.NewStyle().Foreground(lipgloss.Color("241")),
			Footer:    a.styles.Help,
			Muted:     a.styles.Help,
		}
		content = a.diff.Render(diffStyles, a.width)
	} else if a.validate.IsActive() {
		validateStyles := components.ValidateStyles{
			Valid:   a.styles.Active,
			Invalid: a.styles.ErrorMsg,
			Error:   a.styles.Help,
		}
		content = a.validate.Render(validateStyles, a.width)
	} else if a.detail.IsActive() {
		a.detail.SetWidth(a.width)
		detailStyles := components.DetailStyles{
			Header: a.styles.Title,
			Footer: a.styles.Help,
			Muted:  lipgloss.NewStyle().Foreground(a.styles.Muted),
		}
		content = a.detail.Render(detailStyles, a.width)
	} else if a.info.IsActive() {
		infoStyles := components.InfoStyles{
			Label: lipgloss.NewStyle().Foreground(a.styles.Active.GetForeground()),
		}
		content = a.info.Render(infoStyles, a.width)
	} else if a.search.IsActive() {
		searchStyles := components.SearchStyles{
			Selected: a.styles.Active,
			Normal:   a.styles.Inactive,
		}
		content = a.search.Render(searchStyles, a.width)
	} else {
		content = a.list.Render(listStyles, a.width)
	}

	a.status.SetWidth(a.width)
	statusBar := a.status.Render(components.StatusStyles{
		Bar: a.styles.Status,
	})

	view := fmt.Sprintf("%s\n\n%s\n%s", a.styles.Title.Render("omo-switch"), content, statusBar)

	if a.help.IsActive() {
		a.help.SetSize(a.width, a.height)
		helpStyles := components.HelpStyles{
			Box: lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				Padding(1, 2).
				BorderForeground(a.styles.Accent),
		}
		helpOverlay := a.help.Render(helpStyles)
		return view + "\n" + helpOverlay
	}

	return view
}

// Messages
type configsLoadedMsg struct {
	groups   []domain.Group
	active   string
	reloaded bool
}

type errMsg struct {
	err error
}

type clearErrMsg struct{}

// switchCompleteMsg is sent after an async SwitchConfig call completes.
type switchCompleteMsg struct {
	alias string
	err   error
}

// clearMessageMsg clears the status bar message.
type clearMessageMsg struct{}

// validateCompleteMsg is sent after async validation completes.
type validateCompleteMsg struct {
	results []components.ValidationResult
}

// backupsLoadedMsg is sent after async ListBackups call completes.
type backupsLoadedMsg struct {
	backups []infrastructure.BackupInfo
	err     error
}

// restoreCompleteMsg is sent after async RestoreBackup call completes.
type restoreCompleteMsg struct {
	err error
}

// Commands
func loadConfigs(service configService) tea.Cmd {
	return func() tea.Msg {
		groups, err := service.ListConfigs()
		if err != nil {
			return errMsg{err}
		}
		active, err := service.GetActiveConfig()
		if err != nil {
			return errMsg{err}
		}
		return configsLoadedMsg{groups: groups, active: active, reloaded: false}
	}
}

func reloadConfigsCmd(service configService) tea.Cmd {
	return func() tea.Msg {
		groups, err := service.ReloadConfigs()
		if err != nil {
			return errMsg{err}
		}
		active, err := service.GetActiveConfig()
		if err != nil {
			return errMsg{err}
		}
		return configsLoadedMsg{groups: groups, active: active, reloaded: true}
	}
}

func allGroupsEmpty(groups []domain.Group) bool {
	for _, g := range groups {
		if len(g.Configs) > 0 {
			return false
		}
	}
	return true
}

func switchConfigCmd(service configService, alias string) tea.Cmd {
	return func() tea.Msg {
		err := service.SwitchConfig(alias)
		return switchCompleteMsg{alias: alias, err: err}
	}
}

func clearMessageAfter(d time.Duration) tea.Cmd {
	return tea.Tick(d, func(time.Time) tea.Msg {
		return clearMessageMsg{}
	})
}

func validateConfigsCmd(service configService, groups []domain.Group) tea.Cmd {
	return func() tea.Msg {
		var results []components.ValidationResult
		for _, g := range groups {
			for _, cfg := range g.Configs {
				isValid, errMsg, err := service.ValidateConfig(cfg.Alias)
				if err != nil {
					results = append(results, components.ValidationResult{
						Alias:   cfg.Alias,
						IsValid: false,
						Error:   err.Error(),
					})
					continue
				}
				results = append(results, components.ValidationResult{
					Alias:   cfg.Alias,
					IsValid: isValid,
					Error:   errMsg,
				})
			}
		}
		return validateCompleteMsg{results: results}
	}
}

func loadBackupsCmd(service configService) tea.Cmd {
	return func() tea.Msg {
		backups, err := service.ListBackups()
		return backupsLoadedMsg{backups: backups, err: err}
	}
}

func restoreBackupCmd(service configService, timestamp string) tea.Cmd {
	return func() tea.Msg {
		err := service.RestoreBackup(timestamp)
		return restoreCompleteMsg{err: err}
	}
}

// Key handling
func (a *App) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if a.help.IsActive() {
		a.help.Hide()
		return a, nil
	}

	if a.validate.IsActive() {
		return a.handleValidateKey(msg)
	}

	if a.search.IsActive() {
		return a.handleSearchKey(msg)
	}

	if a.detail.IsActive() {
		return a.handleDetailKey(msg)
	}

	if a.backup.IsActive() {
		return a.handleBackupKey(msg)
	}

	if a.diff.IsActive() {
		return a.handleDiffKey(msg)
	}

	if a.info.IsActive() {
		return a.handleInfoKey(msg)
	}

	switch {
	case key.Matches(msg, a.keys.Quit):
		return a, tea.Quit
	case key.Matches(msg, a.keys.Reload):
		a.loading = true
		a.err = nil
		return a, reloadConfigsCmd(a.service)
	case key.Matches(msg, a.keys.Enter):
		sel := a.list.Selected()
		if sel != nil {
			return a, switchConfigCmd(a.service, sel.Alias)
		}
		return a, nil
	case a.isEmpty || a.err != nil:
		return a, nil
	case key.Matches(msg, a.keys.Up):
		a.list.MoveUp()
	case key.Matches(msg, a.keys.Down):
		a.list.MoveDown()
	case key.Matches(msg, a.keys.Top):
		a.list.MoveToTop()
	case key.Matches(msg, a.keys.Bottom):
		a.list.MoveToBottom()
	case key.Matches(msg, a.keys.Help):
		a.help.Toggle()
	case key.Matches(msg, a.keys.Search):
		a.search.Activate(a.list.Items())
	case key.Matches(msg, a.keys.Detail):
		a.showDetail()
	case key.Matches(msg, a.keys.Validate):
		return a, validateConfigsCmd(a.service, a.groups)
	case key.Matches(msg, a.keys.Backup):
		a.loading = true
		return a, loadBackupsCmd(a.service)
	case key.Matches(msg, a.keys.Diff):
		a.showDiff()
	case key.Matches(msg, a.keys.Info):
		a.showInfo()
	}
	return a, nil
}

func (a *App) handleDetailKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, a.keys.Quit):
		a.detail.Hide()
	case key.Matches(msg, a.keys.Up):
		a.detail.ScrollUp()
	case key.Matches(msg, a.keys.Down):
		a.detail.ScrollDown()
	case key.Matches(msg, a.keys.Top):
		a.detail.ScrollToTop()
	case key.Matches(msg, a.keys.Bottom):
		a.detail.ScrollToBottom()
	default:
		if msg.String() == "esc" {
			a.detail.Hide()
		}
	}
	return a, nil
}

func (a *App) showDetail() {
	sel := a.list.Selected()
	if sel == nil {
		return
	}
	content, err := a.service.ShowConfig(sel.Alias)
	if err != nil {
		a.err = err
		return
	}
	a.detail.Show(sel.Alias, sel.FileName, content, a.height)
}

func (a *App) handleSearchKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		a.search.Deactivate()
		return a, nil
	case "enter":
		sel := a.search.Selected()
		if sel != nil {
			a.list.SetCursorToItem(sel.Alias)
		}
		a.search.Deactivate()
		return a, nil
	}

	a.search.Update(msg)
	return a, nil
}

func (a *App) handleValidateKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		a.validate.Hide()
	case "up", "k":
		a.validate.MoveUp()
	case "down", "j":
		a.validate.MoveDown()
	}
	return a, nil
}

func (a *App) handleBackupKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if a.backup.IsConfirming() {
		switch msg.String() {
		case "y":
			ts := a.backup.Selected()
			if ts != "" {
				a.backup.CancelConfirm()
				return a, restoreBackupCmd(a.service, ts)
			}
		case "n", "esc":
			a.backup.CancelConfirm()
		}
		return a, nil
	}

	switch msg.String() {
	case "esc":
		a.backup.Hide()
	case "up", "k":
		a.backup.MoveUp()
	case "down", "j":
		a.backup.MoveDown()
	case "enter":
		a.backup.StartConfirm()
	}
	return a, nil
}

func (a *App) handleDiffKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		a.diff.Hide()
	case "up", "k":
		a.diff.ScrollUp()
	case "down", "j":
		a.diff.ScrollDown()
	case "g", "home":
		a.diff.ScrollToTop()
	case "G", "end":
		a.diff.ScrollToBottom()
	}
	return a, nil
}

func (a *App) showDiff() {
	sel := a.list.Selected()
	if sel == nil {
		return
	}

	activeContent, err := a.service.ShowConfig(a.active)
	if err != nil {
		a.err = err
		return
	}

	selectedContent, err := a.service.ShowConfig(sel.Alias)
	if err != nil {
		a.err = err
		return
	}

	a.diff.Show(a.active, sel.Alias, activeContent, selectedContent, a.height)
}

func (a *App) handleInfoKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		a.info.Hide()
	}
	return a, nil
}

func (a *App) showInfo() {
	sel := a.list.Selected()
	if sel == nil {
		return
	}

	filePath, err := a.service.GetConfigPath(sel.Alias)
	if err != nil {
		a.err = err
		return
	}

	info := components.GetConfigInfo(sel.Alias, filePath)
	info.IsValid = sel.IsValid
	a.info.Show(info)
}
