package help

import (
	"strings"

	"github.com/jspback/bingus/bta/internal/ui"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func (m helpModel) Init() tea.Cmd {
	return nil
}

func (m helpModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c", "esc", "enter":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m helpModel) View() string {
	var sb strings.Builder
	styles := ui.CommonStyles()

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#5F87AF")).
		Padding(1, 2)

	logo := styles.TitleStyle.Render(`
     ██╗  ██╗███████╗██╗     ██████╗      ██████╗ ███████╗███╗   ██╗████████╗███████╗██████╗ 
     ██║  ██║██╔════╝██║     ██╔══██╗    ██╔════╝ ██╔════╝████╗  ██║╚══██╔══╝██╔════╝██╔══██╗
     ███████║█████╗  ██║     ██████╔╝    ██║      █████╗  ██╔██╗ ██║   ██║   █████╗  ██████╔╝
     ██╔══██║██╔══╝  ██║     ██╔═══╝     ██║      ██╔══╝  ██║╚██╗██║   ██║   ██╔══╝  ██╔══██╗
     ██║  ██║███████╗███████╗██║         ╚██████╗ ███████╗██║ ╚████║   ██║   ███████╗██║  ██║
     ╚═╝  ╚═╝╚══════╝╚══════╝╚═╝          ╚═════╝ ╚══════╝╚═╝  ╚═══╝   ╚═╝   ╚══════╝╚═╝  ╚═╝
       `)

	sb.WriteString(logo)
	sb.WriteString("\n")

	aboutContent := styles.DescriptionStyle.Render(
		"Bingus is a network discovery tool that helps you find hosts on your local network \n" +
			"and scan for open ports. Use it to explore your network securely and efficiently.")

	aboutBox := boxStyle.BorderForeground(lipgloss.Color("#5F87AF")).Render(
		styles.SectionStyle.Render("About Bingus") + "\n\n" + aboutContent)

	sb.WriteString(aboutBox)
	sb.WriteString("\n\n")

	commandsContent := strings.Builder{}
	commandsContent.WriteString(styles.CommandStyle.Render("Host Scan"))
	commandsContent.WriteString("\n")
	commandsContent.WriteString(styles.DescriptionStyle.Render("Scans your local network for active hosts using ICMP echo requests."))
	commandsContent.WriteString("\n")
	commandsContent.WriteString(styles.DescriptionStyle.Render("You can configure the timeout (ms) and max hosts to scan."))
	commandsContent.WriteString("\n\n")

	commandsContent.WriteString(styles.CommandStyle.Render("Port Scan"))
	commandsContent.WriteString("\n")
	commandsContent.WriteString(styles.DescriptionStyle.Render("Scans selected hosts for open TCP ports."))
	commandsContent.WriteString("\n")
	commandsContent.WriteString(styles.DescriptionStyle.Render("Configure port range and connection timeout for scanning."))
	commandsContent.WriteString("\n\n")

	commandsContent.WriteString(styles.CommandStyle.Render("Help"))
	commandsContent.WriteString("\n")
	commandsContent.WriteString(styles.DescriptionStyle.Render("Displays this help information."))

	commandsBox := boxStyle.BorderForeground(lipgloss.Color("#5F5FAF")).Render(
		styles.SectionStyle.Render("Available Commands") + "\n\n" + commandsContent.String())

	sb.WriteString(commandsBox)
	sb.WriteString("\n\n")

	sb.WriteString(styles.HelpStyle.Render("Press Enter, Esc, or q to return to main menu"))

	return sb.String()
}

func CreateHelpMenu() tea.Model {
	return &helpModel{
		width:  80,
		height: 24,
	}
}
