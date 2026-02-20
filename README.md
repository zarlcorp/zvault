```
   ________  ________  ________  ______    ________  ________  ________  ________
  ╱        ╲╱        ╲╱        ╲╱      ╲  ╱        ╲╱        ╲╱        ╲╱        ╲
 ╱-        ╱    /    ╱     /   ╱       ╱ ╱         ╱    /    ╱    /    ╱    /    ╱
╱        _╱         ╱        _╱       ╱_╱       --╱         ╱        _╱       __╱
╲________╱╲___╱____╱╲____╱___╱╲________╱╲________╱╲________╱╲____╱___╱╲______╱
```

# zvault

Encrypted local storage for secrets, keys, notes.

Manage your sensitive data with a simple terminal interface. zvault stores everything encrypted on your machine -- your data, your device, your keys. No sync, no cloud, no servers. Download the binary and run.

## Install

### Via Homebrew

```bash
brew install zarlcorp/tap/zvault
```

### Via Go

```bash
go install github.com/zarlcorp/zvault/cmd/zvault@latest
```

## Quickstart

Launch the interactive TUI:

```bash
zvault
```

Or use the CLI directly:

```bash
# store a password
zvault secret store -t password -n github

# add a high-priority task
zvault task add -p h -d tomorrow "deploy to prod"

# list your secrets
zvault secret list

# search for a secret
zvault secret search github

# check version
zvault version
```

## Commands

### Secrets

```bash
zvault secret store -t <type> -n <name> [--tags tag1,tag2]
zvault secret get <id-or-name> [--show]
zvault secret list [-t <type>] [--tag <tag>]
zvault secret delete <id-or-name>
zvault secret search <query>
```

Secret types: `password`, `apikey`, `sshkey`, `note`.

Use `--show` with `get` to reveal sensitive values (masked by default).

### Tasks

```bash
zvault task add [-p <h|m|l>] [-d <date>] [--tags tag1,tag2] <title>
zvault task list [--pending] [--done] [-p <h|m|l>] [--tag <tag>]
zvault task done <id>
zvault task edit <id> <new title>
zvault task rm <id>
zvault task clear
```

Due date formats: `YYYY-MM-DD`, `today`, `tomorrow`, `next week`, `+3d`.

### Export

```bash
zvault export [--tasks] [--secrets] [--pending] [--done]
```

Exports vault data as markdown. Without flags, exports everything.

### Shell Completions

```bash
# bash
eval "$(zvault completion bash)"

# zsh
eval "$(zvault completion zsh)"

# fish
zvault completion fish | source
```

### Version

```bash
zvault version
```

## Configuration

zvault stores its encrypted vault in `~/.local/share/zvault/`. The vault is initialized on first use via the TUI.

Set `ZVAULT_PASSWORD` to skip interactive password prompts (useful for scripting).

Set `NO_COLOR` to disable colored output.

## Development

```bash
make build    # build binary to bin/zvault
make test     # run tests with race detector
make lint     # run golangci-lint
make clean    # remove build artifacts
```

Build with a specific version:

```bash
make build VERSION=1.0.0
```

## Learn More

- [zarlcorp.github.io/zvault](https://zarlcorp.github.io/zvault) -- documentation and install instructions
- [zarlcorp/core](https://github.com/zarlcorp/core) -- shared packages for zarlcorp tools
- [MANIFESTO.md](https://github.com/zarlcorp/core/blob/main/MANIFESTO.md) -- why we exist and what we're building

---

MIT License
