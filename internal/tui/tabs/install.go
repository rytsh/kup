package tabs

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/rytsh/kup/internal/config"
	"github.com/rytsh/kup/internal/installer"
)

// InstallState represents the current state of the install tab
type InstallState int

const (
	StateList InstallState = iota
	StateConfirm
	StateInstalling
	StateDone
	StateError
)

// InstallModel is the model for the install tab
type InstallModel struct {
	config *config.Config
	styles interface{}
	width  int
	height int

	// Tools
	tools         []installer.Tool
	installerInst *installer.Installer

	// UI state
	state       InstallState
	selectedIdx int
	statusMsg   string
	errorMsg    string

	// Confirmation dialog
	confirmTool *installer.Tool

	// Progress
	spinner          spinner.Model
	progress         progress.Model
	downloadProgress installer.DownloadProgress
}

// NewInstallModel creates a new install model
func NewInstallModel(cfg *config.Config, styles interface{}) InstallModel {
	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#7C3AED"))

	prog := progress.New(progress.WithDefaultGradient())

	return InstallModel{
		config:        cfg,
		styles:        styles,
		tools:         installer.GetTools(),
		installerInst: installer.NewInstaller(cfg),
		state:         StateList,
		spinner:       sp,
		progress:      prog,
	}
}

// Init implements tea.Model
func (m InstallModel) Init() tea.Cmd {
	return m.spinner.Tick
}

// SetSize sets the size of the install view
func (m *InstallModel) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.progress.Width = width - 10
}

// UpdateStyles updates the styles
func (m *InstallModel) UpdateStyles(styles interface{}) {
	m.styles = styles
}

// Update implements tea.Model
func (m InstallModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch m.state {
		case StateList:
			return m.handleListKeys(msg)
		case StateConfirm:
			return m.handleConfirmKeys(msg)
		case StateError, StateDone:
			if msg.String() == "enter" || msg.String() == "esc" {
				m.state = StateList
				m.statusMsg = ""
				m.errorMsg = ""
				return m, nil
			}
		}

	case spinner.TickMsg:
		if m.state == StateInstalling {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			cmds = append(cmds, cmd)
		}

	case progress.FrameMsg:
		progressModel, cmd := m.progress.Update(msg)
		m.progress = progressModel.(progress.Model)
		cmds = append(cmds, cmd)

	case installProgressMsg:
		m.downloadProgress = msg.progress
		if msg.progress.Done {
			m.state = StateDone
			m.statusMsg = fmt.Sprintf("%s installed successfully!", msg.progress.Tool)
		}
		if msg.progress.Error != nil {
			m.state = StateError
			m.errorMsg = msg.progress.Error.Error()
		}
		if msg.progress.Total > 0 {
			percent := float64(msg.progress.Downloaded) / float64(msg.progress.Total)
			cmds = append(cmds, m.progress.SetPercent(percent))
		}

	case installCompleteMsg:
		if msg.err != nil {
			m.state = StateError
			m.errorMsg = msg.err.Error()
		} else {
			m.state = StateDone
			m.statusMsg = fmt.Sprintf("%s installed successfully!", msg.tool)
		}
	}

	return m, tea.Batch(cmds...)
}

func (m InstallModel) handleListKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.selectedIdx > 0 {
			m.selectedIdx--
		}
	case "down", "j":
		if m.selectedIdx < len(m.tools)-1 {
			m.selectedIdx++
		}
	case "enter", " ":
		tool := m.tools[m.selectedIdx]
		m.confirmTool = &tool
		m.state = StateConfirm
	}
	return m, nil
}

func (m InstallModel) handleConfirmKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y", "Y", "enter":
		if m.confirmTool != nil {
			m.state = StateInstalling
			return m, m.startInstall(*m.confirmTool)
		}
	case "n", "N", "esc":
		m.state = StateList
		m.confirmTool = nil
	}
	return m, nil
}

type installProgressMsg struct {
	progress installer.DownloadProgress
}

type installCompleteMsg struct {
	tool string
	err  error
}

func (m InstallModel) startInstall(tool installer.Tool) tea.Cmd {
	return func() tea.Msg {
		progressCh := make(chan installer.DownloadProgress, 10)
		ctx := context.Background()

		go func() {
			err := m.installerInst.Install(ctx, tool, progressCh)
			if err != nil {
				progressCh <- installer.DownloadProgress{
					Tool:  tool.Name,
					Error: err,
				}
			}
			close(progressCh)
		}()

		// Return first progress message
		for prog := range progressCh {
			return installProgressMsg{progress: prog}
		}

		return installCompleteMsg{tool: tool.Name}
	}
}

// View implements tea.Model
func (m InstallModel) View() string {
	switch m.state {
	case StateList:
		return m.viewList()
	case StateConfirm:
		return m.viewConfirm()
	case StateInstalling:
		return m.viewInstalling()
	case StateDone:
		return m.viewDone()
	case StateError:
		return m.viewError()
	}
	return ""
}

