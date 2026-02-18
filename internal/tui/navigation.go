package tui

import "github.com/zarlcorp/zvault/internal/vault"

// viewID identifies the active view.
type viewID int

const (
	viewPassword viewID = iota
	viewMenu
	viewSecretList
	viewSecretDetail
	viewSecretForm
	viewTaskList
	viewTaskDetail
	viewTaskForm
)

// viewTitle returns the display title for a view.
func viewTitle(id viewID) string {
	switch id {
	case viewPassword:
		return "unlock"
	case viewMenu:
		return "menu"
	case viewSecretList:
		return "secrets"
	case viewSecretDetail:
		return "secret"
	case viewSecretForm:
		return "edit secret"
	case viewTaskList:
		return "tasks"
	case viewTaskDetail:
		return "task"
	case viewTaskForm:
		return "edit task"
	default:
		return ""
	}
}

// parentView returns the logical parent for back-navigation.
func parentView(id viewID) viewID {
	switch id {
	case viewSecretList, viewTaskList:
		return viewMenu
	case viewSecretDetail, viewSecretForm:
		return viewSecretList
	case viewTaskDetail, viewTaskForm:
		return viewTaskList
	default:
		return viewMenu
	}
}

// navigateMsg requests a view transition.
type navigateMsg struct {
	view viewID
	data any // optional payload (e.g., secret ID for detail view)
}

// errMsg carries a transient error for display.
type errMsg struct {
	err error
}

// vaultOpenedMsg signals the vault was opened successfully.
type vaultOpenedMsg struct {
	vault *vault.Vault
}
