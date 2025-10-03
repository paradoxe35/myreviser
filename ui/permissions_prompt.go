package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/paradoxe35/myreviser/internal/permissions"
)

type permissionPrompt struct {
	root                   *fyne.Container
	infoLabel              *widget.Label
	accessibilitySection   fyne.CanvasObject
	inputMonitoringSection fyne.CanvasObject
	dividerAboveList       *widget.Separator
	dividerBelowList       *widget.Separator
	restartRow             fyne.CanvasObject
	restartNotice          *widget.Label
}

func newPermissionPrompt() *permissionPrompt {
	title := widget.NewLabelWithStyle("macOS Permissions Required", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

	info := widget.NewLabel("Grant these permissions so MyReviser can listen for shortcuts and automate edits.")
	info.Wrapping = fyne.TextWrapWord

	accessibilityButton := newPermissionButton("Open Settings", func() {
		permissions.RequestPermission(permissions.Accessibility)
		permissions.OpenPreference(permissions.Accessibility)
	})

	accessibilitySection := buildPermissionRow(
		"Accessibility",
		"Needed to control the keyboard and clipboard during revisions.",
		accessibilityButton,
	)

	inputMonitoringButton := newPermissionButton("Open Settings", func() {
		permissions.OpenPreference(permissions.InputMonitoring)
	})

	inputMonitoringSection := buildPermissionRow(
		"Input Monitoring",
		"Allows MyReviser to detect global shortcuts while running in the background.",
		inputMonitoringButton,
	)

	restartNotice := widget.NewLabelWithStyle(
		"Permissions granted. Quit and reopen MyReviser to finish applying them.",
		fyne.TextAlignLeading,
		fyne.TextStyle{Bold: true},
	)
	restartNotice.Wrapping = fyne.TextWrapWord

	restartRow := container.NewHBox(widget.NewIcon(theme.ConfirmIcon()), restartNotice)
	restartRow.Hide()

	dividerAbove := widget.NewSeparator()
	dividerBelow := widget.NewSeparator()
	dividerBelow.Hide()

	body := container.NewVBox(
		title,
		info,
		dividerAbove,
		accessibilitySection,
		inputMonitoringSection,
		dividerBelow,
		restartRow,
	)

	root := container.NewPadded(body)

	return &permissionPrompt{
		root:                   root,
		infoLabel:              info,
		accessibilitySection:   accessibilitySection,
		inputMonitoringSection: inputMonitoringSection,
		dividerAboveList:       dividerAbove,
		dividerBelowList:       dividerBelow,
		restartRow:             restartRow,
		restartNotice:          restartNotice,
	}
}

func (p *permissionPrompt) canvasObject() fyne.CanvasObject {
	return p.root
}

func (p *permissionPrompt) update(state permissions.State, showRestart bool) {
	pendingAccessibility := !state.AccessibilityGranted
	pendingInput := !state.InputMonitoringGranted
	hasPending := pendingAccessibility || pendingInput

	if pendingAccessibility {
		p.accessibilitySection.Show()
	} else {
		p.accessibilitySection.Hide()
	}

	if pendingInput {
		p.inputMonitoringSection.Show()
	} else {
		p.inputMonitoringSection.Hide()
	}

	switch {
	case hasPending:
		p.infoLabel.SetText("Grant these permissions so MyReviser can listen for shortcuts and automate edits.")
		p.dividerAboveList.Show()
		p.dividerBelowList.Hide()
		p.restartRow.Hide()
		p.root.Show()
	case showRestart:
		p.infoLabel.SetText("All permissions are enabled. Please restart MyReviser to complete the update.")
		p.dividerAboveList.Hide()
		p.dividerBelowList.Show()
		p.restartRow.Show()
		p.root.Show()
	default:
		p.restartRow.Hide()
		p.root.Hide()
	}

	p.root.Refresh()
}

func buildPermissionRow(title, description string, button *widget.Button) fyne.CanvasObject {
	titleLabel := widget.NewLabelWithStyle(title, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	descriptionLabel := widget.NewLabel(description)
	descriptionLabel.Wrapping = fyne.TextWrapWord

	textColumn := container.NewVBox(titleLabel, descriptionLabel)
	textColumn.Resize(fyne.NewSize(380, textColumn.MinSize().Height))

	row := container.NewBorder(nil, nil, textColumn, button, nil)
	return container.NewVBox(row)
}

func newPermissionButton(label string, tapped func()) *widget.Button {
	btn := widget.NewButtonWithIcon(label, theme.SettingsIcon(), tapped)
	btn.Importance = widget.MediumImportance
	btn.IconPlacement = widget.ButtonIconLeadingText
	return btn
}
