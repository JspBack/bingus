package main

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/jspback/bingus/bta/internal/help"
	"github.com/jspback/bingus/bta/internal/ping"
	"github.com/jspback/bingus/bta/internal/port"
	"github.com/jspback/bingus/bta/internal/ui"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type item string

func (i item) String() string {
	return string(i)
}

func (i item) FilterValue() string { return i.String() }

type itemDelegate struct{}

func (d itemDelegate) Height() int                             { return 1 }
func (d itemDelegate) Spacing() int                            { return 0 }
func (d itemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(item)
	if !ok {
		return
	}

	str := fmt.Sprintf("%d. %s", index+1, i)

	fn := ui.CommonStyles().ItemStyle.Render
	if index == m.Index() {
		fn = func(s ...string) string {
			return ui.CommonStyles().SelectedItemStyle.Render("> " + strings.Join(s, " "))
		}
	}

	fmt.Fprint(w, fn(str))
}

type model struct {
	list     list.Model
	choice   string
	quitting bool
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.list.SetWidth(msg.Width)
		return m, nil

	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case "q", "ctrl+c", "esc":
			m.quitting = true
			return m, tea.Quit

		case "enter":
			i, ok := m.list.SelectedItem().(item)
			if ok {
				m.choice = string(i)
			}
			return m, nil
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m model) View() string {
	if m.choice != "" {
		return ""
	}
	if m.quitting {
		return ui.CommonStyles().QuitTextStyle.Render("Goodbye !")
	}

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#5F87AF")).
		Padding(1, 2)

	logo := ui.CommonStyles().TitleStyle.Render(`
     ██████╗ ██╗███╗   ██╗ ██████╗ ██╗   ██╗███████╗
     ██╔══██╗██║████╗  ██║██╔════╝ ██║   ██║██╔════╝
     ██████╔╝██║██╔██╗ ██║██║  ███╗██║   ██║███████╗
     ██╔══██╗██║██║╚██╗██║██║   ██║██║   ██║╚════██║
     ██████╔╝██║██║ ╚████║╚██████╔╝╚██████╔╝███████║
     ╚═════╝ ╚═╝╚═╝  ╚═══╝ ╚═════╝  ╚══════╝
     
     🌐 Network Scanner Tool 🌐
`)

	var sb strings.Builder
	sb.WriteString(logo)
	sb.WriteString("\n\n")

	listView := m.list.View()
	sb.WriteString(boxStyle.Render(listView))

	return sb.String()
}

type appModel struct {
	mainMenu   model
	pingUI     ping.UIPingModel
	portUI     port.UIPortModel
	helpUI     tea.Model
	activeView string
}

func (m appModel) Init() tea.Cmd {
	return m.mainMenu.Init()
}

func (m appModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m.activeView {
	case "main":
		newM, cmd := m.mainMenu.Update(msg)
		m.mainMenu = newM.(model)

		if m.mainMenu.choice == "Host Scan" {
			m.activeView = "ping"
			m.pingUI = ping.NewUIPingModel()
			return m, m.pingUI.Init()
		} else if m.mainMenu.choice == "Port Scan" {
			m.activeView = "port"
			m.portUI = port.NewUIPortModel()

			if len(m.pingUI.GetHosts()) > 0 {
				m.portUI.SetHosts(m.pingUI.GetHosts())
			}

			return m, m.portUI.Init()
		} else if m.mainMenu.choice == "Help" {
			m.activeView = "help"
			return m, nil
		}

		if m.mainMenu.quitting {
			return m, tea.Quit
		}
		return m, cmd

	case "ping":
		newM, cmd := m.pingUI.Update(msg)
		m.pingUI = newM.(ping.UIPingModel)
		if m.pingUI.View() == "Returning to main menu...\n" {
			m.activeView = "main"
			m.mainMenu.choice = ""
			return m, nil
		}
		return m, cmd

	case "port":
		newM, cmd := m.portUI.Update(msg)
		m.portUI = newM.(port.UIPortModel)
		if m.portUI.View() == "Returning to main menu...\n" {
			m.activeView = "main"
			m.mainMenu.choice = ""
			return m, nil
		}
		return m, cmd

	case "help":
		newM, cmd := m.helpUI.Update(msg)
		m.helpUI = newM
		if cmd != nil {
			m.activeView = "main"
			m.mainMenu.choice = ""
			return m, nil
		}
		return m, cmd

	default:
		return m, nil
	}
}

func (m appModel) View() string {
	switch m.activeView {
	case "ping":
		return m.pingUI.View()
	case "port":
		return m.portUI.View()
	case "help":
		return m.helpUI.View()
	default:
		return m.mainMenu.View()
	}
}

func main() {
	items := []list.Item{
		item("Host Scan"),
		item("Port Scan"),
		item("Help"),
	}

	l := list.New(items, itemDelegate{}, 20, 20)
	l.Title = "Select an option to begin scanning your network"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(true)
	l.Styles.Title = ui.CommonStyles().TitleStyle
	l.Styles.PaginationStyle = ui.CommonStyles().PaginationStyle
	l.Styles.HelpStyle = ui.CommonStyles().HelpStyle

	app := appModel{
		mainMenu:   model{list: l},
		pingUI:     ping.NewUIPingModel(),
		portUI:     port.NewUIPortModel(),
		helpUI:     help.CreateHelpMenu(),
		activeView: "main",
	}

	p := tea.NewProgram(app)

	ping.SetProgram(p)
	port.SetProgram(p)

	if _, err := p.Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
