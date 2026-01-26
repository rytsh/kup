package tui

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/rytsh/kup/internal/config"
	"github.com/rytsh/kup/internal/tui/tabs"
)

// Tab represents a tab in the application
type Tab int

const (
	TabInstall Tab = iota
	TabSettings
)

func (t Tab) String() string {
	switch t {
	case TabInstall:
		return "Install"
	case TabSettings:
		return "Settings"
	default:
		return "Unknown"
	}
}

// KeyMap defines the keybindings for the application
type KeyMap struct {
	NextTab key.Binding
	PrevTab key.Binding
	Quit    key.Binding
	Help    key.Binding
	Select  key.Binding
	Back    key.Binding
	Up      key.Binding
	Down    key.Binding
	Left    key.Binding
	Right   key.Binding
	Confirm key.Binding
	Cancel  key.Binding
}

// DefaultKeyMap returns the default keybindings
func DefaultKeyMap() KeyMap {
	return KeyMap{
		NextTab: key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("tab", "next tab"),
		),
		PrevTab: key.NewBinding(
			key.WithKeys("shift+tab"),
			key.WithHelp("shift+tab", "prev tab"),
		),
		Quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q", "quit"),
		),
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "help"),
		),
		Select: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "select"),
		),
		Back: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "back"),
		),
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("up/k", "up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("down/j", "down"),
		),
		Left: key.NewBinding(
			key.WithKeys("left", "h"),
			key.WithHelp("left/h", "left"),
		),
		Right: key.NewBinding(
			key.WithKeys("right", "l"),
			key.WithHelp("right/l", "right"),
		),
		Confirm: key.NewBinding(
			key.WithKeys("y"),
			key.WithHelp("y", "confirm"),
		),
		Cancel: key.NewBinding(
			key.WithKeys("n"),
			key.WithHelp("n", "cancel"),
		),
	}
}

// Model is the main application model
type Model struct {
	config *config.Config
	styles *Styles
	keyMap KeyMap
	width  int
	height int

	// Current state
	activeTab Tab
	showHelp  bool

	// Tab models
	installTab  tabs.InstallModel
	settingsTab tabs.SettingsModel
}

// New creates a new application model
func New(cfg *config.Config) Model {
	theme := GetTheme(cfg.Theme)
	styles := NewStyles(theme)

	return Model{
		config:      cfg,
		styles:      styles,
		keyMap:      DefaultKeyMap(),
		activeTab:   TabInstall,
		installTab:  tabs.NewInstallModel(cfg, styles),
		settingsTab: tabs.NewSettingsModel(cfg, styles),
	}
}

// Init implements tea.Model
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.installTab.Init(),
		m.settingsTab.Init(),
	)
}

// Update implements tea.Model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.installTab.SetSize(msg.Width, msg.Height-6)
		m.settingsTab.SetSize(msg.Width, msg.Height-6)

	case tea.KeyMsg:
		// Global key handling
		switch {
		case key.Matches(msg, m.keyMap.Quit):
			// Don't quit if we're in the middle of input
			if m.activeTab == TabSettings && m.settingsTab.IsInputFocused() {
				break
			}
			return m, tea.Quit

		case key.Matches(msg, m.keyMap.NextTab):
			if !m.settingsTab.IsInputFocused() {
				m.activeTab = (m.activeTab + 1) % 2
				return m, nil
			}

		case key.Matches(msg, m.keyMap.PrevTab):
			if !m.settingsTab.IsInputFocused() {
				m.activeTab = (m.activeTab + 2 - 1) % 2
				return m, nil
			}

		case key.Matches(msg, m.keyMap.Help):
			m.showHelp = !m.showHelp
			return m, nil
		}

	case tabs.ThemeChangedMsg:
		// Update theme when settings change it
		theme := GetTheme(msg.Theme)
		m.styles = NewStyles(theme)
		m.installTab.UpdateStyles(m.styles)
		m.settingsTab.UpdateStyles(m.styles)
	}

	// Update active tab
	switch m.activeTab {
	case TabInstall:
		newInstall, cmd := m.installTab.Update(msg)
		m.installTab = newInstall.(tabs.InstallModel)
		cmds = append(cmds, cmd)
	case TabSettings:
		newSettings, cmd := m.settingsTab.Update(msg)
		m.settingsTab = newSettings.(tabs.SettingsModel)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

// View implements tea.Model
func (m Model) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	// Build tab bar
	tabBar := m.renderTabBar()

	// Build content
	var content string
	switch m.activeTab {
	case TabInstall:
		content = m.installTab.View()
	case TabSettings:
		content = m.settingsTab.View()
	}

	// Build help
	help := m.renderHelp()

	// Combine everything
	return lipgloss.JoinVertical(
		lipgloss.Left,
		tabBar,
		content,
		help,
	)
}

func (m Model) renderTabBar() string {
	tabs := []Tab{TabInstall, TabSettings}
	var renderedTabs []string

	for _, t := range tabs {
		var style lipgloss.Style
		if t == m.activeTab {
			style = m.styles.TabActive
		} else {
			style = m.styles.TabInactive
		}
		renderedTabs = append(renderedTabs, style.Render(t.String()))
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, renderedTabs...)
}

func (m Model) renderHelp() string {
	if m.showHelp {
		return m.styles.Help.Render(
			m.styles.HelpKey.Render("tab") + m.styles.HelpDesc.Render(" switch tabs  ") +
				m.styles.HelpKey.Render("up/down") + m.styles.HelpDesc.Render(" navigate  ") +
				m.styles.HelpKey.Render("enter") + m.styles.HelpDesc.Render(" select  ") +
				m.styles.HelpKey.Render("esc") + m.styles.HelpDesc.Render(" back  ") +
				m.styles.HelpKey.Render("q") + m.styles.HelpDesc.Render(" quit  ") +
				m.styles.HelpKey.Render("?") + m.styles.HelpDesc.Render(" toggle help"),
		)
	}
	return m.styles.Help.Render(
		m.styles.HelpKey.Render("?") + m.styles.HelpDesc.Render(" help  ") +
			m.styles.HelpKey.Render("q") + m.styles.HelpDesc.Render(" quit"),
	)
}

// Run starts the TUI application
func Run(cfg *config.Config) error {
	p := tea.NewProgram(
		New(cfg),
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	_, err := p.Run()
	return err
}
