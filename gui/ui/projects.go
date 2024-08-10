package ui

import (
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"gitlab.com/massimo-ua/projecta/gui/core"
)

func ProjectsPage(projects core.ProjectRepository, window fyne.Window, onSelectProject func(projectID string)) fyne.CanvasObject {
	collection, _, err := projects.Find(0, 5)

	if err != nil {
		dialog.ShowError(fmt.Errorf("failed to fetch projects %s", err.Error()), window)
	}

	list := widget.NewList(
		func() int { return len(collection) }, // Replace with actual data
		func() fyne.CanvasObject {
			return Card(CardOptions{
				Title:    "",
				Content:  "",
				OnOkText: "View",
				OnOk: func() {
					// Do nothing
				},
			})
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			// Update the card with the data for each list item
			card := obj.(*fyne.Container)

			// Retrieve the VBox container inside the card (which contains the labels)
			cardContent := card.Objects[0].(*fyne.Container)

			// Update the title and content labels
			titleLabel := cardContent.Objects[0].(*widget.Label)
			contentLabel := cardContent.Objects[2].(*widget.Label)

			item := collection[id]
			titleLabel.SetText(item.Name)
			contentLabel.SetText(item.Description)
		},
	)

	list.OnSelected = func(id widget.ListItemID) {
		onSelectProject(collection[id].ID) // Replace with actual data
	}

	return container.NewGridWrap(fyne.NewSize(300, list.MinSize().Height), list)
}
