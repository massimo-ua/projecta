package ui

import (
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"gitlab.com/massimo-ua/projecta/gui/core"
)

const (
	inputWidth = 300
)

func AuthPage(auth core.AuthProvider, window fyne.Window, onSuccess func()) fyne.CanvasObject {
	usernameEntry := widget.NewEntry()
	usernameEntry.SetPlaceHolder("Username")
	wrappedUsernameEntry := container.NewGridWrap(fyne.NewSize(inputWidth, usernameEntry.MinSize().Height), usernameEntry)
	passwordEntry := widget.NewPasswordEntry()
	passwordEntry.SetPlaceHolder("Password")
	wrapppedPasswordEntry := container.NewGridWrap(fyne.NewSize(inputWidth, passwordEntry.MinSize().Height), passwordEntry)

	loginButton := widget.NewButton("Login", func() {
		err := auth.Login(usernameEntry.Text, passwordEntry.Text)
		if err != nil {
			dialog.ShowError(fmt.Errorf("failed to complete authorisation"), window)
			return
		}
		onSuccess()
	})

	form := container.NewVBox(
		wrappedUsernameEntry,
		wrapppedPasswordEntry,
		loginButton,
	)

	return container.NewCenter(form)
}
