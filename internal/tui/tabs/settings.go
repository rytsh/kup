package tabs

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/rytsh/kup/internal/config"
)

// Styles interface for accessing styles from parent
type Styles interface {
	GetTheme() Theme
	GetInputLabel() lipgloss.Style
	GetInputField() lipgloss.Style
	GetInputFocused() lipgloss.Style
	GetListItem() lipgloss.Style
	GetListItemSelected() lipgloss.Style
	GetListItemDesc() lipgloss.Style
	GetTitle() lipgloss.Style
	GetStatusSuccess() lipgloss.Style
	GetStatusError() lipgloss.Style
	GetContent() lipgloss.Style
	GetButton() lipgloss.Style
	GetButtonFocused() lipgloss.Style
}

// Theme interface
type Theme interface{}

// ThemeChangedMsg is sent when the theme is changed
type ThemeChangedMsg struct {
	Theme string
}

// SettingField represents a configurable setting
type SettingField int

const (
	FieldBinPath SettingField = iota
	FieldArchitecture
	FieldShowExplanation
	FieldTheme
	FieldTimeout
	FieldProxyURL
	FieldCount
)

func (f SettingField) String() string {
	switch f {
	case FieldBinPath:
		return "Binary Path"
	case FieldArchitecture:
		return "Architecture"
	case FieldShowExplanation:
		return "Show Command Explanation"
	case FieldTheme:
		return "Theme"
	case FieldTimeout:
		return "Download Timeout (seconds)"
	case FieldProxyURL:
		return "Proxy URL"
	default:
		return "Unknown"
	}
}

func (f SettingField) Description() string {
	switch f {
	case FieldBinPath:
		return "Directory where binaries will be downloaded"
	case FieldArchitecture:
		return "Target architecture for downloads (auto, amd64, arm64)"
	case FieldShowExplanation:
		return "Show command explanation before execution"
	case FieldTheme:
		return "UI color theme (default, dark, light)"
	case FieldTimeout:
		return "Timeout for download operations in seconds"
	case FieldProxyURL:
		return "HTTP proxy URL for downloads (leave empty for direct)"
	default:
		return ""
	}
}

// SettingsModel is the model for the settings tab
type SettingsModel struct {
	config *config.Config
	styles interface{}
	width  int
	height int

	// UI state
	focusedField SettingField
	editing      bool
	statusMsg    string
	statusIsErr  bool

	// Text inputs for editable fields
	binPathInput  textinput.Model
	timeoutInput  textinput.Model
	proxyURLInput textinput.Model

	// Select options
	architectureOptions []string
	architectureIndex   int
	themeOptions        []string
	themeIndex          int
}

// NewSettingsModel creates a new settings model
func NewSettingsModel(cfg *config.Config, styles interface{}) SettingsModel {
	// Binary path input
	binPathInput := textinput.New()
	binPathInput.Placeholder = "~/bin"
	binPathInput.SetValue(cfg.BinPath)
	binPathInput.Width = 40

	// Timeout input
	timeoutInput := textinput.New()
	timeoutInput.Placeholder = "120"
	timeoutInput.SetValue(cfg.Timeout.String())
	timeoutInput.Width = 10

	// Proxy URL input
	proxyURLInput := textinput.New()
	proxyURLInput.Placeholder = "http://proxy:8080"
	proxyURLInput.SetValue(cfg.ProxyURL)
	proxyURLInput.Width = 40

	// Architecture options
	archOptions := []string{"auto", "amd64", "arm64"}
	archIndex := 0
	for i, opt := range archOptions {
		if opt == cfg.Architecture {
			archIndex = i
			break
		}
	}

	// Theme options
	themeOptions := []string{"default", "dark", "light"}
	themeIndex := 0
	for i, opt := range themeOptions {
		if opt == cfg.Theme {
			themeIndex = i
			break
		}
	}

	return SettingsModel{
		config:              cfg,
		styles:              styles,
		focusedField:        FieldBinPath,
		binPathInput:        binPathInput,
		timeoutInput:        timeoutInput,
		proxyURLInput:       proxyURLInput,
		architectureOptions: archOptions,
		architectureIndex:   archIndex,
		themeOptions:        themeOptions,
		themeIndex:          themeIndex,
	}
}

// Init implements tea.Model
func (m SettingsModel) Init() tea.Cmd {
	return nil
}

// SetSize sets the size of the settings view
func (m *SettingsModel) SetSize(width, height int) {
	m.width = width
	m.height = height
}

// IsInputFocused returns true if a text input is focused
func (m *SettingsModel) IsInputFocused() bool {
	return m.editing
}

// UpdateStyles updates the styles
func (m *SettingsModel) UpdateStyles(styles interface{}) {
	m.styles = styles
}

