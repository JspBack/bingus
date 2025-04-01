package ui

import (
	"github.com/charmbracelet/lipgloss"
)

type Styles struct {
	TitleStyle       lipgloss.Style
	InputLabelStyle  lipgloss.Style
	ActiveInputStyle lipgloss.Style

	SectionStyle     lipgloss.Style
	DescriptionStyle lipgloss.Style
	CommandStyle     lipgloss.Style

	CountStyle                lipgloss.Style
	SuccessStyle              lipgloss.Style
	ErrorStyle                lipgloss.Style
	WarningStyle              lipgloss.Style
	ButtonStyle               lipgloss.Style
	ActiveButtonStyle         lipgloss.Style
	CheckedStyle              lipgloss.Style
	ActiveCheckboxStyle       lipgloss.Style
	HostStyle                 lipgloss.Style
	PortStyle                 lipgloss.Style
	HostFoundStyle            lipgloss.Style
	ItemStyle                 lipgloss.Style
	SelectedItemStyle         lipgloss.Style
	DisabledItemStyle         lipgloss.Style
	SelectedDisabledItemStyle lipgloss.Style
	PaginationStyle           lipgloss.Style
	HelpStyle                 lipgloss.Style
	QuitTextStyle             lipgloss.Style
	AsciiArtStyle             lipgloss.Style
}

func CommonStyles() *Styles {
	return &Styles{
		// General styles
		TitleStyle:       lipgloss.NewStyle().MarginLeft(2).Bold(true),
		InputLabelStyle:  lipgloss.NewStyle().Width(10),
		ActiveInputStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("170")),
		CountStyle:       lipgloss.NewStyle().Foreground(lipgloss.Color("170")).Bold(true),
		SuccessStyle:     lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Bold(true),
		ErrorStyle:       lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Bold(true),
		WarningStyle:     lipgloss.NewStyle().Foreground(lipgloss.Color("9")),

		// Button styles
		ButtonStyle:       lipgloss.NewStyle().Background(lipgloss.Color("240")).Padding(0, 3),
		ActiveButtonStyle: lipgloss.NewStyle().Background(lipgloss.Color("170")).Foreground(lipgloss.Color("0")).Padding(0, 3),

		// Checkbox styles
		CheckedStyle:        lipgloss.NewStyle().Foreground(lipgloss.Color("10")),
		ActiveCheckboxStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("170")),

		// Host and port styles
		HostStyle:      lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12")),
		PortStyle:      lipgloss.NewStyle().Foreground(lipgloss.Color("13")),
		HostFoundStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("10")),

		// Menu item styles
		ItemStyle:                 lipgloss.NewStyle().PaddingLeft(4),
		SelectedItemStyle:         lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("170")),
		DisabledItemStyle:         lipgloss.NewStyle().PaddingLeft(4).Foreground(lipgloss.Color("240")).Strikethrough(true),
		SelectedDisabledItemStyle: lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("240")).Strikethrough(true),

		// List styles
		PaginationStyle: lipgloss.NewStyle().PaddingLeft(4),
		HelpStyle:       lipgloss.NewStyle().PaddingLeft(4).PaddingBottom(1),
		QuitTextStyle:   lipgloss.NewStyle().Margin(1, 0, 2, 4),
		AsciiArtStyle:   lipgloss.NewStyle().Foreground(lipgloss.Color("170")).Bold(true),

		// Section and description styles
		SectionStyle:     lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39")),
		CommandStyle:     lipgloss.NewStyle().Foreground(lipgloss.Color("205")),
		DescriptionStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("252")),
	}
}
