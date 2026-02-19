Output docs/demo.gif

Set Shell zsh
Set FontSize 14
Set Width 800
Set Height 500
Set Theme { "name": "Catppuccin Mocha", "black": "#45475a", "red": "#f38ba8", "green": "#a6e3a1", "yellow": "#f9e2af", "blue": "#89b4fa", "magenta": "#f5c2e7", "cyan": "#94e2d5", "white": "#bac2de", "brightBlack": "#585b70", "brightRed": "#f38ba8", "brightGreen": "#a6e3a1", "brightYellow": "#f9e2af", "brightBlue": "#89b4fa", "brightMagenta": "#f5c2e7", "brightCyan": "#94e2d5", "brightWhite": "#a6adc8", "background": "#1e1e2e", "foreground": "#cdd6f4", "selectionBackground": "#585b70", "cursorColor": "#f5e0dc" }
Set Padding 20
Set Framerate 20
Set TypingSpeed 75ms

# set up temp environment
Hide
Type "export XDG_DATA_HOME=$(mktemp -d) ZVAULT_PASSWORD=demo1234"
Enter
Sleep 500ms
Show

# show version
Type "zvault version"
Enter
Sleep 1.5s

# store a secret
Type "echo 'gh_abc123xyz' | zvault secret store -t apikey -n github"
Enter
Sleep 2s

# store another
Type "echo 's3cretP@ss!' | zvault secret store -t password -n aws-prod"
Enter
Sleep 2s

# list secrets
Type "zvault secret list"
Enter
Sleep 2s

# search
Type "zvault secret search github"
Enter
Sleep 2s

# add a task
Type "zvault task add -p h -d tomorrow 'deploy to prod'"
Enter
Sleep 2s

# list tasks
Type "zvault task list"
Enter
Sleep 2s

Sleep 1s
