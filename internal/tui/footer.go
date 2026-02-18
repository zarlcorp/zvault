package tui

import "github.com/zarlcorp/core/pkg/zstyle"

// renderFooter returns context-sensitive keybinding help for the current view.
func renderFooter(id viewID, width int) string {
	_ = width
	return zstyle.RenderFooter(helpFor(id))
}

// helpFor returns the keybinding entries for a given view.
func helpFor(id viewID) []zstyle.HelpPair {
	switch id {
	case viewPassword:
		return []zstyle.HelpPair{
			{Key: "enter", Desc: "submit"},
			{Key: "tab", Desc: "next field"},
			{Key: "ctrl+c", Desc: "quit"},
		}
	case viewMenu:
		return []zstyle.HelpPair{
			{Key: "enter", Desc: "select"},
			{Key: "q", Desc: "quit"},
		}
	case viewSecretList:
		return []zstyle.HelpPair{
			{Key: "enter", Desc: "open"},
			{Key: "n", Desc: "new"},
			{Key: "d", Desc: "delete"},
			{Key: "/", Desc: "search"},
			{Key: "tab", Desc: "filter"},
			{Key: "esc", Desc: "back"},
		}
	case viewSecretDetail:
		return []zstyle.HelpPair{
			{Key: "c", Desc: "copy"},
			{Key: "s", Desc: "show/hide"},
			{Key: "e", Desc: "edit"},
			{Key: "d", Desc: "delete"},
			{Key: "esc", Desc: "back"},
		}
	case viewSecretForm:
		return []zstyle.HelpPair{
			{Key: "tab", Desc: "next"},
			{Key: "shift+tab", Desc: "prev"},
			{Key: "ctrl+s", Desc: "save"},
			{Key: "esc", Desc: "cancel"},
		}
	case viewTaskList:
		return []zstyle.HelpPair{
			{Key: "enter", Desc: "detail"},
			{Key: "n", Desc: "new"},
			{Key: "space", Desc: "done"},
			{Key: "d", Desc: "delete"},
			{Key: "x", Desc: "clear"},
			{Key: "tab", Desc: "filter"},
			{Key: "esc", Desc: "back"},
		}
	case viewTaskDetail:
		return []zstyle.HelpPair{
			{Key: "e", Desc: "edit"},
			{Key: "space", Desc: "done"},
			{Key: "d", Desc: "delete"},
			{Key: "esc", Desc: "back"},
		}
	case viewTaskForm:
		return []zstyle.HelpPair{
			{Key: "tab", Desc: "next field"},
			{Key: "ctrl+s", Desc: "save"},
			{Key: "esc", Desc: "cancel"},
		}
	default:
		return []zstyle.HelpPair{
			{Key: "q", Desc: "quit"},
		}
	}
}
