package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

type CardOptions struct {
	Title    string
	Content  string
	OnOk     func()
	OnOkText string
	OnCancel func()
}

func Card(opts CardOptions) fyne.CanvasObject {
	components := []fyne.CanvasObject{
		widget.NewLabelWithStyle(opts.Title, fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewSeparator(),
		widget.NewLabel(opts.Content),
	}

	var buttons []fyne.CanvasObject
	withOk := opts.OnOk != nil
	if withOk {
		// OK button
		if opts.OnOkText == "" {
			opts.OnOkText = "OK"
		}
		buttons = append(buttons, widget.NewButton(opts.OnOkText, opts.OnOk))
	}

	withCancel := opts.OnCancel != nil
	if withCancel {
		// Cancel button
		buttons = append(buttons, widget.NewButton("Cancel", opts.OnCancel))
	}

	withButtons := withOk || withCancel

	if withButtons {
		// Add a separator between the content and the buttons
		components = append(components, widget.NewSeparator())
		// Add the buttons to the components
		components = append(components, container.NewHBox(buttons...))
	}

	// Create a box for the card content
	contentBox := container.NewVBox(components...)

	// Return the card container directly
	return container.NewPadded(contentBox)
}
