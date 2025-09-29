package ui

import (
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
	"github.com/paradoxe35/myreviser-go/internal/logger"
)

// NotificationManager handles user notifications
type NotificationManager struct {
	app fyne.App
}

// NewNotificationManager creates a new notification manager
func NewNotificationManager(app fyne.App) *NotificationManager {
	return &NotificationManager{
		app: app,
	}
}

// ShowSuccess shows a success notification
func (n *NotificationManager) ShowSuccess(title, content string) {
	notification := fyne.NewNotification(title, content)
	n.app.SendNotification(notification)
	logger.Info("Success notification shown", "title", title, "content", content)
}

// ShowError shows an error notification
func (n *NotificationManager) ShowError(title, content string) {
	notification := fyne.NewNotification(title, content)
	n.app.SendNotification(notification)
	logger.Error("Error notification shown", "title", title, "content", content)
}

// ShowInfo shows an info notification
func (n *NotificationManager) ShowInfo(title, content string) {
	notification := fyne.NewNotification(title, content)
	n.app.SendNotification(notification)
	logger.Info("Info notification shown", "title", title, "content", content)
}

// ShowToast shows a toast-style popup within the window
func ShowToast(window fyne.Window, message string, severity widget.Importance) {
	toast := widget.NewLabel(message)

	// Style based on severity
	switch severity {
	case widget.DangerImportance:
		toast = widget.NewLabelWithStyle(message, fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	case widget.WarningImportance:
		toast = widget.NewLabelWithStyle(message, fyne.TextAlignCenter, fyne.TextStyle{Italic: true})
	case widget.SuccessImportance:
		toast = widget.NewLabelWithStyle(message, fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	default:
		toast = widget.NewLabel(message)
	}

	// Create a popup
	popup := widget.NewModalPopUp(
		toast,
		window.Canvas(),
	)

	// Show the popup
	popup.Show()

	// Auto-hide after 3 seconds
	go func() {
		<-time.After(3 * time.Second)
		popup.Hide()
	}()
}