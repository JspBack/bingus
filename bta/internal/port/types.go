package port

import (
	"context"

	"github.com/jspback/bingus/bta/internal/ui"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
)

type PortResult struct {
	Port  int
	Open  bool
	Error error
}

type PortState int

type HostItem struct {
	Host     string
	Selected bool
}

type portFoundMsg struct {
	Host string
	Port int
	Open bool
}

type scanDoneMsg struct {
	Results map[string][]int
	Err     error
}

type UIPortModel struct {
	state          PortState
	hosts          []HostItem
	cursor         int
	selectAll      bool
	inputs         []textinput.Model
	focusIndex     int
	spinner        spinner.Model
	scanResults    map[string][]int
	currentHost    string
	currentPort    int
	scanProgress   map[string]int
	quitting       bool
	scanning       bool
	error          error
	styles         *ui.Styles
	width          int
	height         int
	cancel         context.CancelFunc
	useCommonPorts bool
}
