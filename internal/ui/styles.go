package ui

import "github.com/charmbracelet/lipgloss"

// Color palette -- warm, distinctive, not generic.
var (
	// Base colors
	ColorBg        = lipgloss.Color("#1a1b26") // deep navy bg
	ColorFg        = lipgloss.Color("#c0caf5") // soft blue-white
	ColorFgDim     = lipgloss.Color("#565f89") // muted comment
	ColorFgMuted   = lipgloss.Color("#7982a9") // slightly brighter muted
	ColorAccent    = lipgloss.Color("#7aa2f7") // bright blue accent
	ColorAccentDim = lipgloss.Color("#3d59a1") // dim blue

	// Priority colors
	ColorHigh   = lipgloss.Color("#f7768e") // red-pink
	ColorMedium = lipgloss.Color("#e0af68") // warm amber
	ColorLow    = lipgloss.Color("#7dcfff") // sky blue
	ColorNone   = lipgloss.Color("#444b6a") // very dim

	// Status colors
	ColorOverdue = lipgloss.Color("#f7768e") // same red
	ColorToday   = lipgloss.Color("#9ece6a") // green
	ColorSoon    = lipgloss.Color("#e0af68") // amber
	ColorFuture  = lipgloss.Color("#7dcfff") // sky

	// UI chrome
	ColorBorder    = lipgloss.Color("#3b4261")
	ColorSelection = lipgloss.Color("#283457") // selection background
	ColorSuccess   = lipgloss.Color("#9ece6a")
	ColorError     = lipgloss.Color("#f7768e")

	// List tag colors (rotate through these)
	ListColors = []lipgloss.Color{
		lipgloss.Color("#bb9af7"), // purple
		lipgloss.Color("#7aa2f7"), // blue
		lipgloss.Color("#2ac3de"), // cyan
		lipgloss.Color("#9ece6a"), // green
		lipgloss.Color("#e0af68"), // amber
		lipgloss.Color("#ff9e64"), // orange
	}
)

// Styles
var (
	// Header
	HeaderStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorFg).
			PaddingLeft(1)

	ViewTitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorAccent)

	CountStyle = lipgloss.NewStyle().
			Foreground(ColorFgDim)

	// Tab bar
	ActiveTabStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorAccent).
			Underline(true)

	InactiveTabStyle = lipgloss.NewStyle().
				Foreground(ColorFgDim)

	TabSepStyle = lipgloss.NewStyle().
			Foreground(ColorFgDim)

	// Reminder items
	SelectedStyle = lipgloss.NewStyle().
			Background(ColorSelection).
			PaddingLeft(1).
			PaddingRight(1)

	NormalItemStyle = lipgloss.NewStyle().
			PaddingLeft(2)

	TitleStyle = lipgloss.NewStyle().
			Foreground(ColorFg)

	TitleOverdueStyle = lipgloss.NewStyle().
				Foreground(ColorOverdue).
				Bold(true)

	DueDateStyle = lipgloss.NewStyle().
			Foreground(ColorFgMuted)

	DueDateOverdueStyle = lipgloss.NewStyle().
				Foreground(ColorOverdue)

	NotesStyle = lipgloss.NewStyle().
			Foreground(ColorFgDim).
			Italic(true)

	// Group headers
	GroupHeaderStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(ColorAccent).
				PaddingLeft(1).
				MarginTop(1)

	GroupHeaderOverdueStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(ColorOverdue).
				PaddingLeft(1).
				MarginTop(1)

	// Status bar
	StatusBarStyle = lipgloss.NewStyle().
			Foreground(ColorFgDim).
			PaddingLeft(1)

	StatusErrorStyle = lipgloss.NewStyle().
				Foreground(ColorError).
				PaddingLeft(1)

	StatusSuccessStyle = lipgloss.NewStyle().
				Foreground(ColorSuccess).
				PaddingLeft(1)

	// Input
	InputLabelStyle = lipgloss.NewStyle().
			Foreground(ColorAccent).
			Bold(true)

	InputStyle = lipgloss.NewStyle().
			Foreground(ColorFg)

	// Help overlay
	HelpKeyStyle = lipgloss.NewStyle().
			Foreground(ColorAccent).
			Bold(true).
			Width(12)

	HelpDescStyle = lipgloss.NewStyle().
			Foreground(ColorFgMuted)

	HelpTitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorFg).
			MarginBottom(1)

	// Spinner
	SpinnerStyle = lipgloss.NewStyle().
			Foreground(ColorAccent)

	// Confirmation
	ConfirmStyle = lipgloss.NewStyle().
			Foreground(ColorError).
			Bold(true)
)

// ListTagColor returns a consistent color for a list name.
func ListTagColor(listName string) lipgloss.Color {
	h := 0
	for _, c := range listName {
		h = h*31 + int(c)
	}
	if h < 0 {
		h = -h
	}
	return ListColors[h%len(ListColors)]
}

// ListTagStyle returns a styled tag for a list name.
func ListTagStyle(listName string) lipgloss.Style {
	return lipgloss.NewStyle().Foreground(ListTagColor(listName))
}
