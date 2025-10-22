# VHS tape file for generating rain.gif
# Install vhs: go install github.com/charmbracelet/vhs@latest
# Generate: vhs demos/rain.tape

Output assets/rain.gif

Set FontSize 14
Set Width 1200
Set Height 600
Set Theme "TokyoNight"

# Tokyo Night theme
Type "./syscgo -effect rain -theme tokyo-night -duration 3"
Enter
Sleep 3.2s
Ctrl+C

# Solarized theme
Type "./syscgo -effect rain -theme solarized -duration 3"
Enter
Sleep 3.2s
Ctrl+C

# Gruvbox theme
Type "./syscgo -effect rain -theme gruvbox -duration 3"
Enter
Sleep 3.2s
Ctrl+C

# Nord theme
Type "./syscgo -effect rain -theme nord -duration 3"
Enter
Sleep 3.2s
