package ui

import (
	"fmt"
	"regexp"
	"sync"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

var NonWhiteSpaceRegex = regexp.MustCompile(`\S`)

type UI struct {
	mx *sync.Mutex

	app   *tview.Application
	frame *tview.Frame
	log   *tview.TextView

	pageIndex int
	pages     []*Page
	pageView  *tview.Pages

	localPort  uint16
	remoteHost string
	remotePort uint16

	dataUp   uint64
	dataDown uint64
}

func New(localPort uint16, remoteHost string, remotePort uint16) *UI {
	app := tview.NewApplication()

	pageView := tview.NewPages()

	flex := tview.NewFlex().SetDirection(tview.FlexRow)
	flex.AddItem(pageView, 0, 3, true)

	log := tview.NewTextView()
	log.SetDynamicColors(true)
	log.SetBorder(true)
	log.SetBorderColor(tcell.ColorGrey)
	log.ScrollToEnd()
	flex.AddItem(log, 0, 1, false)

	frame := tview.NewFrame(flex)

	app.SetRoot(frame, true)

	ui := &UI{
		mx: &sync.Mutex{},

		app:   app,
		frame: frame,
		log:   log,

		pageIndex: -1,
		pageView:  pageView,
		pages:     []*Page{},

		localPort:  localPort,
		remoteHost: remoteHost,
		remotePort: remotePort,

		dataUp:   0,
		dataDown: 0,
	}

	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if ui.pageIndex == -1 {
			return event
		}

		if event.Key() == tcell.KeyCtrlQ {
			ui.pageIndex -= 1
			if ui.pageIndex < 0 {
				ui.pageIndex = len(ui.pages) - 1
			}

			ui.pageView.SwitchToPage(fmt.Sprintf("%v", ui.pageIndex))
		}

		if event.Key() == tcell.KeyCtrlE {
			ui.pageIndex += +1
			if ui.pageIndex >= len(ui.pages) {
				ui.pageIndex = 0
			}

			ui.pageView.SwitchToPage(fmt.Sprintf("%v", ui.pageIndex))
		}

		return event
	})

	ui.redraw()
	return ui
}

func (ui *UI) Run() error {
	return ui.app.Run()
}

func (ui *UI) WriteLog(msg string, args ...any) {
	ui.mx.Lock()
	defer ui.mx.Unlock()

	ui.log.Write([]byte(fmt.Sprintf(msg, args...)))
	ui.redraw()
	ui.app.Draw()
}

func (ui *UI) AddPage() int {
	ui.mx.Lock()
	defer ui.mx.Unlock()

	pageName := fmt.Sprintf("%v", len(ui.pages))
	page := NewPage(pageName, ui.app, ui.pageView)
	ui.pages = append(ui.pages, page)

	if ui.pageIndex == -1 {
		ui.pageIndex = 0
		ui.pageView.SwitchToPage(pageName)
	}

	ui.redraw()
	ui.app.Draw()

	return len(ui.pages) - 1
}

func (ui *UI) SetPageConnected(page int) {
	ui.mx.Lock()
	defer ui.mx.Unlock()

	p := ui.pages[page]
	p.active = true
	p.title.SetText(fmt.Sprintf("Connection %v (🟢 Connected)", page))

	ui.redraw()
	ui.app.Draw()
}

func (ui *UI) SetPageClosed(page int) {
	ui.mx.Lock()
	defer ui.mx.Unlock()

	p := ui.pages[page]
	p.closed = true
	p.active = false
	p.title.SetText(fmt.Sprintf("Connection %v (⚫ Disconnected)", page))

	ui.redraw()
	ui.app.Draw()
}

func (ui *UI) AddClientData(page int, data []byte) {
	ui.mx.Lock()
	defer ui.mx.Unlock()

	p := ui.pages[page]
	p.AddClientData(data)

	ui.dataUp += uint64(len(data))
	ui.redraw()
	ui.app.Draw()
}

func (ui *UI) AddServerData(page int, data []byte) {
	ui.mx.Lock()
	defer ui.mx.Unlock()

	p := ui.pages[page]
	p.AddServerData(data)

	ui.dataDown += uint64(len(data))
	ui.redraw()
	ui.app.Draw()
}

func (ui *UI) redraw() {
	ui.frame.Clear()
	pendingConnections, activeConnections, closedConnections := ui.getConnectionCounts()

	bytesText := fmt.Sprintf(
		"%v 👆    %v 👇        %v 🟠    %v 🟢    %v ⚫",
		formatBytes(ui.dataUp),
		formatBytes(ui.dataDown),
		pendingConnections,
		activeConnections,
		closedConnections)

	ui.frame.AddText("👀 [white::b]tcpwatch", true, tview.AlignLeft, tcell.ColorWhite)
	ui.frame.AddText("v0.1.1 @lukejoshuapark", true, tview.AlignLeft, tcell.ColorGrey)
	ui.frame.AddText(fmt.Sprintf("Port %v 👂", ui.localPort), true, tview.AlignRight, tcell.ColorWhite)
	ui.frame.AddText(fmt.Sprintf("%v:%v 🌐", ui.remoteHost, ui.remotePort), true, tview.AlignRight, tcell.ColorWhite)
	ui.frame.AddText("Next Conn. <CTRL-E>    Previous Conn. <CTRL-Q>    Quit <CTRL+C>", false, tview.AlignLeft, tcell.ColorGrey)
	ui.frame.AddText(bytesText, false, tview.AlignRight, tcell.ColorWhite)
}

func (ui *UI) getConnectionCounts() (int, int, int) {
	pendingConnections := 0
	activeConnections := 0
	closedConnections := 0

	for _, page := range ui.pages {
		if page.active {
			activeConnections++
			continue
		}

		if page.closed {
			closedConnections++
			continue
		}

		pendingConnections++
	}

	return pendingConnections, activeConnections, closedConnections
}
