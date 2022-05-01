package easygate

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/mazznoer/colorgrad"
	"github.com/rivo/tview"
)

const (
	splashText = "\n" +
		" _____                 ____       _       " + "\n" +
		"| ____|__ _ ___ _   _ / ___| __ _| |_ ___ \n" +
		"|  _| / _` / __| | | | |  _ / _` | __/ _ \\" + "\n" +
		"| |__| (_| \\__ | |_| | |_| | (_| | ||  __/" + "\n" +
		"|_____\\__,_|___/\\__, |\\____|\\__,_|\\__\\___|" + "\n" +
		"                |___/                     " + "\n"
	splashVersion   = "⛩️ v0.1.0 ⛩️"
	serverStatusOff = "Status: [#D3DEDC:#7C99AC:-]  OFF  [-:-:-]"
	serverStatusOn  = "Status: [#B4E197:#4E944F:-]  O N  [-:-:-]"
	maxLogViewCount = 3000
)

type Ui struct {
	tview         *tview.Application
	page          *tview.Pages
	logView       *tview.TextView
	serviceStatus *tview.TextView
	splash        *tview.TextView
}

type App struct {
	ui     Ui
	config *Config
	server *Server
	log    *Logger
}

func NewApp() *App {
	app := new(App)
	config, err := LoadConfig()
	if err != nil {
		panic(err)
	}
	app.log = GetLogger()
	app.log.SetLevel(GetLogLevelFromString(config.LogLevel))
	app.config = config
	app.server = NewServer(app.config)

	app.ui.tview = tview.NewApplication()
	app.ui.tview.EnableMouse(true)

	configForm := tview.NewForm().
		AddInputField("Proxy - URL", app.config.Proxy.Url, 50, nil, app.makeChangeInput(&app.config.Proxy.Url)).
		AddInputField("Proxy - User name", app.config.Proxy.UserName, 25, nil, app.makeChangeInput(&app.config.Proxy.UserName)).
		AddPasswordField("Proxy - Password", app.config.Proxy.Password, 25, '*', app.makeChangeInput(&app.config.Proxy.Password)).
		AddInputField("Service - Port", app.config.Serve.ListenPort, 15, nil, app.makeChangeInput(&app.config.Serve.ListenPort)).
		AddInputField("Service - Pac file path", app.config.Serve.PacFilePath, 75, nil, app.makeChangeInput(&app.config.Serve.PacFilePath))

	configFrame := tview.NewFrame(configForm).
		AddText("Configurations are automatically saved.", true, tview.AlignLeft, tcell.ColorDefault).
		AddText("You need restart app, If you changed configuration.", true, tview.AlignLeft, tcell.ColorDefault).
		AddText("ESC: Back | Tab: Move", false, tview.AlignLeft, tcell.ColorDefault)
	configFrame.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			app.ui.page.SwitchToPage("MAIN")
		} else if event.Key() == tcell.KeyRune {
		}
		return event
	})

	app.ui.serviceStatus = tview.NewTextView().SetDynamicColors(true).
		SetText(serverStatusOff)

	messageArea := tview.NewTextView().SetDynamicColors(true).SetText("").SetTextColor(tcell.ColorRed)

	app.ui.logView = tview.NewTextView().
		SetWrap(true).
		SetWordWrap(true).
		SetScrollable(true).
		SetChangedFunc(func() {
			if app.ui.logView.GetOriginalLineCount() > maxLogViewCount {
				app.ui.logView.SetText("log clear...\n")
			}
			app.ui.tview.Draw()
		})
	app.ui.logView.SetTextColor(tcell.NewHexColor(0xE8D0F2)).SetBackgroundColor(tcell.NewHexColor(0x554A59))

	flexArea := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(
			tview.NewFlex().
				AddItem(app.ui.serviceStatus, 20, 1, true).
				AddItem(messageArea, 0, 1, false), 1, 1, false).
		AddItem(app.ui.logView, 0, 1, true)
	mainFrame := tview.NewFrame(flexArea).
		AddText("ESC: Exit | Space: ON/OFF | F1: Configuration", false, tview.AlignLeft, tcell.ColorDefault)
	mainFrame.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		messageArea.SetText("")
		switch event.Key() {
		case tcell.KeyF1:
			if app.server.IsRunning() {
				messageArea.SetText("You need stop service!!")
				return event
			}
			app.ui.page.SwitchToPage("CONFIG")
		case tcell.KeyEscape:
			app.ui.tview.Stop()
		case tcell.KeyRune:
			switch event.Rune() {
			case ' ':
				if app.server.IsRunning() {
					app.server.Stop()
					app.ui.serviceStatus.SetText(serverStatusOff)
				} else {
					app.server.Start(&app.config.Serve)
					app.ui.serviceStatus.SetText(serverStatusOn)
				}
			}
		}
		return event
	})

	app.ui.splash = tview.NewTextView().
		SetDynamicColors(true).
		SetRegions(true).
		SetTextAlign(tview.AlignCenter).
		SetChangedFunc(func() {
			app.ui.tview.Draw()
		})
	app.ui.page = tview.NewPages().
		AddPage("MAIN", mainFrame, true, false).
		AddPage("CONFIG", configFrame, true, false).
		AddPage("SPLASH", app.ui.splash, true, true)

	app.ui.tview.EnableMouse(true)
	app.ui.tview.SetRoot(app.ui.page, true).SetFocus(app.ui.page)
	app.log.SetExternalWriter(app.ui.logView.Write)
	return app
}

func (app *App) Run() {
	go app.opening(func() {
		// Finish opning
		app.ui.tview.QueueUpdateDraw(func() {
			app.ui.page.SwitchToPage("MAIN")
			app.log.Info("Proxy: %s  / User: %s", app.config.Proxy.Url, app.config.Proxy.UserName)
			app.log.Info("Listen port: %s", app.config.Serve.ListenPort)
			app.log.Info("Pac: [%s]", app.config.Serve.PacFilePath)
		})
	})
	if err := app.ui.tview.Run(); err != nil {
		panic(err)
	}
}

func (app *App) opening(finishFunc func()) {
	grad, err := colorgrad.NewGradient().
		HtmlColors("#F27121", "#8A2387", "#E94057", "#F27121").
		Build()
	if err != nil {
		panic(err)
	}
	for offset := 0.0; offset < 1; offset += 0.01 {
		resultText := ""
		for _, line := range strings.Split(splashText, "\n") {
			lineLength := len(line)
			for i, c := range line {
				colorOffset := math.Abs(float64(i)/float64(lineLength) - offset)
				if colorOffset > 1 {
					colorOffset -= math.Floor(colorOffset)
				}
				resultText += "[" + grad.At(colorOffset).Hex() + "::]" + string(c) + "[-:-:-]"
			}
			resultText += "\n"
		}
		resultText += splashVersion
		app.ui.splash.Clear()
		fmt.Fprintf(app.ui.splash, "%s", resultText)
		time.Sleep(10 * time.Millisecond)
	}
	finishFunc()
}

func (app *App) makeChangeInput(ptr *string) func(string) {
	return func(text string) {
		*ptr = text
		app.config.Save()
	}
}
