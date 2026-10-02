package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Theme colors - Monochrome (ASCII style)
var (
	BgBase       = lipgloss.Color("#1a1a1a")
	Primary      = lipgloss.Color("#ffffff")
	Secondary    = lipgloss.Color("#cccccc")
	Accent       = lipgloss.Color("#ffffff")
	FgPrimary    = lipgloss.Color("#ffffff")
	FgSecondary  = lipgloss.Color("#cccccc")
	FgMuted      = lipgloss.Color("#666666")
	ErrorColor   = lipgloss.Color("#ffffff")
	WarningColor = lipgloss.Color("#888888")
)

// Styles
var (
	checkMark   = lipgloss.NewStyle().Foreground(Accent).SetString("[OK]")
	failMark    = lipgloss.NewStyle().Foreground(ErrorColor).SetString("[FAIL]")
	skipMark    = lipgloss.NewStyle().Foreground(WarningColor).SetString("[SKIP]")
	headerStyle = lipgloss.NewStyle().Foreground(Primary).Bold(true)
)

type installStep int

const (
	stepWelcome installStep = iota
	stepInstalling
	stepComplete
)

type taskStatus int

const (
	statusPending taskStatus = iota
	statusRunning
	statusComplete
	statusFailed
	statusSkipped
)

type installTask struct {
	name        string
	description string
	execute     func(*model) error
	optional    bool
	status      taskStatus
}

type model struct {
	step             installStep
	tasks            []installTask
	currentTaskIndex int
	width            int
	height           int
	spinner          spinner.Model
	errors           []string
	uninstallMode    bool
	selectedOption   int // 0 = Install, 1 = Uninstall
}

type taskCompleteMsg struct {
	index   int
	success bool
	error   string
}

func newModel() model {
	s := spinner.New()
	s.Style = lipgloss.NewStyle().Foreground(Secondary)
	s.Spinner = spinner.Dot

	return model{
		step:             stepWelcome,
		currentTaskIndex: -1,
		spinner:          s,
		errors:           []string{},
		selectedOption:   0,
	}
}

func (m model) Init() tea.Cmd {
	return m.spinner.Tick
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			// Allow exit from any step except during installation
			if m.step != stepInstalling {
				return m, tea.Quit
			}
		case "up", "k":
			if m.step == stepWelcome && m.selectedOption > 0 {
				m.selectedOption--
			}
		case "down", "j":
			if m.step == stepWelcome && m.selectedOption < 1 {
				m.selectedOption++
			}
		case "enter":
			if m.step == stepWelcome {
				m.uninstallMode = m.selectedOption == 1
				m.initTasks()
				m.step = stepInstalling
				m.currentTaskIndex = 0
				m.tasks[0].status = statusRunning
				return m, tea.Batch(
					m.spinner.Tick,
					executeTask(0, &m),
				)
			} else if m.step == stepComplete {
				return m, tea.Quit
			}
		}

	case taskCompleteMsg:
		// Update task status
		if msg.success {
			m.tasks[msg.index].status = statusComplete
		} else {
			if m.tasks[msg.index].optional {
				m.tasks[msg.index].status = statusSkipped
				m.errors = append(m.errors, fmt.Sprintf("%s (skipped): %s", m.tasks[msg.index].name, msg.error))
			} else {
				m.tasks[msg.index].status = statusFailed
				m.errors = append(m.errors, fmt.Sprintf("%s: %s", m.tasks[msg.index].name, msg.error))
				m.step = stepComplete
				return m, nil
			}
		}

		// Move to next task
		m.currentTaskIndex++
		if m.currentTaskIndex >= len(m.tasks) {
			m.step = stepComplete
			return m, nil
		}

		// Start next task
		m.tasks[m.currentTaskIndex].status = statusRunning
		return m, executeTask(m.currentTaskIndex, &m)

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m *model) initTasks() {
	if m.uninstallMode {
		m.tasks = []installTask{
			{name: "Check privileges", description: "Checking root access", execute: checkPrivileges, status: statusPending},
			{name: "Remove syscgo", description: "Removing /usr/local/bin/syscgo", execute: removeSyscgoBinary, status: statusPending},
			{name: "Remove syscgo-tui", description: "Removing /usr/local/bin/syscgo-tui", execute: removeTuiBinary, status: statusPending},
			{name: "Remove assets", description: "Removing /usr/local/share/syscgo", execute: removeAssets, status: statusPending},
		}
	} else {
		m.tasks = []installTask{
			{name: "Check privileges", description: "Checking root access", execute: checkPrivileges, status: statusPending},
			{name: "Build syscgo", description: "Building syscgo binary", execute: buildBinary, status: statusPending},
			{name: "Build syscgo-tui", description: "Building syscgo-tui binary", execute: buildTuiBinary, status: statusPending},
			{name: "Install assets", description: "Installing assets to /usr/local/share/syscgo", execute: installAssets, status: statusPending},
			{name: "Install syscgo", description: "Installing syscgo to /usr/local/bin", execute: installBinary, status: statusPending},
			{name: "Install syscgo-tui", description: "Installing syscgo-tui to /usr/local/bin", execute: installTuiBinary, status: statusPending},
		}
	}
}