func (m InstallModel) viewList() string {
	var s strings.Builder

	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#7C3AED")).
		MarginBottom(1).
		Render("Install Kubernetes Tools")

	s.WriteString(title + "\n\n")

	osInfo, archInfo := installer.GetSystemInfo()
	sysInfo := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#6C7086")).
		Render(fmt.Sprintf("System: %s/%s  |  Target: %s", osInfo, archInfo, m.config.BinPath))
	s.WriteString(sysInfo + "\n\n")

	for i, tool := range m.tools {
		cursor := "  "
		if i == m.selectedIdx {
			cursor = "> "
		}

		// Tool name style
		nameStyle := lipgloss.NewStyle()
		if i == m.selectedIdx {
			nameStyle = nameStyle.Bold(true).Foreground(lipgloss.Color("#7C3AED"))
		} else {
			nameStyle = nameStyle.Foreground(lipgloss.Color("#CDD6F4"))
		}

		// Status
		status := m.installerInst.GetInstalledVersion(tool)
		statusStyle := lipgloss.NewStyle()
		if status == "installed" {
			statusStyle = statusStyle.Foreground(lipgloss.Color("#A6E3A1"))
		} else {
			statusStyle = statusStyle.Foreground(lipgloss.Color("#6C7086"))
		}

		name := nameStyle.Render(tool.Name)
		statusText := statusStyle.Render(fmt.Sprintf("[%s]", status))

		s.WriteString(fmt.Sprintf("%s%s %s\n", cursor, name, statusText))

		// Description
		descStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6C7086")).
			PaddingLeft(4)
		s.WriteString(descStyle.Render(tool.Description) + "\n\n")
	}

	// Instructions
	instructions := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#6C7086")).
		MarginTop(1).
		Render("\nup/down: select  enter: install")
	s.WriteString(instructions)

	return s.String()
}

func (m InstallModel) viewConfirm() string {
	if m.confirmTool == nil {
		return ""
	}

	var s strings.Builder

	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#F59E0B")).
		Render(fmt.Sprintf("Install %s?", m.confirmTool.Name))

	s.WriteString(title + "\n\n")

	// Show explanation if enabled
	if m.config.ShowExplanation {
		explanationBox := lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#F59E0B")).
			Padding(1, 2).
			Width(m.width - 4)

		commandTitle := lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#06B6D4")).
			Render("Command to execute:")

		command := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#A6E3A1")).
			Render(m.confirmTool.GetCommand(m.config))

		explanation := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#CDD6F4")).
			MarginTop(1).
			Render(m.confirmTool.Explanation)

		content := commandTitle + "\n\n" + command + "\n" + explanation
		s.WriteString(explanationBox.Render(content) + "\n\n")
	}

	// Buttons
	confirmStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#1E1E2E")).
		Background(lipgloss.Color("#A6E3A1")).
		Padding(0, 2)

	cancelStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#CDD6F4")).
		Background(lipgloss.Color("#45475A")).
		Padding(0, 2)

	buttons := confirmStyle.Render("[Y] Yes, install") + "  " + cancelStyle.Render("[N] Cancel")
	s.WriteString(buttons)

	return s.String()
}

func (m InstallModel) viewInstalling() string {
	var s strings.Builder

	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#06B6D4")).
		Render(fmt.Sprintf("Installing %s...", m.confirmTool.Name))

	s.WriteString(title + "\n\n")
	s.WriteString(m.spinner.View() + " Downloading...\n\n")

	if m.downloadProgress.Total > 0 {
		s.WriteString(m.progress.View() + "\n")
		percent := float64(m.downloadProgress.Downloaded) / float64(m.downloadProgress.Total) * 100
		s.WriteString(fmt.Sprintf("%.1f%% (%d / %d bytes)\n",
			percent, m.downloadProgress.Downloaded, m.downloadProgress.Total))
	}

	return s.String()
}

func (m InstallModel) viewDone() string {
	var s strings.Builder

	successStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#A6E3A1"))

	s.WriteString(successStyle.Render("Installation Complete!") + "\n\n")
	s.WriteString(m.statusMsg + "\n\n")

	infoStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#6C7086"))
	s.WriteString(infoStyle.Render(fmt.Sprintf("Binary installed to: %s", m.config.BinPath)) + "\n\n")

	s.WriteString("Press Enter to continue...")

	return s.String()
}

func (m InstallModel) viewError() string {
	var s strings.Builder

	errorStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#F38BA8"))

	s.WriteString(errorStyle.Render("Installation Failed") + "\n\n")

	errorBox := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#F38BA8")).
		Padding(1, 2).
		Foreground(lipgloss.Color("#F38BA8"))

	s.WriteString(errorBox.Render(m.errorMsg) + "\n\n")

	// Show manual command
	if m.confirmTool != nil && m.config.ShowExplanation {
		manualStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6C7086"))
		s.WriteString(manualStyle.Render("You can try running the command manually:") + "\n\n")

		cmdStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#A6E3A1"))
		s.WriteString(cmdStyle.Render(m.confirmTool.GetCommand(m.config)) + "\n\n")
	}

	s.WriteString("Press Enter to continue...")

	return s.String()
}
