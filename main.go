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

// Extend TextArea
type CustomTextArea struct {
	widget.Entry
}

var (
	state    *AppState
	myWindow fyne.Window
	textArea *CustomTextArea
)

// Extend NewMultiLineEntry
func CustomNewMultiLineEntry() *CustomTextArea {
	e := &CustomTextArea{}
	e.ExtendBaseWidget(e)
	e.MultiLine = true
	e.Wrapping = fyne.TextWrap(fyne.TextTruncateClip)
	return e
}

// Apply custom shortcut on local widget
func (c *CustomTextArea) TypedShortcut(s fyne.Shortcut) {
	if _, ok := s.(*desktop.CustomShortcut); !ok {
		c.Entry.TypedShortcut(s)
		return
	} else if ok {
		t := s.(*desktop.CustomShortcut)
		if t.Modifier == fyne.KeyModifierControl {
			switch t.KeyName {
			case fyne.KeyS:
				state.saveFile(myWindow, textArea, false)
			}
		}

	}
}

func main() {
	fmt.Println("Stinote is a simple sticky note application")

	myApp := app.NewWithID("stinote")
	myWindow = myApp.NewWindow("Stinote")

	// SplashWindow is a window borderless
	drv := fyne.CurrentApp().Driver()
	if drv, ok := drv.(desktop.Driver); ok {
		myWindow = drv.CreateSplashWindow()
		myWindow.Resize(fyne.NewSize(consts.WindowWidth, consts.WindowHeight))
	}

	textArea = CustomNewMultiLineEntry()
	textArea.Wrapping = fyne.TextWrapWord
	textArea.TextStyle = fyne.TextStyle{Monospace: true}

	// Initialize app state
	state = &AppState{}

	myApp.Lifecycle().SetOnEnteredForeground(func() {
		window.SetWindowAlwaysOnTop(myWindow)
		window.SetWindowOnTopRightCorner(myWindow)
	})

	// Apply custom shortcut on global window
	ctrls := &desktop.CustomShortcut{
		Modifier: fyne.KeyModifierControl,
		KeyName:  fyne.KeyS,
	}
	myWindow.Canvas().AddShortcut(ctrls, func(shortcut fyne.Shortcut) {
		state.saveFile(myWindow, textArea, false)
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
func (state *AppState) openFile(win fyne.Window, textArea *CustomTextArea) {
	fileDialog := dialog.NewFileOpen(func(file fyne.URIReadCloser, err error) {
		if err == nil && file != nil {
			data, err := io.ReadAll(file)
			if err == nil {
				textArea.SetText(string(data))
			}
			state.fileName = file.URI().Path()
		}
		// Put back original size
		win.Resize(fyne.NewSize(consts.WindowWidth, consts.WindowHeight))
		window.SetWindowOnTopRightCorner(win)
	}, win)

	// Resize the file dialog to be of a reasonable size
	fileDialog.Resize(fyne.NewSize(800, 600))

	// The file dialog maximum size is bound to the window size
	win.Resize(fyne.NewSize(800, 600))

	fileDialog.Show()
}

// saveFile saves the content of the text area to a file. If saveAs is true, it displays a "Save As" dialog.
func (state *AppState) saveFile(win fyne.Window, textArea *CustomTextArea, saveAs bool) {
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
			// Put back original size
			win.Resize(fyne.NewSize(consts.WindowWidth, consts.WindowHeight))
			window.SetWindowOnTopRightCorner(win)
		}, win)

	if saveAs || state.fileName == "" {
		// Resize the file dialog to be of a reasonable size
		saveAsDialog.Resize(fyne.NewSize(800, 600))
		// The file dialog maximum size is bound to the window size
		win.Resize(fyne.NewSize(800, 600))

		saveAsDialog.Show()
	} else {
		err := os.WriteFile(state.fileName, []byte(textArea.Text), 0644)
		if err != nil {
			dialog.ShowError(err, win)
			return
		}
	}
}