func (m model) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	var content strings.Builder

	// ASCII Header - render as a single block to avoid lipgloss line-by-line padding issues
	header := `▄▀▀▀▀ █   █ ▄▀▀▀▀ ▄▀▀▀▀       ▄▀▀▀▀ ▄▀▀▀▄    ▄▀    ▄▀
 ▀▀▀▄ ▀▀▀▀█  ▀▀▀▄ █     ▀▀▀▀▀ █ ▀▀█ █   █  ▄▀    ▄▀
▀▀▀▀  ▀▀▀▀▀ ▀▀▀▀   ▀▀▀▀        ▀▀▀   ▀▀▀  ▀     ▀
             /// SEE YOU SPACE COWBOY//               `

	content.WriteString(headerStyle.Render(header))
	content.WriteString("\n\n")

	// Main content based on step
	var mainContent string
	switch m.step {
	case stepWelcome:
		mainContent = m.renderWelcome()
	case stepInstalling:
		mainContent = m.renderInstalling()
	case stepComplete:
		mainContent = m.renderComplete()
	}

	// Wrap in border
	mainStyle := lipgloss.NewStyle().
		Padding(1, 2).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(Primary).
		Width(m.width - 4)
	content.WriteString(mainStyle.Render(mainContent))
	content.WriteString("\n")

	// Help text
	helpText := m.getHelpText()
	if helpText != "" {
		helpStyle := lipgloss.NewStyle().
			Foreground(FgMuted).
			Italic(true).
			Align(lipgloss.Center)
		content.WriteString("\n" + helpStyle.Render(helpText))
	}

	// Wrap everything in background with centering
	bgStyle := lipgloss.NewStyle().
		Background(BgBase).
		Foreground(FgPrimary).
		Width(m.width).
		Height(m.height).
		Align(lipgloss.Center, lipgloss.Top)

	return bgStyle.Render(content.String())
}

func (m model) renderWelcome() string {
	var b strings.Builder

	b.WriteString("Select an option:\n\n")

	// Install option
	installPrefix := "  "
	if m.selectedOption == 0 {
		installPrefix = lipgloss.NewStyle().Foreground(Primary).Render("▸ ")
	}
	b.WriteString(installPrefix + "Install syscgo\n")
	b.WriteString("    Builds binary and installs system-wide to /usr/local/bin\n\n")

	// Uninstall option
	uninstallPrefix := "  "
	if m.selectedOption == 1 {
		uninstallPrefix = lipgloss.NewStyle().Foreground(Primary).Render("▸ ")
	}
	b.WriteString(uninstallPrefix + "Uninstall syscgo\n")
	b.WriteString("    Removes syscgo from your system\n\n")

	b.WriteString(lipgloss.NewStyle().Foreground(FgMuted).Render("Requires root privileges"))

	return b.String()
}

func (m model) renderInstalling() string {
	var b strings.Builder

	// Render all tasks with their current status
	for i, task := range m.tasks {
		var line string
		switch task.status {
		case statusPending:
			line = lipgloss.NewStyle().Foreground(FgMuted).Render("  " + task.name)
		case statusRunning:
			line = m.spinner.View() + " " + lipgloss.NewStyle().Foreground(Secondary).Render(task.description)
		case statusComplete:
			line = checkMark.String() + " " + task.name
		case statusFailed:
			line = failMark.String() + " " + task.name
		case statusSkipped:
			line = skipMark.String() + " " + task.name
		}

		b.WriteString(line)
		if i < len(m.tasks)-1 {
			b.WriteString("\n")
		}
	}

	// Show errors at bottom if any
	if len(m.errors) > 0 {
		b.WriteString("\n\n")
		for _, err := range m.errors {
			b.WriteString(lipgloss.NewStyle().Foreground(WarningColor).Render(err))
			b.WriteString("\n")
		}
	}

	return b.String()
}

