package port

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/jspback/bingus/bta/internal/ui"
	"github.com/jspback/bingus/bta/internal/utils"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	StateHostSelection PortState = iota
	StatePortConfig
	StateScanning
	StateResults
)

func NewUIPortModel() UIPortModel {
	styles := ui.CommonStyles()

	inputs := make([]textinput.Model, 3)

	inputs[0] = textinput.New()
	inputs[0].Placeholder = "1"
	inputs[0].Focus()
	inputs[0].Width = 20
	inputs[0].Prompt = "Start Port: "
	inputs[0].SetValue("1")

	inputs[1] = textinput.New()
	inputs[1].Placeholder = "1024"
	inputs[1].Width = 20
	inputs[1].Prompt = "End Port: "
	inputs[1].SetValue("1024")

	inputs[2] = textinput.New()
	inputs[2].Placeholder = "500"
	inputs[2].Width = 20
	inputs[2].Prompt = "Timeout (ms): "
	inputs[2].SetValue("500")

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = styles.SuccessStyle

	return UIPortModel{
		state:          StateHostSelection,
		hosts:          []HostItem{},
		inputs:         inputs,
		focusIndex:     0,
		spinner:        s,
		scanResults:    make(map[string][]int),
		scanProgress:   make(map[string]int),
		styles:         styles,
		width:          80,
		height:         24,
		useCommonPorts: false,
	}
}

func (m UIPortModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m *UIPortModel) SetHosts(hosts []string) {
	m.hosts = make([]HostItem, len(hosts))
	for i, host := range hosts {
		m.hosts[i] = HostItem{Host: host, Selected: false}
	}
}

func startScan(ctx context.Context, hosts []string, startPort, endPort int, timeout time.Duration, useCommonPorts bool) tea.Cmd {
	return func() tea.Msg {
		portFoundCh := make(chan PortResult, 100)

		hostPortMap := make(map[string]map[int]bool)
		for _, host := range hosts {
			hostPortMap[host] = make(map[int]bool)
		}

		portsToScan := []int{}
		if useCommonPorts {
			portsToScan = utils.GetCommonPorts()
		} else {
			for port := startPort; port <= endPort; port++ {
				portsToScan = append(portsToScan, port)
			}
		}

		for _, host := range hosts {
			go func(host string) {
				for _, port := range portsToScan {
					hostPortMap[host][port] = false
				}
			}(host)
		}

		go func() {
			currentHost := ""
			currentPort := 0

			for _, host := range hosts {
				for _, port := range portsToScan {
					currentHost = host
					currentPort = port

					program.Send(portFoundMsg{
						Host: currentHost,
						Port: currentPort,
						Open: false,
					})
				}
			}

			for result := range portFoundCh {
				if result.Open {
					for host, ports := range hostPortMap {
						if _, exists := ports[result.Port]; exists {
							program.Send(portFoundMsg{
								Host: host,
								Port: result.Port,
								Open: true,
							})
						}
					}
				}
			}
		}()

		results, err := portDiscovery(ctx, hosts, portsToScan, timeout, portFoundCh)

		close(portFoundCh)

		return scanDoneMsg{Results: results, Err: err}
	}
}

var program *tea.Program

func SetProgram(p *tea.Program) {
	program = p
}

