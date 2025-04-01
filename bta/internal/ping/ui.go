package ping

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/jspback/bingus/bta/internal/ui"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	StateInput PingState = iota
	StateScanning
	StateResults
)

func NewUIPingModel() UIPingModel {
	styles := ui.CommonStyles()

	inputs := make([]textinput.Model, 2)

	inputs[0] = textinput.New()
	inputs[0].Placeholder = "1000"
	inputs[0].Focus()
	inputs[0].Width = 20
	inputs[0].Prompt = "Timeout (ms): "
	inputs[0].SetValue("1000")

	inputs[1] = textinput.New()
	inputs[1].Placeholder = "50"
	inputs[1].Width = 20
	inputs[1].Prompt = "Max hosts: "
	inputs[1].SetValue("50")

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = styles.SuccessStyle

	return UIPingModel{
		state:      StateInput,
		inputs:     inputs,
		focusIndex: 0,
		spinner:    s,
		styles:     styles,
		width:      80,
		height:     24,
	}
}

func (m UIPingModel) Init() tea.Cmd {
	return textinput.Blink
}

func startScan(timeout time.Duration, maxHosts int) tea.Cmd {
	return func() tea.Msg {
		hostFoundCh := make(chan string, 100)

		go func() {
			for host := range hostFoundCh {
				program.Send(hostFoundMsg(host))
			}
		}()

		hosts, err := hostDiscovery(timeout, hostFoundCh, maxHosts)

		close(hostFoundCh)

		return scanDoneMsg{Hosts: hosts, Err: err}
	}
}

var program *tea.Program

func SetProgram(p *tea.Program) {
	program = p
}

func (m UIPingModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			if m.state == StateScanning {
				m.state = StateResults
				return m, nil
			}
			m.quitting = true
			return m, nil

		case "tab", "shift+tab":
			if m.state == StateInput {
				if msg.String() == "tab" {
					m.focusIndex = (m.focusIndex + 1) % len(m.inputs)
				} else {
					m.focusIndex = (m.focusIndex - 1 + len(m.inputs)) % len(m.inputs)
				}

				for i := 0; i < len(m.inputs); i++ {
					if i == m.focusIndex {
						m.inputs[i].Focus()
					} else {
						m.inputs[i].Blur()
					}
				}

				return m, nil
			}

		case "enter":
			if m.state == StateInput {
				m.state = StateScanning
				m.scanning = true

				timeoutStr := m.inputs[0].Value()
				maxHostsStr := m.inputs[1].Value()

				timeout := 1000
				maxHosts := 50

				if t, err := strconv.Atoi(timeoutStr); err == nil && t > 0 {
					timeout = t
				}

				if h, err := strconv.Atoi(maxHostsStr); err == nil && h > 0 {
					maxHosts = h
				}

				return m, tea.Batch(
					m.spinner.Tick,
					startScan(time.Duration(timeout)*time.Millisecond, maxHosts),
				)
			} else if m.state == StateResults {
				m.quitting = true
				return m, nil
			}
		}

	case spinner.TickMsg:
		if m.state == StateScanning {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}

	case hostFoundMsg:
		host := string(msg)
		m.scanResult = append(m.scanResult, host)
		if m.state == StateScanning {
			return m, m.spinner.Tick
		}

	case scanDoneMsg:
		m.scanning = false
		m.state = StateResults
		if msg.Err != nil {
			m.error = msg.Err
		}
		return m, nil
	}

	if m.state == StateInput {
		cmd := m.updateInputs(msg)
		return m, cmd
	}

	return m, nil
}

func (m *UIPingModel) updateInputs(msg tea.Msg) tea.Cmd {
	var cmds []tea.Cmd
	for i := range m.inputs {
		m.inputs[i], _ = m.inputs[i].Update(msg)
	}
	return tea.Batch(cmds...)
}