func (m model) renderComplete() string {
	var b strings.Builder

	if len(m.errors) > 0 {
		// Installation failed
		b.WriteString(lipgloss.NewStyle().Foreground(ErrorColor).Bold(true).Render("Installation failed"))
		b.WriteString("\n\n")
		b.WriteString(lipgloss.NewStyle().Foreground(FgSecondary).Render("Errors encountered:"))
		b.WriteString("\n")
		for _, err := range m.errors {
			b.WriteString(lipgloss.NewStyle().Foreground(WarningColor).Render("• " + err))
			b.WriteString("\n")
		}
	} else {
		// Installation succeeded
		if m.uninstallMode {
			b.WriteString(lipgloss.NewStyle().Foreground(Accent).Bold(true).Render("✓ Uninstallation complete!"))
			b.WriteString("\n\n")
			b.WriteString(lipgloss.NewStyle().Foreground(FgSecondary).Render("syscgo and syscgo-tui have been removed from your system."))
		} else {
			b.WriteString(lipgloss.NewStyle().Foreground(Accent).Bold(true).Render("✓ Installation complete!"))
			b.WriteString("\n\n")
			b.WriteString(lipgloss.NewStyle().Foreground(FgSecondary).Render("Installed binaries:"))
			b.WriteString("\n")
			b.WriteString(lipgloss.NewStyle().Foreground(Accent).Render("  • /usr/local/bin/syscgo"))
			b.WriteString("\n")
			b.WriteString(lipgloss.NewStyle().Foreground(Accent).Render("  • /usr/local/bin/syscgo-tui"))
			b.WriteString("\n\n")
			b.WriteString(lipgloss.NewStyle().Foreground(FgSecondary).Render("Installed assets:"))
			b.WriteString("\n")
			b.WriteString(lipgloss.NewStyle().Foreground(Accent).Render("  • /usr/local/share/syscgo/"))
			b.WriteString("\n\n")
			b.WriteString(lipgloss.NewStyle().Foreground(FgSecondary).Render("Try them out:"))
			b.WriteString("\n")
			b.WriteString(lipgloss.NewStyle().Foreground(Accent).Render("  syscgo -effect fire -theme dracula"))
			b.WriteString("\n")
			b.WriteString(lipgloss.NewStyle().Foreground(Accent).Render("  syscgo -effect aquarium -theme nord"))
			b.WriteString("\n")
			b.WriteString(lipgloss.NewStyle().Foreground(Accent).Render("  syscgo-tui"))
			b.WriteString("\n\n")
			b.WriteString(lipgloss.NewStyle().Foreground(FgMuted).Render("Launch syscgo-tui for an interactive TUI to browse and select animations!"))
		}
	}

	return b.String()
}

func (m model) getHelpText() string {
	switch m.step {
	case stepWelcome:
		return "↑/↓: Navigate  •  Enter: Continue  •  Q/Ctrl+C: Quit"
	case stepComplete:
		return "Enter: Exit  •  Q/Ctrl+C: Quit"
	default:
		return "Q/Ctrl+C: Cancel"
	}
}

func executeTask(index int, m *model) tea.Cmd {
	return func() tea.Msg {
		// Simulate work delay for visibility
		time.Sleep(200 * time.Millisecond)

		err := m.tasks[index].execute(m)

		if err != nil {
			return taskCompleteMsg{
				index:   index,
				success: false,
				error:   err.Error(),
			}
		}

		return taskCompleteMsg{
			index:   index,
			success: true,
		}
	}
}

// Task functions

func checkPrivileges(m *model) error {
	if os.Geteuid() != 0 {
		return fmt.Errorf("installer must be run with sudo or as root")
	}
	return nil
}

func buildBinary(m *model) error {
	cmd := exec.Command("go", "build", "-o", "syscgo", "./cmd/syscgo")
	cmd.Dir = getProjectRoot()
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("build failed: %s", string(output))
	}
	return nil
}

func buildTuiBinary(m *model) error {
	cmd := exec.Command("go", "build", "-o", "syscgo-tui", "./cmd/syscgo-tui")
	cmd.Dir = getProjectRoot()
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("build failed: %s", string(output))
	}
	return nil
}

