package cli

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"golang.org/x/term"

	"github.com/zarlcorp/zvault/internal/vault"
)

// Run dispatches the top-level CLI subcommand.
func Run(args []string, version string) {
	if len(args) == 0 {
		printUsage()
		os.Exit(1)
	}

	switch args[0] {
	case "version":
		fmt.Printf("zvault %s\n", version)
	case "secret":
		runSecret(args[1:])
	case "task":
		runTask(args[1:])
	case "export":
		runExport(args[1:])
	case "completion":
		runCompletion(args[1:])
	case "help", "--help", "-h":
		printUsage()
	default:
		errf("unknown command %q", args[0])
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Fprint(os.Stderr, `Usage: zvault <command> [args]

Commands:
  secret      manage secrets (store, get, list, delete, search)
  task        manage tasks (add, list, done, edit, rm, clear)
  export      export vault data as markdown
  completion  generate shell completions
  version     print version
  help        show this help

Run 'zvault <command> --help' for command-specific help.
`)
}

// openVault prompts for the master password and opens the vault.
// It reads from ZVAULT_PASSWORD env var first, then prompts interactively.
func openVault() *vault.Vault {
	dir := vault.DefaultDir()

	password := vault.PasswordFromEnv()
	if password == "" {
		password = promptPassword("vault password: ")
	}

	v, err := vault.Open(dir, password)
	if err != nil {
		if strings.Contains(err.Error(), "no such file") || strings.Contains(err.Error(), "does not exist") {
			errf("vault not found — run the TUI to initialize: zvault")
			os.Exit(1)
		}
		errf("open vault: %v", err)
		os.Exit(1)
	}
	return v
}

// promptPassword reads a password from the terminal with masked input.
func promptPassword(prompt string) string {
	fmt.Fprint(os.Stderr, prompt)
	b, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Fprintln(os.Stderr) // newline after masked input
	if err != nil {
		errf("read password: %v", err)
		os.Exit(1)
	}
	return string(b)
}

// promptLine reads a single line from stdin with a prompt.
func promptLine(prompt string) string {
	fmt.Fprint(os.Stderr, muted(prompt))
	scanner := bufio.NewScanner(os.Stdin)
	if scanner.Scan() {
		return strings.TrimSpace(scanner.Text())
	}
	return ""
}

// promptConfirm asks a y/n question and returns true for yes.
func promptConfirm(prompt string) bool {
	fmt.Fprint(os.Stderr, yellow(prompt)+" [y/N] ")
	scanner := bufio.NewScanner(os.Stdin)
	if scanner.Scan() {
		ans := strings.TrimSpace(strings.ToLower(scanner.Text()))
		return ans == "y" || ans == "yes"
	}
	return false
}

// readStdin reads all of stdin if it's piped (not a terminal).
func readStdin() (string, bool) {
	info, err := os.Stdin.Stat()
	if err != nil {
		return "", false
	}
	if (info.Mode() & os.ModeCharDevice) != 0 {
		return "", false // terminal, not piped
	}
	b, err := io.ReadAll(os.Stdin)
	if err != nil {
		return "", false
	}
	return strings.TrimSpace(string(b)), true
}

// parseDate parses ISO date (2026-03-01) or relative date strings.
func parseDate(s string) (time.Time, error) {
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	switch strings.ToLower(s) {
	case "today":
		return today, nil
	case "tomorrow":
		return today.AddDate(0, 0, 1), nil
	case "next week":
		return today.AddDate(0, 0, 7), nil
	}

	// +Nd relative format
	if strings.HasPrefix(s, "+") && strings.HasSuffix(s, "d") {
		ds := s[1 : len(s)-1]
		var days int
		if _, err := fmt.Sscanf(ds, "%d", &days); err == nil {
			return today.AddDate(0, 0, days), nil
		}
	}

	// ISO format
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid date %q (use YYYY-MM-DD, tomorrow, next week, +Nd)", s)
	}
	return t, nil
}

// parseTags splits a comma-separated tag string.
func parseTags(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	var tags []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			tags = append(tags, p)
		}
	}
	return tags
}

// parseIDs splits a comma-separated ID string.
func parseIDs(s string) []string {
	parts := strings.Split(s, ",")
	var ids []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			ids = append(ids, p)
		}
	}
	return ids
}

// formatDueDate formats a due date relative to today.
func formatDueDate(t *time.Time) string {
	if t == nil {
		return ""
	}

	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	due := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
	diff := int(due.Sub(today).Hours() / 24)

	switch {
	case diff < 0:
		return red(fmt.Sprintf("overdue %dd", -diff))
	case diff == 0:
		return red("today")
	case diff == 1:
		return yellow("tomorrow")
	case diff <= 7:
		return yellow(fmt.Sprintf("in %dd", diff))
	default:
		return t.Format("2006-01-02")
	}
}

// hasFlag checks if a flag is present in args.
func hasFlag(args []string, flag string) bool {
	for _, a := range args {
		if a == flag {
			return true
		}
	}
	return false
}

// flagValue returns the value of a flag (e.g., -t password returns "password").
// Returns empty string if not found.
func flagValue(args []string, flag string) string {
	for i, a := range args {
		if a == flag && i+1 < len(args) {
			return args[i+1]
		}
	}
	return ""
}

// stripFlags removes known flag-value pairs and bare flags from args, returning positional args.
func stripFlags(args []string, valueFlags []string, bareFlags []string) []string {
	vf := make(map[string]bool, len(valueFlags))
	for _, f := range valueFlags {
		vf[f] = true
	}
	bf := make(map[string]bool, len(bareFlags))
	for _, f := range bareFlags {
		bf[f] = true
	}

	var pos []string
	skip := false
	for _, a := range args {
		if skip {
			skip = false
			continue
		}
		if vf[a] {
			skip = true
			continue
		}
		if bf[a] {
			continue
		}
		pos = append(pos, a)
	}
	return pos
}
