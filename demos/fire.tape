# VHS tape file for generating fire.gif
# Install vhs: go install github.com/charmbracelet/vhs@latest
# Generate: vhs demos/fire.tape

Output assets/fire.gif

Set FontSize 14
Set Width 1200
Set Height 600
Set Theme "Dracula"

Type "./syscgo -effect fire -theme dracula -duration 5"
Enter
Sleep 5.5s
