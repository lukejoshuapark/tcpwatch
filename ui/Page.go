package ui

import (
	"fmt"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type Page struct {
	active bool
	closed bool

	title          *tview.TextView
	clientUTF8View *tview.TextView
	serverUTF8View *tview.TextView
	clientHexView  *tview.TextView
	serverHexView  *tview.TextView
}

func NewPage(name string, app *tview.Application, pageView *tview.Pages) *Page {
	clientUTF8View := tview.NewTextView()
	clientUTF8View.ScrollToEnd()
	clientUTF8View.SetDynamicColors(true)
	clientUTF8View.SetBorder(true)
	clientUTF8View.SetBorderColor(tcell.ColorGrey)
	clientUTF8View.SetWordWrap(false)
	clientUTF8View.SetWrap(true)

	clientHexView := tview.NewTextView()
	clientHexView.ScrollToEnd()
	clientHexView.SetDynamicColors(true)
	clientHexView.SetBorder(true)
	clientHexView.SetBorderColor(tcell.ColorGrey)
	clientHexView.SetWordWrap(false)
	clientHexView.SetWrap(true)

	serverUTF8View := tview.NewTextView()
	serverUTF8View.ScrollToEnd()
	serverUTF8View.SetDynamicColors(true)
	serverUTF8View.SetBorder(true)
	serverUTF8View.SetBorderColor(tcell.ColorGrey)
	serverUTF8View.SetWordWrap(false)
	serverUTF8View.SetWrap(true)

	serverHexView := tview.NewTextView()
	serverHexView.ScrollToEnd()
	serverHexView.SetDynamicColors(true)
	serverHexView.SetBorder(true)
	serverHexView.SetBorderColor(tcell.ColorGrey)
	serverHexView.SetWordWrap(false)
	serverHexView.SetWrap(true)

	outerFlex := tview.NewFlex()
	outerFlex.SetDirection(tview.FlexRow)

	title := tview.NewTextView()
	title.SetText(fmt.Sprintf("Connection %v (🟠 Pending)", name))
	outerFlex.AddItem(title, 2, 1, false)

	innerFlex := tview.NewFlex()
	innerFlex.AddItem(clientUTF8View, 0, 1, true)
	innerFlex.AddItem(serverUTF8View, 0, 1, false)
	innerFlex.AddItem(clientHexView, 0, 1, false)
	innerFlex.AddItem(serverHexView, 0, 1, false)

	outerFlex.AddItem(innerFlex, 0, 1, true)

	currentlyFocused := 0
	focusChanged := func() {
		switch currentlyFocused {
		case 0:
			app.SetFocus(clientUTF8View)
		case 1:
			app.SetFocus(serverUTF8View)
		case 2:
			app.SetFocus(clientHexView)
		case 3:
			app.SetFocus(serverHexView)
		}
	}

	scrollChanged := func() {
		switch currentlyFocused {
		case 0:
			row, column := clientUTF8View.GetScrollOffset()
			serverUTF8View.ScrollTo(row, column)
		case 1:
			row, column := serverUTF8View.GetScrollOffset()
			clientUTF8View.ScrollTo(row, column)
		case 2:
			row, column := clientHexView.GetScrollOffset()
			serverHexView.ScrollTo(row, column)
		case 3:
			row, column := serverHexView.GetScrollOffset()
			clientHexView.ScrollTo(row, column)
		}
	}

	innerFlex.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyLeft {
			currentlyFocused--
			if currentlyFocused < 0 {
				currentlyFocused = 3
			}

			focusChanged()
		}

		if event.Key() == tcell.KeyRight {
			currentlyFocused++
			if currentlyFocused > 3 {
				currentlyFocused = 0
			}

			focusChanged()
		}

		if event.Key() == tcell.KeyUp || event.Key() == tcell.KeyDown || event.Key() == tcell.KeyPgUp || event.Key() == tcell.KeyPgDn {
			go func() {
				time.Sleep(time.Millisecond * 10)
				scrollChanged()
			}()
		}

		return event
	})

	pageView.AddPage(name, outerFlex, true, false)

	return &Page{
		active: false,
		closed: false,

		title:          title,
		clientUTF8View: clientUTF8View,
		serverUTF8View: serverUTF8View,
		clientHexView:  clientHexView,
		serverHexView:  serverHexView,
	}
}

func (page *Page) AddClientData(data []byte) {
	bwcu := page.clientUTF8View.BatchWriter()
	bwsu := page.serverUTF8View.BatchWriter()
	bwch := page.clientHexView.BatchWriter()
	bwsh := page.serverHexView.BatchWriter()

	defer func() {
		bwcu.Close()
		bwsu.Close()
		bwch.Close()
		bwsh.Close()
	}()

	hex := getCapitalizedSpacedHex(data)
	fmt.Fprintf(bwcu, "[green::-]%v\n", string(data))
	fmt.Fprintf(bwch, "[grey::b]%v\n", hex)
	fmt.Fprintf(bwsu, "%v\n", NonWhiteSpaceRegex.ReplaceAllString(string(data), " "))
	fmt.Fprintf(bwsh, "%v\n", NonWhiteSpaceRegex.ReplaceAllString(hex, " "))
}

func (page *Page) AddServerData(data []byte) {
	bwsu := page.serverUTF8View.BatchWriter()
	bwcu := page.clientUTF8View.BatchWriter()
	bwsh := page.serverHexView.BatchWriter()
	bwch := page.clientHexView.BatchWriter()

	defer func() {
		bwsu.Close()
		bwcu.Close()
		bwsh.Close()
		bwch.Close()
	}()

	hex := getCapitalizedSpacedHex(data)
	fmt.Fprintf(bwsu, "[yellow::-]%v\n", string(data))
	fmt.Fprintf(bwsh, "[grey::b]%v\n", hex)
	fmt.Fprintf(bwcu, "%v\n", NonWhiteSpaceRegex.ReplaceAllString(string(data), " "))
	fmt.Fprintf(bwch, "%v\n", NonWhiteSpaceRegex.ReplaceAllString(hex, " "))
}

func getCapitalizedSpacedHex(data []byte) string {
	var result string
	for i, b := range data {
		result += fmt.Sprintf("%02x", b)

		if i != len(data)-1 {
			result += " "
		}
	}

	return result
}

func formatBytes(bytes uint64) string {
	suffixes := []string{"B", "KB", "MB", "GB", "TB", "PB", "EB"}
	value := float64(bytes)
	ind := 0

	for value >= 1024 && ind < len(suffixes)-1 {
		value /= 1024
		ind++
	}

	if ind == 0 {
		return fmt.Sprintf("%v %s", uint64(value), suffixes[ind])
	}

	return fmt.Sprintf("%.2f %s", value, suffixes[ind])
}
