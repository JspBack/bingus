package ping

import (
	"time"

	"github.com/jspback/bingus/bta/internal/ui"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
)

type PingResult struct {
	IP  string
	RTT time.Duration
}

type hostFoundMsg string

type scanDoneMsg struct {
	Hosts []string
	Err   error
}

type PingState int

type UIPingModel struct {
	state      PingState
	inputs     []textinput.Model
	focusIndex int
	spinner    spinner.Model
	scanResult []string
	quitting   bool
	scanning   bool
	error      error
	styles     *ui.Styles
	width      int
	height     int
}
