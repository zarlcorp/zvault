```
   ________  ________  ________  ______    ________  ________  ________  ________
  ╱        ╲╱        ╲╱        ╲╱      ╲  ╱        ╲╱        ╲╱        ╲╱        ╲
 ╱-        ╱    /    ╱     /   ╱       ╱ ╱         ╱    /    ╱    /    ╱    /    ╱
╱        _╱         ╱        _╱       ╱_╱       --╱         ╱        _╱       __╱
╲________╱╲___╱____╱╲____╱___╱╲________╱╲________╱╲________╱╲____╱___╱╲______╱
```

# zvault

Encrypted local storage for secrets, keys, notes.

Manage your sensitive data with a simple terminal interface. zvault stores everything encrypted on your machine — your data, your device, your keys. No sync, no cloud, no servers. Download the binary and run.

## Install

### Via Homebrew

```bash
brew install zarlcorp/tap/zvault
```

### Via Go

```bash
go install github.com/zarlcorp/zvault/cmd/zvault@latest
```

## Usage

Start the interactive terminal interface:

```bash
zvault
```

View version information:

```bash
zvault version
```

Planned CLI subcommands (not yet implemented):

```bash
zvault get <path>          # retrieve a secret
zvault set <path>          # store a secret
zvault search <query>      # find secrets by name
```

## Development

Build the binary:

```bash
make build
```

Run tests:

```bash
make test
```

Run linter:

```bash
make lint
```

## Learn More

- [zarlcorp.github.io/zvault](https://zarlcorp.github.io/zvault) — documentation and install instructions
- [zarlcorp/core](https://github.com/zarlcorp/core) — shared packages for zarlcorp tools
- [MANIFESTO.md](https://github.com/zarlcorp/core/blob/main/MANIFESTO.md) — why we exist and what we're building

---

MIT License
