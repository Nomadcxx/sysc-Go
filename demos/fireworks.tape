# VHS tape file for generating fireworks.gif
# Install vhs: go install github.com/charmbracelet/vhs@latest
# Generate: vhs demos/fireworks.tape

Output assets/fireworks.gif

Set FontSize 14
Set Width 1200
Set Height 600
Set Theme "Gruvbox Dark"

Type "./syscgo -effect fireworks -theme gruvbox -duration 5"
Enter
Sleep 5.5s
