package cli

import (
	"fmt"
	"os"
)

func runCompletion(args []string) {
	if len(args) == 0 {
		errf("shell type required (bash, zsh, fish)")
		os.Exit(1)
	}

	switch args[0] {
	case "bash":
		fmt.Print(bashCompletion)
	case "zsh":
		fmt.Print(zshCompletion)
	case "fish":
		fmt.Print(fishCompletion)
	default:
		errf("unsupported shell %q (use bash, zsh, or fish)", args[0])
		os.Exit(1)
	}
}

const bashCompletion = `# zvault bash completion
_zvault() {
    local cur prev words cword
    _init_completion || return

    local commands="secret task export completion version help"
    local secret_cmds="store get list delete search"
    local task_cmds="add list ls done edit rm clear"
    local secret_types="password apikey sshkey note"
    local shells="bash zsh fish"
    local priorities="h m l"

    case "${cword}" in
        1)
            COMPREPLY=($(compgen -W "${commands}" -- "${cur}"))
            return
            ;;
        2)
            case "${prev}" in
                secret)
                    COMPREPLY=($(compgen -W "${secret_cmds}" -- "${cur}"))
                    return
                    ;;
                task)
                    COMPREPLY=($(compgen -W "${task_cmds}" -- "${cur}"))
                    return
                    ;;
                completion)
                    COMPREPLY=($(compgen -W "${shells}" -- "${cur}"))
                    return
                    ;;
            esac
            ;;
        *)
            case "${words[1]}" in
                secret)
                    case "${prev}" in
                        -t) COMPREPLY=($(compgen -W "${secret_types}" -- "${cur}")) ;;
                    esac
                    ;;
                task)
                    case "${prev}" in
                        -p) COMPREPLY=($(compgen -W "${priorities}" -- "${cur}")) ;;
                    esac
                    ;;
            esac
            ;;
    esac
}

complete -F _zvault zvault
`

const zshCompletion = `#compdef zvault

_zvault() {
    local -a commands secret_cmds task_cmds

    commands=(
        'secret:manage secrets'
        'task:manage tasks'
        'export:export vault data'
        'completion:generate shell completions'
        'version:print version'
        'help:show help'
    )

    secret_cmds=(
        'store:create a new secret'
        'get:retrieve a secret'
        'list:list secrets'
        'delete:delete a secret'
        'search:search secrets'
    )

    task_cmds=(
        'add:add a new task'
        'list:list tasks'
        'ls:list tasks'
        'done:mark task(s) complete'
        'edit:rename a task'
        'rm:delete task(s)'
        'clear:delete all completed tasks'
    )

    if (( CURRENT == 2 )); then
        _describe 'command' commands
        return
    fi

    case "${words[2]}" in
        secret)
            if (( CURRENT == 3 )); then
                _describe 'secret command' secret_cmds
                return
            fi
            case "${words[3]}" in
                store)
                    _arguments \
                        '-t[secret type]:type:(password apikey sshkey note)' \
                        '-n[secret name]:name:' \
                        '--tags[tags]:tags:'
                    ;;
                get)
                    _arguments '--show[reveal sensitive values]'
                    ;;
                list)
                    _arguments \
                        '-t[filter by type]:type:(password apikey sshkey note)' \
                        '--tag[filter by tag]:tag:'
                    ;;
            esac
            ;;
        task)
            if (( CURRENT == 3 )); then
                _describe 'task command' task_cmds
                return
            fi
            case "${words[3]}" in
                add)
                    _arguments \
                        '-p[priority]:priority:(h m l)' \
                        '-d[due date]:date:' \
                        '--tags[tags]:tags:'
                    ;;
                list)
                    _arguments \
                        '--pending[show only pending]' \
                        '--done[show only completed]' \
                        '-p[filter by priority]:priority:(h m l)' \
                        '--tag[filter by tag]:tag:'
                    ;;
            esac
            ;;
        completion)
            if (( CURRENT == 3 )); then
                _values 'shell' bash zsh fish
            fi
            ;;
        export)
            _arguments \
                '--tasks[export tasks]' \
                '--secrets[export secrets]' \
                '--pending[pending tasks only]' \
                '--done[completed tasks only]'
            ;;
    esac
}

_zvault "$@"
`