// Update implements tea.Model
func (m SettingsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.editing {
			return m.handleEditingKeys(msg)
		}
		return m.handleNavigationKeys(msg)
	}

	// Update active input if editing
	if m.editing {
		var cmd tea.Cmd
		switch m.focusedField {
		case FieldBinPath:
			m.binPathInput, cmd = m.binPathInput.Update(msg)
			cmds = append(cmds, cmd)
		case FieldTimeout:
			m.timeoutInput, cmd = m.timeoutInput.Update(msg)
			cmds = append(cmds, cmd)
		case FieldProxyURL:
			m.proxyURLInput, cmd = m.proxyURLInput.Update(msg)
			cmds = append(cmds, cmd)
		}
	}

	return m, tea.Batch(cmds...)
}

func (m SettingsModel) handleEditingKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg.String() {
	case "enter":
		m.editing = false
		m.saveCurrentField()
		m.blurAllInputs()
		return m, m.saveConfig()

	case "esc":
		m.editing = false
		m.revertCurrentField()
		m.blurAllInputs()
		return m, nil
	}

	// Update the focused input
	var cmd tea.Cmd
	switch m.focusedField {
	case FieldBinPath:
		m.binPathInput, cmd = m.binPathInput.Update(msg)
		cmds = append(cmds, cmd)
	case FieldTimeout:
		m.timeoutInput, cmd = m.timeoutInput.Update(msg)
		cmds = append(cmds, cmd)
	case FieldProxyURL:
		m.proxyURLInput, cmd = m.proxyURLInput.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m SettingsModel) handleNavigationKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.focusedField > 0 {
			m.focusedField--
		}
		return m, nil

	case "down", "j":
		if m.focusedField < FieldCount-1 {
			m.focusedField++
		}
		return m, nil

	case "enter", " ":
		return m.handleSelect()

	case "left", "h":
		return m.handleLeft()

	case "right", "l":
		return m.handleRight()
	}

	return m, nil
}

func (m SettingsModel) handleSelect() (tea.Model, tea.Cmd) {
	switch m.focusedField {
	case FieldBinPath:
		m.editing = true
		m.binPathInput.Focus()
		return m, textinput.Blink

	case FieldArchitecture:
		// Cycle through options
		m.architectureIndex = (m.architectureIndex + 1) % len(m.architectureOptions)
		m.config.Architecture = m.architectureOptions[m.architectureIndex]
		return m, m.saveConfig()

	case FieldShowExplanation:
		// Toggle boolean
		m.config.ShowExplanation = !m.config.ShowExplanation
		return m, m.saveConfig()

	case FieldTheme:
		// Cycle through options
		m.themeIndex = (m.themeIndex + 1) % len(m.themeOptions)
		m.config.Theme = m.themeOptions[m.themeIndex]
		return m, tea.Batch(
			m.saveConfig(),
			func() tea.Msg { return ThemeChangedMsg{Theme: m.config.Theme} },
		)

	case FieldTimeout:
		m.editing = true
		m.timeoutInput.Focus()
		return m, textinput.Blink

	case FieldProxyURL:
		m.editing = true
		m.proxyURLInput.Focus()
		return m, textinput.Blink
	}

	return m, nil
}

func (m SettingsModel) handleLeft() (tea.Model, tea.Cmd) {
	switch m.focusedField {
	case FieldArchitecture:
		m.architectureIndex--
		if m.architectureIndex < 0 {
			m.architectureIndex = len(m.architectureOptions) - 1
		}
		m.config.Architecture = m.architectureOptions[m.architectureIndex]
		return m, m.saveConfig()

	case FieldShowExplanation:
		m.config.ShowExplanation = !m.config.ShowExplanation
		return m, m.saveConfig()

	case FieldTheme:
		m.themeIndex--
		if m.themeIndex < 0 {
			m.themeIndex = len(m.themeOptions) - 1
		}
		m.config.Theme = m.themeOptions[m.themeIndex]
		return m, tea.Batch(
			m.saveConfig(),
			func() tea.Msg { return ThemeChangedMsg{Theme: m.config.Theme} },
		)
	}

	return m, nil
}

func (m SettingsModel) handleRight() (tea.Model, tea.Cmd) {
	switch m.focusedField {
	case FieldArchitecture:
		m.architectureIndex = (m.architectureIndex + 1) % len(m.architectureOptions)
		m.config.Architecture = m.architectureOptions[m.architectureIndex]
		return m, m.saveConfig()

	case FieldShowExplanation:
		m.config.ShowExplanation = !m.config.ShowExplanation
		return m, m.saveConfig()

	case FieldTheme:
		m.themeIndex = (m.themeIndex + 1) % len(m.themeOptions)
		m.config.Theme = m.themeOptions[m.themeIndex]
		return m, tea.Batch(
			m.saveConfig(),
			func() tea.Msg { return ThemeChangedMsg{Theme: m.config.Theme} },
		)
	}

	return m, nil
}

func (m *SettingsModel) saveCurrentField() {
	switch m.focusedField {
	case FieldBinPath:
		m.config.BinPath = m.binPathInput.Value()
	case FieldTimeout:
		if val, err := time.ParseDuration(m.timeoutInput.Value()); err == nil && val > 0 {
			m.config.Timeout = val
		} else {
			m.timeoutInput.SetValue(m.config.Timeout.String())
		}
	case FieldProxyURL:
		m.config.ProxyURL = m.proxyURLInput.Value()
	}
}