func (m UIPortModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			if m.scanning && m.cancel != nil {
				m.cancel()
				m.scanning = false
				m.state = StateResults
				return m, nil
			}

			if m.state == StatePortConfig {
				m.state = StateHostSelection
				return m, nil
			}

			m.quitting = true
			return m, nil

		case "up", "k":
			if m.state == StateHostSelection {
				if m.cursor > 0 {
					m.cursor--
				} else if len(m.hosts) > 0 {
					m.cursor = len(m.hosts)
				}
			} else if m.state == StatePortConfig {
				if m.focusIndex == len(m.inputs) {
					m.inputs[len(m.inputs)-1].Focus()
					m.focusIndex = len(m.inputs) - 1
				} else if m.focusIndex > 0 {
					m.focusIndex--
					for i := range m.inputs {
						if i == m.focusIndex {
							m.inputs[i].Focus()
						} else {
							m.inputs[i].Blur()
						}
					}
				}
			}
			return m, nil

		case "down", "j":
			if m.state == StateHostSelection {
				if m.cursor < len(m.hosts) {
					m.cursor++
				} else {
					m.cursor = 0
				}
			} else if m.state == StatePortConfig {
				if m.focusIndex < len(m.inputs)-1 {
					m.focusIndex++
					for i := range m.inputs {
						if i == m.focusIndex {
							m.inputs[i].Focus()
						} else {
							m.inputs[i].Blur()
						}
					}
				} else if m.focusIndex == len(m.inputs)-1 {
					m.inputs[m.focusIndex].Blur()
					m.focusIndex = len(m.inputs)
				}
			}
			return m, nil

		case " ":
			if m.state == StateHostSelection {
				if m.cursor < len(m.hosts) {
					m.hosts[m.cursor].Selected = !m.hosts[m.cursor].Selected
				} else {
					m.selectAll = !m.selectAll
					for i := range m.hosts {
						m.hosts[i].Selected = m.selectAll
					}
				}
			} else if m.state == StatePortConfig && m.focusIndex == len(m.inputs) {
				m.useCommonPorts = !m.useCommonPorts
			}
			return m, nil

		case "tab", "shift+tab":
			if m.state == StatePortConfig {
				if msg.String() == "tab" {
					m.focusIndex = (m.focusIndex + 1) % (len(m.inputs) + 1)
				} else {
					m.focusIndex = (m.focusIndex - 1 + len(m.inputs) + 1) % (len(m.inputs) + 1)
				}

				for i := range m.inputs {
					if i == m.focusIndex {
						m.inputs[i].Focus()
					} else {
						m.inputs[i].Blur()
					}
				}

				return m, nil
			}

		case "enter":
			if m.state == StateHostSelection {
				selectedCount := 0
				for _, h := range m.hosts {
					if h.Selected {
						selectedCount++
					}
				}

				if selectedCount > 0 {
					m.state = StatePortConfig
					return m, nil
				}

				return m, nil

			} else if m.state == StatePortConfig {
				startPortStr := m.inputs[0].Value()
				endPortStr := m.inputs[1].Value()
				timeoutStr := m.inputs[2].Value()

				startPort := 1
				endPort := 1024
				timeout := 500

				if p, err := strconv.Atoi(startPortStr); err == nil && p > 0 && p < 65536 {
					startPort = p
				}

				if p, err := strconv.Atoi(endPortStr); err == nil && p > 0 && p < 65536 {
					endPort = p
				}

				if t, err := strconv.Atoi(timeoutStr); err == nil && t > 0 {
					timeout = t
				}

				if startPort > endPort {
					startPort, endPort = endPort, startPort
				}

				selectedHosts := make([]string, 0, len(m.hosts))
				for _, h := range m.hosts {
					if h.Selected {
						selectedHosts = append(selectedHosts, h.Host)
					}
				}

				m.state = StateScanning
				m.scanning = true
				m.scanResults = make(map[string][]int)
				m.scanProgress = make(map[string]int)

				ctx, cancel := context.WithCancel(context.Background())
				m.cancel = cancel

				return m, tea.Batch(
					m.spinner.Tick,
					startScan(ctx, selectedHosts, startPort, endPort, time.Duration(timeout)*time.Millisecond, m.useCommonPorts),
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

	case portFoundMsg:

		if msg.Open {
			m.currentHost = fmt.Sprintf("%s:%d", msg.Host, msg.Port)
			m.currentPort = msg.Port
		}

		m.scanProgress[msg.Host]++

		if m.state == StateScanning {
			return m, m.spinner.Tick
		}

	case scanDoneMsg:
		m.scanning = false
		m.state = StateResults
		if msg.Err != nil {
			m.error = msg.Err
		}
		m.scanResults = msg.Results

		if m.cancel != nil {
			m.cancel()
			m.cancel = nil
		}

		return m, nil
	}

	if m.state == StatePortConfig {
		cmd := m.updateInputs(msg)
		return m, cmd
	}

	return m, nil
}

func (m *UIPortModel) updateInputs(msg tea.Msg) tea.Cmd {
	var cmds []tea.Cmd
	for i := range m.inputs {
		m.inputs[i], _ = m.inputs[i].Update(msg)
	}
	return tea.Batch(cmds...)
}

func (m UIPortModel) View() string {
	if m.quitting {
		return "Returning to main menu...\n"
	}

	var sb strings.Builder

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#5F87AF")).
		Padding(1, 2).
		Width(m.width - 4).
		Align(lipgloss.Center)

	logo := m.styles.TitleStyle.Render(`
 ██████╗  ██████╗ ██████╗ ████████╗     ███████╗ ██████╗ █████╗ ███╗   ██╗
  ██╔══██╗██╔═══██╗██╔══██╗╚══██╔══╝     ██╔════╝██╔════╝██╔══██╗████╗  ██║
  ██████╔╝██║   ██║██████╔╝   ██║        ███████╗██║     ███████║██╔██╗ ██║
  ██╔═══╝ ██║   ██║██╔══██╗   ██║        ╚════██║██║     ██╔══██║██║╚██╗██║
  ██║     ╚██████╔╝██║  ██║   ██║        ███████║╚██████╗██║  ██║██║ ╚████║
  ╚═╝      ╚═════╝ ╚═╝  ╚═╝   ╚═╝        ╚══════╝ ╚═════╝╚═╝  ╚═╝╚═╝  ╚═══╝
    `)

	sb.WriteString(logo)
	sb.WriteString("\n")

	contentBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#5F5FAF")).
		Padding(1, 2).
		Width(m.width - 10)

	switch m.state {
	case StateHostSelection:
		sb.WriteString(boxStyle.Render(m.styles.SectionStyle.Render("Select hosts to scan")))
		sb.WriteString("\n\n")

		if len(m.hosts) == 0 {
			sb.WriteString(contentBox.Render(m.styles.WarningStyle.Render("No hosts found. Run a host scan first.")))
		} else {
			var hostsContent strings.Builder

			for i, host := range m.hosts {
				checkbox := "[ ]"
				if host.Selected {
					checkbox = "[x]"
				}
				if i == m.cursor {
					hostsContent.WriteString(fmt.Sprintf("> %s %s\n", checkbox, host.Host))
				} else {
					hostsContent.WriteString(fmt.Sprintf("  %s %s\n", checkbox, host.Host))
				}
			}

			selectAllCheckbox := "[ ]"
			if m.selectAll {
				selectAllCheckbox = "[x]"
			}

			if m.cursor == len(m.hosts) {
				hostsContent.WriteString(fmt.Sprintf("> %s Select All\n", selectAllCheckbox))
			} else {
				hostsContent.WriteString(fmt.Sprintf("  %s Select All\n", selectAllCheckbox))
			}

			sb.WriteString(contentBox.Render(hostsContent.String()))
		}

		sb.WriteString("\n\n")
		sb.WriteString(m.styles.HelpStyle.Render("↑/↓: Navigate • Space: Toggle selection • Enter: Continue • Esc: Back"))

	case StatePortConfig:
		sb.WriteString(boxStyle.Render(m.styles.SectionStyle.Render("Configure Port Scan")))
		sb.WriteString("\n\n")

		var inputsContent strings.Builder

		var selectedHosts []string
		for _, h := range m.hosts {
			if h.Selected {
				selectedHosts = append(selectedHosts, h.Host)
			}
		}

		inputsContent.WriteString(m.styles.SuccessStyle.Render(fmt.Sprintf("Selected %d hosts:\n", len(selectedHosts))))

		maxHosts := 5
		for i, h := range selectedHosts {
			if i >= maxHosts && len(selectedHosts) > maxHosts {
				inputsContent.WriteString(fmt.Sprintf("... and %d more\n", len(selectedHosts)-maxHosts))
				break
			}
			inputsContent.WriteString(fmt.Sprintf("• %s\n", h))
		}
		inputsContent.WriteString("\n")

		for i, input := range m.inputs {
			inputsContent.WriteString(input.View())
			if i < len(m.inputs)-1 {
				inputsContent.WriteString("\n\n")
			}
		}

		inputsContent.WriteString("\n\n")

		commonPortsCheckbox := "[ ]"
		if m.useCommonPorts {
			commonPortsCheckbox = "[x]"
		}

		checkboxStyle := m.styles.ItemStyle
		if m.focusIndex == len(m.inputs) {
			checkboxStyle = m.styles.SelectedItemStyle
			inputsContent.WriteString(checkboxStyle.Render(fmt.Sprintf("> %s Use Common Ports (overrides port range)", commonPortsCheckbox)))
		} else {
			inputsContent.WriteString(checkboxStyle.Render(fmt.Sprintf("  %s Use Common Ports (overrides port range)", commonPortsCheckbox)))
		}

		if m.useCommonPorts {
			inputsContent.WriteString("\n\n")
			inputsContent.WriteString(m.styles.SectionStyle.Render("Common ports include: 21, 22, 23, 25, 53, 80, 443, 3306, 3389, 8080, etc."))
		}

		sb.WriteString(contentBox.Render(inputsContent.String()))
		sb.WriteString("\n\n")
		sb.WriteString(m.styles.HelpStyle.Render("Tab: Switch fields • Space: Toggle checkbox • Enter: Start scan • Esc: Back"))

	case StateScanning:
		sb.WriteString(boxStyle.Render(m.styles.SectionStyle.Render("Port Scanning in Progress")))
		sb.WriteString("\n\n")

		var scanContent strings.Builder
		scanContent.WriteString(fmt.Sprintf("%s Scanning ports...\n\n", m.spinner.View()))

		if m.useCommonPorts {
			scanContent.WriteString(m.styles.SuccessStyle.Render("Scanning common ports only\n\n"))
		}

		if m.currentHost != "" {
			scanContent.WriteString(m.styles.SuccessStyle.Render(fmt.Sprintf("Found open port: %s\n", m.currentHost)))
		}

		for host, count := range m.scanProgress {
			scanContent.WriteString(fmt.Sprintf("Scanned %d ports on %s\n", count, host))
		}

		sb.WriteString(contentBox.Render(scanContent.String()))
		sb.WriteString("\n\n")
		sb.WriteString(m.styles.HelpStyle.Render("Esc: Stop scan"))

	case StateResults:
		sb.WriteString(boxStyle.Render(m.styles.SectionStyle.Render("Port Scan Results")))
		sb.WriteString("\n\n")

		var resultsContent strings.Builder

		if m.error != nil {
			resultsContent.WriteString(m.styles.ErrorStyle.Render(fmt.Sprintf("Error: %v\n\n", m.error)))
		}

		totalOpenPorts := 0
		for host, ports := range m.scanResults {
			openPortCount := len(ports)
			totalOpenPorts += openPortCount

			resultsContent.WriteString(m.styles.HostFoundStyle.Render(fmt.Sprintf("Host: %s\n", host)))

			if openPortCount > 0 {
				resultsContent.WriteString(m.styles.SuccessStyle.Render(fmt.Sprintf("  %d open ports:\n", openPortCount)))

				var portsStr strings.Builder
				for i, port := range ports {
					portsStr.WriteString(fmt.Sprintf("  %5d", port))
					if (i+1)%5 == 0 {
						portsStr.WriteString("\n")
					}
				}
				if openPortCount%5 != 0 {
					portsStr.WriteString("\n")
				}

				resultsContent.WriteString(portsStr.String())
			} else {
				resultsContent.WriteString(m.styles.WarningStyle.Render("  No open ports found\n"))
			}

			resultsContent.WriteString("\n")
		}

		resultsContent.WriteString(m.styles.SectionStyle.Render(fmt.Sprintf("Total: %d open ports found across %d hosts",
			totalOpenPorts, len(m.scanResults))))

		sb.WriteString(contentBox.Render(resultsContent.String()))
		sb.WriteString("\n\n")
		sb.WriteString(m.styles.HelpStyle.Render("Enter: Return to main menu"))
	}

	return sb.String()
}
