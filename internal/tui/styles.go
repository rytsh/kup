package tui

import "github.com/charmbracelet/lipgloss"

// Theme holds colors for the TUI
type Theme struct {
	Primary     lipgloss.Color
	Secondary   lipgloss.Color
	Accent      lipgloss.Color
	Background  lipgloss.Color
	Foreground  lipgloss.Color
	Muted       lipgloss.Color
	Success     lipgloss.Color
	Warning     lipgloss.Color
	Error       lipgloss.Color
	Border      lipgloss.Color
	TabActive   lipgloss.Color
	TabInactive lipgloss.Color
}

var (
	DefaultTheme = Theme{
		Primary:     lipgloss.Color("#7C3AED"),
		Secondary:   lipgloss.Color("#06B6D4"),
		Accent:      lipgloss.Color("#F59E0B"),
		Background:  lipgloss.Color("#1E1E2E"),
		Foreground:  lipgloss.Color("#CDD6F4"),
		Muted:       lipgloss.Color("#6C7086"),
		Success:     lipgloss.Color("#A6E3A1"),
		Warning:     lipgloss.Color("#F9E2AF"),
		Error:       lipgloss.Color("#F38BA8"),
		Border:      lipgloss.Color("#45475A"),
		TabActive:   lipgloss.Color("#7C3AED"),
		TabInactive: lipgloss.Color("#45475A"),
	}

	DarkTheme = Theme{
		Primary:     lipgloss.Color("#BB86FC"),
		Secondary:   lipgloss.Color("#03DAC6"),
		Accent:      lipgloss.Color("#CF6679"),
		Background:  lipgloss.Color("#121212"),
		Foreground:  lipgloss.Color("#E1E1E1"),
		Muted:       lipgloss.Color("#888888"),
		Success:     lipgloss.Color("#4CAF50"),
		Warning:     lipgloss.Color("#FFC107"),
		Error:       lipgloss.Color("#CF6679"),
		Border:      lipgloss.Color("#333333"),
		TabActive:   lipgloss.Color("#BB86FC"),
		TabInactive: lipgloss.Color("#333333"),
	}

	LightTheme = Theme{
		Primary:     lipgloss.Color("#6200EE"),
		Secondary:   lipgloss.Color("#03DAC6"),
		Accent:      lipgloss.Color("#FF5722"),
		Background:  lipgloss.Color("#FFFFFF"),
		Foreground:  lipgloss.Color("#1E1E1E"),
		Muted:       lipgloss.Color("#757575"),
		Success:     lipgloss.Color("#4CAF50"),
		Warning:     lipgloss.Color("#FF9800"),
		Error:       lipgloss.Color("#F44336"),
		Border:      lipgloss.Color("#E0E0E0"),
		TabActive:   lipgloss.Color("#6200EE"),
		TabInactive: lipgloss.Color("#E0E0E0"),
	}
)

// GetTheme returns the theme based on name
func GetTheme(name string) Theme {
	switch name {
	case "dark":
		return DarkTheme
	case "light":
		return LightTheme
	default:
		return DefaultTheme
	}
}

// Styles holds all the lipgloss styles for the application
type Styles struct {
	Theme Theme

	// App container
	App lipgloss.Style

	// Tabs
	TabBar       lipgloss.Style
	TabActive    lipgloss.Style
	TabInactive  lipgloss.Style
	TabSeparator lipgloss.Style

	// Content area
	Content lipgloss.Style

	// List items
	ListItem         lipgloss.Style
	ListItemSelected lipgloss.Style
	ListItemDesc     lipgloss.Style

	// Input fields
	InputLabel   lipgloss.Style
	InputField   lipgloss.Style
	InputFocused lipgloss.Style

	// Buttons
	Button        lipgloss.Style
	ButtonFocused lipgloss.Style

	// Status messages
	StatusSuccess lipgloss.Style
	StatusWarning lipgloss.Style
	StatusError   lipgloss.Style

	// Dialog/Modal
	DialogBox    lipgloss.Style
	DialogTitle  lipgloss.Style
	DialogButton lipgloss.Style

	// Help text
	Help     lipgloss.Style
	HelpKey  lipgloss.Style
	HelpDesc lipgloss.Style

	// Title
	Title    lipgloss.Style
	Subtitle lipgloss.Style

	// Progress
	ProgressBar lipgloss.Style

	// Command explanation
	CommandBox         lipgloss.Style
	CommandText        lipgloss.Style
	CommandExplanation lipgloss.Style
}

