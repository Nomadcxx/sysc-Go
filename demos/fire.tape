# VHS tape file for generating fire.gif
# Install vhs: go install github.com/charmbracelet/vhs@latest
# Generate: vhs demos/fire.tape

Output assets/fire.gif

Set FontSize 14
Set Width 1200
Set Height 600
Set Theme "Dracula"

# Dracula theme
Type "./syscgo -effect fire -theme dracula -duration 3"
Enter
Sleep 3.2s
Ctrl+C

# Gruvbox theme
Type "./syscgo -effect fire -theme gruvbox -duration 3"
Enter
Sleep 3.2s
Ctrl+C

# Nord theme
Type "./syscgo -effect fire -theme nord -duration 3"
Enter
Sleep 3.2s
Ctrl+C

# Tokyo Night theme
Type "./syscgo -effect fire -theme tokyo-night -duration 3"
Enter
Sleep 3.2s
