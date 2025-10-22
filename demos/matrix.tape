# VHS tape file for generating matrix.gif
# Install vhs: go install github.com/charmbracelet/vhs@latest
# Generate: vhs demos/matrix.tape

Output assets/matrix.gif

Set FontSize 14
Set Width 1200
Set Height 600
Set Theme "Nord"

Type "./syscgo -effect matrix -theme nord -duration 5"
Enter
Sleep 5.5s
