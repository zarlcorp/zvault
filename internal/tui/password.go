package tui

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/zarlcorp/core/pkg/zstyle"
	"github.com/zarlcorp/zvault/internal/vault"
)

// passwordField identifies which input is focused.
type passwordField int

const (
	fieldPassword passwordField = iota
	fieldConfirm
)

// passwordModel handles vault unlock and creation.
type passwordModel struct {
	password textinput.Model
	confirm  textinput.Model
	focused  passwordField
	firstRun bool
	err      string
	vaultDir string
	width    int
	height   int
}

func newPasswordModel(vaultDir string) passwordModel {
	pw := textinput.New()
	pw.Placeholder = "master password"
	pw.EchoMode = textinput.EchoPassword
	pw.EchoCharacter = '•'
	pw.Focus()
	pw.PromptStyle = lipgloss.NewStyle().Foreground(zstyle.ZvaultAccent)
	pw.TextStyle = lipgloss.NewStyle().Foreground(zstyle.Text)

	cf := textinput.New()
	cf.Placeholder = "confirm password"
	cf.EchoMode = textinput.EchoPassword
	cf.EchoCharacter = '•'
	cf.PromptStyle = lipgloss.NewStyle().Foreground(zstyle.ZvaultAccent)
	cf.TextStyle = lipgloss.NewStyle().Foreground(zstyle.Text)

	firstRun := !vaultDirExists(vaultDir)

	return passwordModel{
		password: pw,
		confirm:  cf,
		focused:  fieldPassword,
		firstRun: firstRun,
		vaultDir: vaultDir,
	}
}

func vaultDirExists(dir string) bool {
	_, err := os.Stat(dir)
	return err == nil
}

func (m passwordModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m passwordModel) Update(msg tea.Msg) (passwordModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// clear error on any key
		m.err = ""

		switch {
		case key.Matches(msg, zstyle.KeyEnter):
			return m.submit()
		case key.Matches(msg, zstyle.KeyTab):
			if m.firstRun {
				return m.nextField(), nil
			}
		}

	case errMsg:
		m.err = msg.err.Error()
		return m, nil
	}

	return m.updateInputs(msg)
}

func (m passwordModel) submit() (passwordModel, tea.Cmd) {
	pw := m.password.Value()
	if pw == "" {
		m.err = "password cannot be empty"
		return m, nil
	}

	if m.firstRun {
		if m.focused == fieldPassword {
			// move to confirm field
			return m.nextField(), nil
		}
		if pw != m.confirm.Value() {
			m.err = "passwords do not match"
			m.confirm.SetValue("")
			return m, nil
		}
	}

	dir := m.vaultDir
	return m, openVaultCmd(dir, pw)
}

func (m passwordModel) nextField() passwordModel {
	if m.focused == fieldPassword {
		m.focused = fieldConfirm
		m.password.Blur()
		m.confirm.Focus()
	} else {
		m.focused = fieldPassword
		m.confirm.Blur()
		m.password.Focus()
	}
	return m
}

func (m passwordModel) updateInputs(msg tea.Msg) (passwordModel, tea.Cmd) {
	var cmd tea.Cmd
	if m.focused == fieldPassword {
		m.password, cmd = m.password.Update(msg)
	} else {
		m.confirm, cmd = m.confirm.Update(msg)
	}
	return m, cmd
}

func (m passwordModel) View() string {
	var b strings.Builder

	// logo
	indent := lipgloss.NewStyle().MarginLeft(2)
	logo := indent.Render(
		zstyle.StyledLogo(lipgloss.NewStyle().Foreground(zstyle.ZvaultAccent)),
	)
	toolName := indent.Render(zstyle.MutedText.Render("zvault"))
	b.WriteString(fmt.Sprintf("\n%s\n%s\n\n", logo, toolName))

	// title
	if m.firstRun {
		title := lipgloss.NewStyle().
			Foreground(zstyle.ZvaultAccent).
			Bold(true).
			Render("Create New Vault")
		b.WriteString(fmt.Sprintf("  %s\n\n", title))
		desc := zstyle.MutedText.Render("Choose a master password to protect your vault.")
		b.WriteString(fmt.Sprintf("  %s\n\n", desc))
	} else {
		title := lipgloss.NewStyle().
			Foreground(zstyle.ZvaultAccent).
			Bold(true).
			Render("Unlock Vault")
		b.WriteString(fmt.Sprintf("  %s\n\n", title))
		desc := zstyle.MutedText.Render("Enter your master password.")
		b.WriteString(fmt.Sprintf("  %s\n\n", desc))
	}

	// password field
	label := zstyle.Subtext1
	pwLabel := lipgloss.NewStyle().Foreground(label).Render("Password")
	b.WriteString(fmt.Sprintf("  %s\n", pwLabel))
	b.WriteString(fmt.Sprintf("  %s\n", m.password.View()))

	// confirm field (first-run only)
	if m.firstRun {
		b.WriteString("\n")
		cfLabel := lipgloss.NewStyle().Foreground(label).Render("Confirm")
		b.WriteString(fmt.Sprintf("  %s\n", cfLabel))
		b.WriteString(fmt.Sprintf("  %s\n", m.confirm.View()))
	}

	// error display
	if m.err != "" {
		b.WriteString("\n")
		errText := zstyle.StatusErr.Render("  " + m.err)
		b.WriteString(errText)
		b.WriteString("\n")
	}

	return b.String()
}

// openVaultCmd returns a command that tries to open the vault.
func openVaultCmd(dir, password string) tea.Cmd {
	return func() tea.Msg {
		v, err := vault.Open(dir, password)
		if err != nil {
			return errMsg{err: err}
		}
		return vaultOpenedMsg{vault: v}
	}
}
