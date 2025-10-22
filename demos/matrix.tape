# VHS tape file for generating matrix.gif
# Install vhs: go install github.com/charmbracelet/vhs@latest
# Generate: vhs demos/matrix.tape

Output assets/matrix.gif

Set FontSize 14
Set Width 1200
Set Height 600
Set Theme "nord"

# Nord theme
Type "./syscgo -effect matrix -theme nord -duration 3"
Enter
Sleep 3.2s
Ctrl+C

# Dracula theme
Type "./syscgo -effect matrix -theme dracula -duration 3"
Enter
Sleep 3.2s
Ctrl+C

# Catppuccin theme
Type "./syscgo -effect matrix -theme catppuccin -duration 3"
Enter
Sleep 3.2s
Ctrl+C

# Material theme
Type "./syscgo -effect matrix -theme material -duration 3"
Enter
Sleep 3.2s
