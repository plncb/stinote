package main

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func main() {
	fmt.Println("Stinote is a simple sticky note application")

	myApp := app.New()
	myWindow := myApp.NewWindow("Stinote")
	myWindow.Resize(fyne.NewSize(320, 285))

	textArea := widget.NewMultiLineEntry()
	textArea.Wrapping = fyne.TextWrapWord
	textArea.TextStyle = fyne.TextStyle{Monospace: true}

	myWindow.SetContent(container.NewVBox(textArea))

	myWindow.ShowAndRun()
}