// NewStyles creates styles based on a theme
func NewStyles(theme Theme) *Styles {
	s := &Styles{Theme: theme}

	// App container
	s.App = lipgloss.NewStyle().
		Padding(1, 2)

	// Tab bar
	s.TabBar = lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderBottom(true).
		BorderForeground(theme.Border).
		MarginBottom(1)

	s.TabActive = lipgloss.NewStyle().
		Bold(true).
		Foreground(theme.Primary).
		Background(theme.Background).
		Padding(0, 2).
		BorderStyle(lipgloss.NormalBorder()).
		BorderBottom(true).
		BorderForeground(theme.Primary)

	s.TabInactive = lipgloss.NewStyle().
		Foreground(theme.Muted).
		Padding(0, 2).
		BorderStyle(lipgloss.HiddenBorder()).
		BorderBottom(true)

	s.TabSeparator = lipgloss.NewStyle().
		Foreground(theme.Muted).
		Padding(0, 1)

	// Content area
	s.Content = lipgloss.NewStyle().
		Padding(1, 0)

	// List items
	s.ListItem = lipgloss.NewStyle().
		Foreground(theme.Foreground).
		Padding(0, 2)

	s.ListItemSelected = lipgloss.NewStyle().
		Foreground(theme.Primary).
		Bold(true).
		Padding(0, 2).
		Background(lipgloss.Color("#313244"))

	s.ListItemDesc = lipgloss.NewStyle().
		Foreground(theme.Muted).
		Padding(0, 4)

	// Input fields
	s.InputLabel = lipgloss.NewStyle().
		Foreground(theme.Foreground).
		Bold(true).
		MarginRight(1)

	s.InputField = lipgloss.NewStyle().
		Foreground(theme.Foreground).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(theme.Border).
		Padding(0, 1)

	s.InputFocused = lipgloss.NewStyle().
		Foreground(theme.Primary).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(theme.Primary).
		Padding(0, 1)

	// Buttons
	s.Button = lipgloss.NewStyle().
		Foreground(theme.Foreground).
		Background(theme.Border).
		Padding(0, 2).
		MarginRight(1)

	s.ButtonFocused = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFFFF")).
		Background(theme.Primary).
		Bold(true).
		Padding(0, 2).
		MarginRight(1)

	// Status messages
	s.StatusSuccess = lipgloss.NewStyle().
		Foreground(theme.Success).
		Bold(true)

	s.StatusWarning = lipgloss.NewStyle().
		Foreground(theme.Warning).
		Bold(true)

	s.StatusError = lipgloss.NewStyle().
		Foreground(theme.Error).
		Bold(true)

	// Dialog
	s.DialogBox = lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(theme.Primary).
		Padding(1, 2).
		Width(60)

	s.DialogTitle = lipgloss.NewStyle().
		Foreground(theme.Primary).
		Bold(true).
		MarginBottom(1)

	s.DialogButton = lipgloss.NewStyle().
		Foreground(theme.Foreground).
		Background(theme.Border).
		Padding(0, 2)

	// Help
	s.Help = lipgloss.NewStyle().
		Foreground(theme.Muted).
		MarginTop(1)

	s.HelpKey = lipgloss.NewStyle().
		Foreground(theme.Secondary).
		Bold(true)

	s.HelpDesc = lipgloss.NewStyle().
		Foreground(theme.Muted)

	// Titles
	s.Title = lipgloss.NewStyle().
		Foreground(theme.Primary).
		Bold(true).
		MarginBottom(1)

	s.Subtitle = lipgloss.NewStyle().
		Foreground(theme.Muted).
		Italic(true)

	// Progress bar
	s.ProgressBar = lipgloss.NewStyle().
		Foreground(theme.Primary)

	// Command explanation box
	s.CommandBox = lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(theme.Accent).
		Padding(1, 2).
		MarginTop(1).
		MarginBottom(1)

	s.CommandText = lipgloss.NewStyle().
		Foreground(theme.Secondary).
		Bold(true)

	s.CommandExplanation = lipgloss.NewStyle().
		Foreground(theme.Foreground).
		MarginTop(1)

	return s
}
