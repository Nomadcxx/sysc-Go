# VHS tape file for generating rain.gif
# Install vhs: go install github.com/charmbracelet/vhs@latest
# Generate: vhs demos/rain.tape

Output assets/rain.gif

Set FontSize 14
Set Width 1200
Set Height 600
Set Theme "Tokyo Night"

Type "./syscgo -effect rain -theme tokyo-night -duration 5"
Enter
Sleep 5.5s
