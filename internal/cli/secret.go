package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/zarlcorp/zvault/internal/secret"
	"github.com/zarlcorp/zvault/internal/vault"
)

func runSecret(args []string) {
	if len(args) == 0 {
		printSecretUsage()
		os.Exit(1)
	}

	switch args[0] {
	case "store":
		runSecretStore(args[1:])
	case "get":
		runSecretGet(args[1:])
	case "list", "ls":
		runSecretList(args[1:])
	case "delete", "rm":
		runSecretDelete(args[1:])
	case "search":
		runSecretSearch(args[1:])
	case "help", "--help", "-h":
		printSecretUsage()
	default:
		errf("unknown secret command %q", args[0])
		printSecretUsage()
		os.Exit(1)
	}
}

func printSecretUsage() {
	fmt.Fprint(os.Stderr, `Usage: zvault secret <command>

Commands:
  store   create a new secret
  get     retrieve a secret
  list    list secrets
  delete  delete a secret
  search  search secrets

Store flags:
  -t <type>         secret type: password, apikey, sshkey, note
  -n <name>         secret name
  --tags tag1,tag2  optional tags

Get flags:
  --show            reveal sensitive values (masked by default)

List flags:
  -t <type>         filter by type
  --tag <tag>       filter by tag
`)
}

func runSecretStore(args []string) {
	typ := flagValue(args, "-t")
	name := flagValue(args, "-n")
	tags := parseTags(flagValue(args, "--tags"))

	if typ == "" {
		errf("secret type required (-t password|apikey|sshkey|note)")
		os.Exit(1)
	}
	if name == "" {
		errf("secret name required (-n <name>)")
		os.Exit(1)
	}

	var sec secret.Secret

	var err error
	switch secret.Type(typ) {
	case secret.TypePassword:
		url := promptLine("url: ")
		username := promptLine("username: ")
		password := promptPassword("password: ")
		sec, err = secret.NewPassword(name, url, username, password)

	case secret.TypeAPIKey:
		service := promptLine("service: ")
		key := promptPassword("api key: ")
		sec, err = secret.NewAPIKey(name, service, key)

	case secret.TypeNote:
		content, piped := readStdin()
		if !piped {
			content = promptLine("content: ")
		}
		sec, err = secret.NewNote(name, content)

	case secret.TypeSSHKey:
		label := promptLine("label: ")
		fmt.Fprintln(os.Stderr, muted("enter private key (or path to file):"))
		privKey := promptLine("")
		fmt.Fprintln(os.Stderr, muted("enter public key (or path to file):"))
		pubKey := promptLine("")

		// try reading as file paths
		if data, e := os.ReadFile(privKey); e == nil {
			privKey = strings.TrimSpace(string(data))
		}
		if data, e := os.ReadFile(pubKey); e == nil {
			pubKey = strings.TrimSpace(string(data))
		}

		sec, err = secret.NewSSHKey(name, label, privKey, pubKey)

	default:
		errf("unknown secret type %q (use password, apikey, sshkey, note)", typ)
		os.Exit(1)
	}

	if err != nil {
		errf("create secret: %v", err)
		os.Exit(1)
	}

	sec.Tags = tags

	v := openVault()
	defer v.Close()

	if err := v.Secrets().Add(sec); err != nil {
		errf("store secret: %v", err)
		os.Exit(1)
	}

	fmt.Printf("%s %s stored\n", green(sec.ID), bold(sec.Name))
}

func runSecretGet(args []string) {
	show := hasFlag(args, "--show")
	pos := stripFlags(args, nil, []string{"--show"})

	if len(pos) == 0 {
		errf("secret ID or name required")
		os.Exit(1)
	}

	v := openVault()
	defer v.Close()

	sec, err := resolveSecret(v, pos[0])
	if err != nil {
		errf("%v", err)
		os.Exit(1)
	}

	printSecretDetail(sec, show)
}

func runSecretList(args []string) {
	typ := flagValue(args, "-t")
	tag := flagValue(args, "--tag")

	v := openVault()
	defer v.Close()

	all, err := v.Secrets().List()
	if err != nil {
		errf("list secrets: %v", err)
		os.Exit(1)
	}

	var filtered []secret.Secret
	for _, sec := range all {
		if typ != "" && string(sec.Type) != typ {
			continue
		}
		if tag != "" && !containsTag(sec.Tags, tag) {
			continue
		}
		filtered = append(filtered, sec)
	}

	if len(filtered) == 0 {
		fmt.Fprintln(os.Stderr, muted("no secrets found"))
		return
	}

	for _, sec := range filtered {
		printSecretRow(sec)
	}
}

