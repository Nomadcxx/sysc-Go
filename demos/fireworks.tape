# VHS tape file for generating fireworks.gif
# Install vhs: go install github.com/charmbracelet/vhs@latest
# Generate: vhs demos/fireworks.tape

Output assets/fireworks.gif

Set FontSize 14
Set Width 1200
Set Height 600
Set Theme "GruvboxDark"

# Gruvbox theme
Type "./syscgo -effect fireworks -theme gruvbox -duration 3"
Enter
Sleep 3.2s
Ctrl+C

# Tokyo Night theme
Type "./syscgo -effect fireworks -theme tokyo-night -duration 3"
Enter
Sleep 3.2s
Ctrl+C

# Nord theme
Type "./syscgo -effect fireworks -theme nord -duration 3"
Enter
Sleep 3.2s
Ctrl+C

# Dracula theme
Type "./syscgo -effect fireworks -theme dracula -duration 3"
Enter
Sleep 3.2s