func (m UIPingModel) View() string {
	if m.quitting {
		return "Returning to main menu...\n"
	}

	var sb strings.Builder

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#5F87AF")).
		Padding(1, 2).
		Width(m.width - 4)

	logo := m.styles.TitleStyle.Render(`
  ██████╗ ██╗███╗   ██╗ ██████╗     ███████╗ ██████╗ █████╗ ███╗   ██╗
  ██╔══██╗██║████╗  ██║██╔════╝     ██╔════╝██╔════╝██╔══██╗████╗  ██║
  ██████╔╝██║██╔██╗ ██║██║  ███╗    ███████╗██║     ███████║██╔██╗ ██║
  ██╔═══╝ ██║██║╚██╗██║██║   ██║    ╚════██║██║     ██╔══██║██║╚██╗██║
  ██║     ██║██║ ╚████║╚██████╔╝    ███████║╚██████╗██║  ██║██║ ╚████║
  ╚═╝     ╚═╝╚═╝  ╚═══╝ ╚═════╝     ╚══════╝ ╚═════╝╚═╝  ╚═╝╚═╝  ╚═══╝
    `)

	sb.WriteString(logo)
	sb.WriteString("\n")

	switch m.state {
	case StateInput:
		description := m.styles.DescriptionStyle.Render("Configure scan parameters to discover hosts on your local network")
		sb.WriteString(boxStyle.Render(description))
		sb.WriteString("\n\n")

		inputBox := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#5F5FAF")).
			Padding(1, 2).
			Width(m.width - 10)

		var inputsContent strings.Builder
		for i, input := range m.inputs {
			inputsContent.WriteString(input.View())
			if i < len(m.inputs)-1 {
				inputsContent.WriteString("\n\n")
			}
		}

		sb.WriteString(inputBox.Render(inputsContent.String()))
		sb.WriteString("\n\n")

		sb.WriteString(m.styles.HelpStyle.Render("Press Enter to start scan, Tab to switch fields, Esc to go back"))

	case StateScanning:
		scanningBox := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#5F5FAF")).
			Padding(1, 2).
			Width(m.width - 10)

		scanningContent := fmt.Sprintf("%s Scanning local network...", m.spinner.View())
		sb.WriteString(scanningBox.Render(scanningContent))
		sb.WriteString("\n\n")

		if len(m.scanResult) > 0 {
			resultsBox := lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("#00FF00")).
				Padding(1, 2).
				Width(m.width - 10)

			var resultsContent strings.Builder
			resultsContent.WriteString(m.styles.SectionStyle.Render("Hosts found so far:"))
			resultsContent.WriteString("\n\n")

			hostStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("#00FF00")).
				Bold(true).
				PaddingLeft(0)

			for _, host := range m.scanResult {
				resultsContent.WriteString(hostStyle.Render(host) + "\n")
			}

			sb.WriteString(resultsBox.Render(resultsContent.String()))
		} else {
			sb.WriteString(m.styles.DescriptionStyle.Render("Searching for hosts..."))
		}

		sb.WriteString("\n\n")
		sb.WriteString(m.styles.HelpStyle.Render("Press Esc to stop scanning"))

	case StateResults:
		resultsBox := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#00FF00")).
			Padding(1, 2).
			Width(m.width - 10)

		var resultsContent strings.Builder

		if m.error != nil {
			resultsContent.WriteString(m.styles.ErrorStyle.Render(fmt.Sprintf("Error: %v\n\n", m.error)))
		}

		resultsContent.WriteString(m.styles.SectionStyle.Render("Scan Results"))
		resultsContent.WriteString("\n\n")

		if len(m.scanResult) > 0 {
			resultsContent.WriteString(m.styles.SuccessStyle.Render(fmt.Sprintf("Found %d hosts:\n\n", len(m.scanResult))))

			hostStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("#00FF00")).
				Bold(true).
				PaddingLeft(0)

			for _, host := range m.scanResult {
				resultsContent.WriteString(hostStyle.Render(host) + "\n")
			}
		} else {
			resultsContent.WriteString(m.styles.WarningStyle.Render("No hosts found on the network.\n"))
		}

		sb.WriteString(resultsBox.Render(resultsContent.String()))
		sb.WriteString("\n\n")
		sb.WriteString(m.styles.HelpStyle.Render("Press Enter to return to main menu"))
	}

	return sb.String()
}

func (m UIPingModel) GetHosts() []string {
	return m.scanResult
}