func (m *SettingsModel) revertCurrentField() {
	switch m.focusedField {
	case FieldBinPath:
		m.binPathInput.SetValue(m.config.BinPath)
	case FieldTimeout:
		m.timeoutInput.SetValue(m.config.Timeout.String())
	case FieldProxyURL:
		m.proxyURLInput.SetValue(m.config.ProxyURL)
	}
}

func (m *SettingsModel) blurAllInputs() {
	m.binPathInput.Blur()
	m.timeoutInput.Blur()
	m.proxyURLInput.Blur()
}

func (m SettingsModel) saveConfig() tea.Cmd {
	return func() tea.Msg {
		if err := m.config.Save(); err != nil {
			return settingsSavedMsg{err: err}
		}
		return settingsSavedMsg{}
	}
}

type settingsSavedMsg struct {
	err error
}

// View implements tea.Model
func (m SettingsModel) View() string {
	var s string

	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#7C3AED")).
		MarginBottom(1).
		Render("Settings")

	s += title + "\n\n"

	// Render each field
	fields := []SettingField{
		FieldBinPath,
		FieldArchitecture,
		FieldShowExplanation,
		FieldTheme,
		FieldTimeout,
		FieldProxyURL,
	}

	for _, field := range fields {
		s += m.renderField(field) + "\n"
	}

	// Status message
	if m.statusMsg != "" {
		statusStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#A6E3A1"))
		if m.statusIsErr {
			statusStyle = statusStyle.Foreground(lipgloss.Color("#F38BA8"))
		}
		s += "\n" + statusStyle.Render(m.statusMsg)
	}

	// Instructions
	instructions := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#6C7086")).
		MarginTop(1).
		Render("\nup/down: navigate  enter/space: edit/toggle  left/right: change value")

	s += instructions

	return s
}

func (m SettingsModel) renderField(field SettingField) string {
	isFocused := field == m.focusedField

	// Label style
	labelStyle := lipgloss.NewStyle().Width(28)
	if isFocused {
		labelStyle = labelStyle.Bold(true).Foreground(lipgloss.Color("#7C3AED"))
	} else {
		labelStyle = labelStyle.Foreground(lipgloss.Color("#CDD6F4"))
	}

	// Value style
	valueStyle := lipgloss.NewStyle()
	if isFocused {
		valueStyle = valueStyle.Foreground(lipgloss.Color("#06B6D4"))
	} else {
		valueStyle = valueStyle.Foreground(lipgloss.Color("#6C7086"))
	}

	// Cursor
	cursor := "  "
	if isFocused {
		cursor = "> "
	}

	label := labelStyle.Render(field.String())
	var value string

	switch field {
	case FieldBinPath:
		if m.editing && isFocused {
			value = m.binPathInput.View()
		} else {
			value = valueStyle.Render(m.config.BinPath)
		}

	case FieldArchitecture:
		value = m.renderSelectValue(m.architectureOptions, m.architectureIndex, isFocused)

	case FieldShowExplanation:
		if m.config.ShowExplanation {
			value = valueStyle.Render("[x] Enabled")
		} else {
			value = valueStyle.Render("[ ] Disabled")
		}

	case FieldTheme:
		value = m.renderSelectValue(m.themeOptions, m.themeIndex, isFocused)

	case FieldTimeout:
		if m.editing && isFocused {
			value = m.timeoutInput.View()
		} else {
			value = valueStyle.Render(fmt.Sprintf("%d seconds", m.config.Timeout))
		}

	case FieldProxyURL:
		if m.editing && isFocused {
			value = m.proxyURLInput.View()
		} else {
			if m.config.ProxyURL == "" {
				value = valueStyle.Render("(none)")
			} else {
				value = valueStyle.Render(m.config.ProxyURL)
			}
		}
	}

	// Description
	descStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#6C7086")).
		Italic(true).
		PaddingLeft(30)

	desc := ""
	if isFocused {
		desc = "\n" + descStyle.Render(field.Description())
	}

	return cursor + label + value + desc
}

func (m SettingsModel) renderSelectValue(options []string, selectedIndex int, isFocused bool) string {
	var parts []string

	for i, opt := range options {
		style := lipgloss.NewStyle()
		if i == selectedIndex {
			if isFocused {
				style = style.Bold(true).Foreground(lipgloss.Color("#06B6D4"))
			} else {
				style = style.Foreground(lipgloss.Color("#CDD6F4"))
			}
			parts = append(parts, style.Render("["+opt+"]"))
		} else {
			style = style.Foreground(lipgloss.Color("#6C7086"))
			parts = append(parts, style.Render(opt))
		}
	}

	arrowStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#6C7086"))
	if isFocused {
		arrowStyle = arrowStyle.Foreground(lipgloss.Color("#7C3AED"))
	}

	left := arrowStyle.Render("<")
	right := arrowStyle.Render(">")

	return left + " " + lipgloss.JoinHorizontal(lipgloss.Center, parts...) + " " + right
}