func installAssets(m *model) error {
	projectRoot := getProjectRoot()
	srcPath := filepath.Join(projectRoot, "assets")
	if err := installAssetFiles(srcPath, "/usr/local/share/syscgo"); err != nil {
		return fmt.Errorf("failed to copy assets: %v", err)
	}
	return nil
}

// installAssetFiles copies repo assets into the system share layout used by the
// CLI and TUI: text files under <share>/assets and BIT fonts under <share>/fonts.
func installAssetFiles(srcAssets, shareRoot string) error {
	entries, err := os.ReadDir(srcAssets)
	if err != nil {
		return err
	}

	assetsDst := filepath.Join(shareRoot, "assets")
	if err := os.MkdirAll(assetsDst, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %v", err)
	}

	for _, entry := range entries {
		src := filepath.Join(srcAssets, entry.Name())
		if entry.IsDir() {
			// Fonts are discovered at the share root, not under assets/.
			if entry.Name() == "fonts" {
				if err := copyDir(src, filepath.Join(shareRoot, "fonts")); err != nil {
					return err
				}
			}
			continue
		}
		if err := copyFile(src, filepath.Join(assetsDst, entry.Name())); err != nil {
			return err
		}
	}

	return nil
}

func installBinary(m *model) error {
	projectRoot := getProjectRoot()
	srcPath := filepath.Join(projectRoot, "syscgo")
	dstPath := "/usr/local/bin/syscgo"

	// Read the source file
	data, err := os.ReadFile(srcPath)
	if err != nil {
		return fmt.Errorf("failed to read binary: %v", err)
	}

	// Write to destination
	err = os.WriteFile(dstPath, data, 0755)
	if err != nil {
		return fmt.Errorf("failed to install binary: %v", err)
	}

	return nil
}

func installTuiBinary(m *model) error {
	projectRoot := getProjectRoot()
	srcPath := filepath.Join(projectRoot, "syscgo-tui")
	dstPath := "/usr/local/bin/syscgo-tui"

	// Read the source file
	data, err := os.ReadFile(srcPath)
	if err != nil {
		return fmt.Errorf("failed to read binary: %v", err)
	}

	// Write to destination
	err = os.WriteFile(dstPath, data, 0755)
	if err != nil {
		return fmt.Errorf("failed to install binary: %v", err)
	}

	return nil
}

func removeSyscgoBinary(m *model) error {
	err := os.Remove("/usr/local/bin/syscgo")
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove binary: %v", err)
	}
	return nil
}

func removeTuiBinary(m *model) error {
	err := os.Remove("/usr/local/bin/syscgo-tui")
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove binary: %v", err)
	}
	return nil
}

func removeAssets(m *model) error {
	err := os.RemoveAll("/usr/local/share/syscgo")
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove assets: %v", err)
	}
	return nil
}

// copyDir recursively copies a directory
func copyDir(src, dst string) error {
	// Get properties of source dir
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	// Create destination directory
	err = os.MkdirAll(dst, srcInfo.Mode())
	if err != nil {
		return err
	}

	// Read source directory
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	// Copy each entry
	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			// Recursively copy subdirectory
			err = copyDir(srcPath, dstPath)
			if err != nil {
				return err
			}
		} else {
			// Copy file
			err = copyFile(srcPath, dstPath)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// copyFile copies a single file
func copyFile(src, dst string) error {
	// Read source file
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}

	// Get source file permissions
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	// Write to destination
	err = os.WriteFile(dst, data, srcInfo.Mode())
	if err != nil {
		return err
	}

	return nil
}

func getProjectRoot() string {
	// Get the directory where the installer is located
	execPath, err := os.Executable()
	if err != nil {
		// Fallback to current directory
		return "."
	}

	// Go up from cmd/installer to project root
	root := filepath.Dir(filepath.Dir(filepath.Dir(execPath)))

	// Check if go.mod exists to verify this is the project root
	if _, err := os.Stat(filepath.Join(root, "go.mod")); err == nil {
		return root
	}

	// Fallback: try to find go.mod by walking up from current directory
	dir, _ := os.Getwd()
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	return "."
}

func main() {
	// Check if go is installed
	if _, err := exec.LookPath("go"); err != nil {
		fmt.Println("Error: Go is not installed or not in PATH")
		fmt.Println("Please install Go from https://golang.org/dl/")
		os.Exit(1)
	}

	p := tea.NewProgram(newModel(), tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}
