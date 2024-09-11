package main

import (
	"fmt"
	"io"
	"os"
	"stinote/consts"
	"stinote/window"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
)

type AppState struct {
	fileName string
}

func main() {
	fmt.Println("Stinote is a simple sticky note application")

	myApp := app.NewWithID("stinote")
	myWindow := myApp.NewWindow("Stinote")

	// SplashWindow is a window borderless
	drv := fyne.CurrentApp().Driver()
	if drv, ok := drv.(desktop.Driver); ok {
		myWindow = drv.CreateSplashWindow()
		myWindow.Resize(fyne.NewSize(consts.WindowWidth, consts.WindowHeight))
	}

	textArea := widget.NewMultiLineEntry()
	textArea.Wrapping = fyne.TextWrapWord
	textArea.TextStyle = fyne.TextStyle{Monospace: true}

	// Initialize app state
	state := &AppState{}

	myApp.Lifecycle().SetOnEnteredForeground(func() {
		window.SetWindowAlwaysOnTop(myWindow)
		window.SetWindowOnTopRightCorner(myWindow)
	})

	myWindow.SetContent(textArea)

	myWindow.SetMainMenu(
		fyne.NewMainMenu(
			fyne.NewMenu("File",
				fyne.NewMenuItem("Open", func() {
					state.openFile(myWindow, textArea)
				}),
				fyne.NewMenuItem("Save", func() {
					state.saveFile(myWindow, textArea, false)
				}),
				fyne.NewMenuItem("Save As", func() {
					state.saveFile(myWindow, textArea, true)
				}),
			)))

	myWindow.ShowAndRun()
}

// openFile displays a file dialog for opening files and loads the selected file into the text area.
func (state *AppState) openFile(win fyne.Window, textArea *widget.Entry) {
	dialog.NewFileOpen(func(file fyne.URIReadCloser, err error) {
		if err == nil && file != nil {
			data, err := io.ReadAll(file)
			if err == nil {
				textArea.SetText(string(data))
			}
			state.fileName = file.URI().Path()
		}
	}, win).Show()
}

// saveFile saves the content of the text area to a file. If saveAs is true, it displays a "Save As" dialog.
func (state *AppState) saveFile(win fyne.Window, textArea *widget.Entry, saveAs bool) {
	saveAsDialog := dialog.NewFileSave(
		func(file fyne.URIWriteCloser, err error) {
			if err == nil && file != nil {
				err := os.WriteFile(file.URI().Path(), []byte(textArea.Text), 0644)
				if err != nil {
					dialog.ShowError(err, win)
					return
				}
				state.fileName = file.URI().Path()
				win.SetTitle(fmt.Sprintf("Stinote - %s", state.fileName))
			}
		}, win)

	if saveAs || state.fileName == "" {
		saveAsDialog.Show()
	} else {
		err := os.WriteFile(state.fileName, []byte(textArea.Text), 0644)
		if err != nil {
			dialog.ShowError(err, win)
			return
		}
	}
}