const fishCompletion = `# zvault fish completion

# top-level commands
complete -c zvault -n '__fish_use_subcommand' -a 'secret' -d 'manage secrets'
complete -c zvault -n '__fish_use_subcommand' -a 'task' -d 'manage tasks'
complete -c zvault -n '__fish_use_subcommand' -a 'export' -d 'export vault data'
complete -c zvault -n '__fish_use_subcommand' -a 'completion' -d 'generate shell completions'
complete -c zvault -n '__fish_use_subcommand' -a 'version' -d 'print version'
complete -c zvault -n '__fish_use_subcommand' -a 'help' -d 'show help'

# secret subcommands
complete -c zvault -n '__fish_seen_subcommand_from secret; and not __fish_seen_subcommand_from store get list delete search' -a 'store' -d 'create a new secret'
complete -c zvault -n '__fish_seen_subcommand_from secret; and not __fish_seen_subcommand_from store get list delete search' -a 'get' -d 'retrieve a secret'
complete -c zvault -n '__fish_seen_subcommand_from secret; and not __fish_seen_subcommand_from store get list delete search' -a 'list' -d 'list secrets'
complete -c zvault -n '__fish_seen_subcommand_from secret; and not __fish_seen_subcommand_from store get list delete search' -a 'delete' -d 'delete a secret'
complete -c zvault -n '__fish_seen_subcommand_from secret; and not __fish_seen_subcommand_from store get list delete search' -a 'search' -d 'search secrets'

# secret store flags
complete -c zvault -n '__fish_seen_subcommand_from store' -s t -d 'secret type' -xa 'password apikey sshkey note'
complete -c zvault -n '__fish_seen_subcommand_from store' -s n -d 'secret name'
complete -c zvault -n '__fish_seen_subcommand_from store' -l tags -d 'comma-separated tags'

# secret get flags
complete -c zvault -n '__fish_seen_subcommand_from get' -l show -d 'reveal sensitive values'

# secret list flags
complete -c zvault -n '__fish_seen_subcommand_from secret; and __fish_seen_subcommand_from list' -s t -d 'filter by type' -xa 'password apikey sshkey note'
complete -c zvault -n '__fish_seen_subcommand_from secret; and __fish_seen_subcommand_from list' -l tag -d 'filter by tag'

# task subcommands
complete -c zvault -n '__fish_seen_subcommand_from task; and not __fish_seen_subcommand_from add list ls done edit rm clear' -a 'add' -d 'add a new task'
complete -c zvault -n '__fish_seen_subcommand_from task; and not __fish_seen_subcommand_from add list ls done edit rm clear' -a 'list' -d 'list tasks'
complete -c zvault -n '__fish_seen_subcommand_from task; and not __fish_seen_subcommand_from add list ls done edit rm clear' -a 'ls' -d 'list tasks'
complete -c zvault -n '__fish_seen_subcommand_from task; and not __fish_seen_subcommand_from add list ls done edit rm clear' -a 'done' -d 'mark task(s) complete'
complete -c zvault -n '__fish_seen_subcommand_from task; and not __fish_seen_subcommand_from add list ls done edit rm clear' -a 'edit' -d 'rename a task'
complete -c zvault -n '__fish_seen_subcommand_from task; and not __fish_seen_subcommand_from add list ls done edit rm clear' -a 'rm' -d 'delete task(s)'
complete -c zvault -n '__fish_seen_subcommand_from task; and not __fish_seen_subcommand_from add list ls done edit rm clear' -a 'clear' -d 'delete all completed tasks'

# task add flags
complete -c zvault -n '__fish_seen_subcommand_from add' -s p -d 'priority' -xa 'h m l'
complete -c zvault -n '__fish_seen_subcommand_from add' -s d -d 'due date'
complete -c zvault -n '__fish_seen_subcommand_from add' -l tags -d 'comma-separated tags'

# task list flags
complete -c zvault -n '__fish_seen_subcommand_from task; and __fish_seen_subcommand_from list' -l pending -d 'show only pending'
complete -c zvault -n '__fish_seen_subcommand_from task; and __fish_seen_subcommand_from list' -l done -d 'show only completed'
complete -c zvault -n '__fish_seen_subcommand_from task; and __fish_seen_subcommand_from list' -s p -d 'filter by priority' -xa 'h m l'
complete -c zvault -n '__fish_seen_subcommand_from task; and __fish_seen_subcommand_from list' -l tag -d 'filter by tag'

# completion subcommands
complete -c zvault -n '__fish_seen_subcommand_from completion' -a 'bash zsh fish' -d 'shell type'

# export flags
complete -c zvault -n '__fish_seen_subcommand_from export' -l tasks -d 'export tasks'
complete -c zvault -n '__fish_seen_subcommand_from export' -l secrets -d 'export secrets'
complete -c zvault -n '__fish_seen_subcommand_from export' -l pending -d 'pending tasks only'
complete -c zvault -n '__fish_seen_subcommand_from export' -l done -d 'completed tasks only'
`