func runSecretDelete(args []string) {
	if len(args) == 0 {
		errf("secret ID required")
		os.Exit(1)
	}

	v := openVault()
	defer v.Close()

	sec, err := resolveSecret(v, args[0])
	if err != nil {
		errf("%v", err)
		os.Exit(1)
	}

	if !promptConfirm(fmt.Sprintf("delete %q?", sec.Name)) {
		fmt.Fprintln(os.Stderr, "cancelled")
		return
	}

	if err := v.Secrets().Delete(sec.ID); err != nil {
		errf("delete secret: %v", err)
		os.Exit(1)
	}

	fmt.Printf("%s deleted\n", bold(sec.Name))
}

func runSecretSearch(args []string) {
	if len(args) == 0 {
		errf("search query required")
		os.Exit(1)
	}

	query := strings.Join(args, " ")

	v := openVault()
	defer v.Close()

	results, err := v.Secrets().Search(query)
	if err != nil {
		errf("search: %v", err)
		os.Exit(1)
	}

	if len(results) == 0 {
		fmt.Fprintln(os.Stderr, muted("no results"))
		return
	}

	for _, sec := range results {
		printSecretRow(sec)
	}
}

// resolveSecret finds a secret by ID, name, or ID prefix.
func resolveSecret(v *vault.Vault, idOrName string) (secret.Secret, error) {
	// try by exact ID first
	sec, err := v.Secrets().Get(idOrName)
	if err == nil {
		return sec, nil
	}

	// try by name (case-insensitive) or ID prefix
	all, err := v.Secrets().List()
	if err != nil {
		return secret.Secret{}, fmt.Errorf("list secrets: %w", err)
	}

	for _, s := range all {
		if strings.EqualFold(s.Name, idOrName) {
			return s, nil
		}
	}

	for _, s := range all {
		if strings.HasPrefix(s.ID, idOrName) {
			return s, nil
		}
	}

	return secret.Secret{}, fmt.Errorf("secret %q not found", idOrName)
}

func printSecretRow(sec secret.Secret) {
	tags := ""
	if len(sec.Tags) > 0 {
		var parts []string
		for _, t := range sec.Tags {
			parts = append(parts, blue("#"+t))
		}
		tags = "  " + strings.Join(parts, " ")
	}

	fmt.Printf("%-10s %-10s %s%s\n",
		muted(sec.ID[:8]),
		peach(string(sec.Type)),
		sec.Name,
		tags,
	)
}

func printSecretDetail(sec secret.Secret, show bool) {
	fmt.Printf("%s %s\n", boldMauve(sec.Name), muted(sec.ID))
	fmt.Printf("  %s %s\n", muted("type:"), peach(string(sec.Type)))

	if len(sec.Tags) > 0 {
		var parts []string
		for _, t := range sec.Tags {
			parts = append(parts, blue("#"+t))
		}
		fmt.Printf("  %s %s\n", muted("tags:"), strings.Join(parts, " "))
	}

	fmt.Printf("  %s %s\n", muted("created:"), sec.CreatedAt.Format("2006-01-02 15:04"))

	mask := "********"

	switch sec.Type {
	case secret.TypePassword:
		fmt.Printf("  %s %s\n", muted("url:"), sec.URL())
		fmt.Printf("  %s %s\n", muted("username:"), sec.Username())
		if show {
			fmt.Printf("  %s %s\n", muted("password:"), sec.Password())
		} else {
			fmt.Printf("  %s %s\n", muted("password:"), muted(mask))
		}

	case secret.TypeAPIKey:
		fmt.Printf("  %s %s\n", muted("service:"), sec.Service())
		if show {
			fmt.Printf("  %s %s\n", muted("key:"), sec.Key())
		} else {
			fmt.Printf("  %s %s\n", muted("key:"), muted(mask))
		}

	case secret.TypeSSHKey:
		fmt.Printf("  %s %s\n", muted("label:"), sec.Label())
		if show {
			fmt.Printf("  %s\n%s\n", muted("private key:"), sec.PrivateKey())
			fmt.Printf("  %s\n%s\n", muted("public key:"), sec.PublicKey())
		} else {
			fmt.Printf("  %s %s\n", muted("private key:"), muted(mask))
			fmt.Printf("  %s %s\n", muted("public key:"), muted(mask))
		}

	case secret.TypeNote:
		if show {
			fmt.Printf("  %s\n%s\n", muted("content:"), sec.Content())
		} else {
			fmt.Printf("  %s %s\n", muted("content:"), muted(mask))
		}
	}
}

func containsTag(tags []string, tag string) bool {
	for _, t := range tags {
		if t == tag {
			return true
		}
	}
	return false
}
