package easygate

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/mazznoer/colorgrad"
	"github.com/rivo/tview"
)

const splashText = "\n" +
	" _____                 ____       _       " + "\n" +
	"| ____|__ _ ___ _   _ / ___| __ _| |_ ___ \n" +
	"|  _| / _` / __| | | | |  _ / _` | __/ _ \\" + "\n" +
	"| |__| (_| \\__ | |_| | |_| | (_| | ||  __/" + "\n" +
	"|_____\\__,_|___/\\__, |\\____|\\__,_|\\__\\___|" + "\n" +
	"                |___/                     " + "\n"
const splashVersion = "⛩️ v0.1.0 ⛩️"

type Ui struct {
	tview      *tview.Application
	page       *tview.Pages
	configForm *tview.Form
	logView    *tview.TextView
	splash     *tview.TextView
}

type App struct {
	ui     Ui
	config *Config
	server *Server
}

func NewApp() *App {
	app := new(App)

	config, err := LoadConfig()
	if err != nil {
		panic(err)
	}
	app.config = config
	app.server = NewServer(app.config)

	app.ui.tview = tview.NewApplication()
	app.ui.tview.EnableMouse(true)

	app.ui.configForm = tview.NewForm().
		AddInputField("Proxy URL", app.config.Proxy.Url, 50, nil, app.makeChangeInput(&app.config.Proxy.Url)).
		AddInputField("User name", app.config.Proxy.UserName, 25, nil, app.makeChangeInput(&app.config.Proxy.UserName)).
		AddPasswordField("Password", app.config.Proxy.Password, 25, '*', app.makeChangeInput(&app.config.Proxy.Password)).
		AddButton("Connect", nil)

	app.ui.logView = tview.NewTextView().
		SetWrap(true).
		SetWordWrap(true).
		SetScrollable(true)

	base := tview.NewFlex().
		AddItem(app.ui.configForm, 0, 1, true).
		AddItem(app.ui.logView, 0, 1, false)

	app.ui.splash = tview.NewTextView().
		SetDynamicColors(true).
		SetRegions(true).
		SetTextAlign(tview.AlignCenter).
		SetChangedFunc(func() {
			app.ui.tview.Draw()
		})
	app.ui.page = tview.NewPages().
		AddPage("MAIN", base, true, true).
		AddPage("SPLASH", app.ui.splash, true, true)

	app.ui.tview.EnableMouse(true)
	app.ui.tview.SetRoot(app.ui.page, true).SetFocus(app.ui.page)

	return app
}

func (app App) Run() {
	// go app.opening(func() {
	// 	// Finish opning
	// 	app.ui.page.SwitchToPage("MAIN")
	// })
	// if err := app.ui.tview.Run(); err != nil {
	// 	panic(err)
	// }
	app.server.Start()
}

func (app App) opening(finishFunc func()) {
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

func (app App) makeChangeInput(ptr *string) func(string) {
	return func(text string) {
		*ptr = text
		app.config.Save()
	}
}
