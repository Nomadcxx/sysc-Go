package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
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
	ctx              context.Context
	cancel           context.CancelFunc
	cancelled        bool
}

type taskCompleteMsg struct {
	index     int
	success   bool
	error     string
	cancelled bool
}

func newModel(parent ...context.Context) model {
	s := spinner.New()
	s.Style = lipgloss.NewStyle().Foreground(Secondary)
	s.Spinner = spinner.Dot
	p := context.Background()
	if len(parent) > 0 && parent[0] != nil {
		p = parent[0]
	}
	ctx, cancel := context.WithCancel(p)

	return model{
		step:             stepWelcome,
		currentTaskIndex: -1,
		spinner:          s,
		errors:           []string{},
		selectedOption:   0,
		ctx:              ctx,
		cancel:           cancel,
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
		case "ctrl+c", "q", "Q":
			if m.step == stepInstalling {
				if m.cancelled {
					// Already cancelling and the in-flight task has not
					// returned. Some calls cannot be interrupted (os.RemoveAll,
					// a single large WriteFile), so honour the key and leave.
					return m, tea.Quit
				}
				// Stop builds immediately, but leave the program up until the
				// in-flight task returns. Quitting first would exit while
				// compilers are still running and could tear a file copy.
				m.cancelled = true
				m.stop()
				return m, nil
			}
			return m, tea.Quit
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
		if m.cancelled || msg.cancelled {
			// Do not start the next copy or remove. Work already finished
			// stays as it was; nothing further is rolled back or replaced.
			m.cancelled = true
			m.stop()
			return m, tea.Quit
		}

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
				m.stop()
				return m, nil
			}
		}

		// Move to next task
		m.currentTaskIndex++
		if m.currentTaskIndex >= len(m.tasks) {
			m.step = stepComplete
			m.stop()
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
		if m.cancelled {
			return "Cancelling..."
		}
		return "Q/Ctrl+C: Cancel"
	}
}

func executeTask(index int, m *model) tea.Cmd {
	return func() tea.Msg {
		ctx := ctxOf(m)

		// Keep the step visible briefly, but don't ignore cancel during the pause.
		timer := time.NewTimer(200 * time.Millisecond)
		defer timer.Stop()
		select {
		case <-ctx.Done():
			return taskCompleteMsg{index: index, success: false, error: ctx.Err().Error(), cancelled: true}
		case <-timer.C:
		}

		err := m.tasks[index].execute(m)
		if err != nil {
			cancelled := errors.Is(err, context.Canceled) || ctx.Err() != nil
			return taskCompleteMsg{
				index:     index,
				success:   false,
				error:     err.Error(),
				cancelled: cancelled,
			}
		}
		if ctx.Err() != nil {
			return taskCompleteMsg{index: index, success: false, error: ctx.Err().Error(), cancelled: true}
		}

		return taskCompleteMsg{
			index:   index,
			success: true,
		}
	}
}

// ctxOf returns the install context, or Background when there is no model.
func ctxOf(m *model) context.Context {
	if m == nil || m.ctx == nil {
		return context.Background()
	}
	return m.ctx
}

// errIfCancelled reports why the install was cancelled, or nil if it still runs.
func errIfCancelled(m *model) error {
	return ctxOf(m).Err()
}

// stop cancels the install context. It is safe to call more than once.
func (m *model) stop() {
	if m.cancel != nil {
		m.cancel()
	}
}

// Task functions

func checkPrivileges(m *model) error {
	if err := errIfCancelled(m); err != nil {
		return err
	}
	if os.Geteuid() != 0 {
		return fmt.Errorf("installer must be run with sudo or as root")
	}
	return nil
}

func buildBinary(m *model) error {
	return buildStaged(m, "syscgo", "./cmd/syscgo")
}

func buildTuiBinary(m *model) error {
	return buildStaged(m, "syscgo-tui", "./cmd/syscgo-tui")
}

func buildStaged(m *model, name, pkg string) error {
	if err := errIfCancelled(m); err != nil {
		return err
	}
	out, err := stagedBinaryPath(name)
	if err != nil {
		return err
	}
	goBin, err := goToolchain()
	if err != nil {
		return err
	}
	// -buildvcs=false: the build runs as root against a user-owned checkout, so
	// git cannot read the repo state and "go build" fails with exit status 128.
	cmd := commandContext(m, goBin, "build", "-buildvcs=false", "-o", out, pkg)
	cmd.Dir = getProjectRoot()
	output, err := cmd.CombinedOutput()
	if err != nil {
		if errIfCancelled(m) != nil {
			return context.Canceled
		}
		return fmt.Errorf("build failed: %s", string(output))
	}
	return nil
}

// Install destinations. Package variables so tests can point them at a temp
// directory instead of the real filesystem.
var (
	binDir   = "/usr/local/bin"
	shareDir = "/usr/local/share/syscgo"
)

// Modes are applied explicitly rather than relying on the process umask: sudo
// keeps the invoking user's stricter umask, so a 0755 request can otherwise
// land as 0700 and leave the installed binary or asset unreadable.
const (
	binMode   = 0o755
	dirMode   = 0o755
	assetMode = 0o644
)

// installFileAtomic writes data to dst by way of a temporary file in dst's own
// directory, so the final rename is atomic and never truncates the destination.
// That keeps a running binary executable (no ETXTBSY) and stops a failed write
// from leaving a truncated, unbootable file behind.
func installFileAtomic(dst string, data []byte, mode os.FileMode) error {
	dir := filepath.Dir(dst)
	tmp, err := os.CreateTemp(dir, "."+filepath.Base(dst)+".tmp-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	committed := false
	defer func() {
		if !committed {
			tmp.Close()
			os.Remove(tmpName)
		}
	}()
	if _, err := tmp.Write(data); err != nil {
		return err
	}
	// Chmod before the rename so the file is never visible at dst with the
	// 0600 mode CreateTemp gives it, and never depends on the umask.
	if err := tmp.Chmod(mode); err != nil {
		return err
	}
	if err := tmp.Sync(); err != nil {
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Rename(tmpName, dst); err != nil {
		return err
	}
	committed = true
	return nil
}

// ensureDir creates dir and sets its mode explicitly, which both defeats a
// restrictive umask and repairs the permissions of a directory a previous run
// already created too tight.
func ensureDir(dir string) error {
	if err := os.MkdirAll(dir, dirMode); err != nil {
		return err
	}
	return os.Chmod(dir, dirMode)
}

func installAssets(m *model) error {
	if err := errIfCancelled(m); err != nil {
		return err
	}
	projectRoot := getProjectRoot()
	srcPath := filepath.Join(projectRoot, "assets")
	if err := installAssetFiles(ctxOf(m), srcPath, shareDir); err != nil {
		if errors.Is(err, context.Canceled) {
			return err
		}
		return fmt.Errorf("failed to copy assets: %v", err)
	}
	return nil
}

// installAssetFiles copies repo assets into the system share layout used by the
// CLI and TUI: text files under <share>/assets and BIT fonts under <share>/fonts.
// ctx must be non-nil; it is checked before every file so a cancel stops the copy.
func installAssetFiles(ctx context.Context, srcAssets, shareRoot string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	entries, err := os.ReadDir(srcAssets)
	if err != nil {
		return err
	}

	assetsDst := filepath.Join(shareRoot, "assets")
	if err := ensureDir(assetsDst); err != nil {
		return fmt.Errorf("failed to create directory: %v", err)
	}

	for _, entry := range entries {
		if err := ctx.Err(); err != nil {
			return err
		}
		src := filepath.Join(srcAssets, entry.Name())
		if entry.IsDir() {
			if entry.Name() != "fonts" {
				continue
			}
			fonts, err := os.ReadDir(src)
			if err != nil {
				return err
			}
			for _, font := range fonts {
				if err := ctx.Err(); err != nil {
					return err
				}
				if font.IsDir() || !strings.EqualFold(filepath.Ext(font.Name()), ".bit") {
					continue
				}
				fontsDst := filepath.Join(shareRoot, "fonts")
				if err := ensureDir(fontsDst); err != nil {
					return fmt.Errorf("failed to create directory: %v", err)
				}
				if err := copyFile(filepath.Join(src, font.Name()), filepath.Join(fontsDst, font.Name())); err != nil {
					return err
				}
			}
			continue
		}
		if !strings.EqualFold(filepath.Ext(entry.Name()), ".txt") {
			continue
		}
		if err := copyFile(src, filepath.Join(assetsDst, entry.Name())); err != nil {
			return err
		}
	}

	return nil
}

func installBinary(m *model) error {
	if err := errIfCancelled(m); err != nil {
		return err
	}
	srcPath, err := stagedBinaryPath("syscgo")
	if err != nil {
		return err
	}
	dstPath := filepath.Join(binDir, "syscgo")

	// Read the source file
	data, err := os.ReadFile(srcPath)
	if err != nil {
		return fmt.Errorf("failed to read binary: %v", err)
	}

	// A cancel that lands before the write must not truncate an existing binary.
	if err := errIfCancelled(m); err != nil {
		return err
	}
	err = installFileAtomic(dstPath, data, binMode)
	if err != nil {
		return fmt.Errorf("failed to install binary: %v", err)
	}

	return nil
}

func installTuiBinary(m *model) error {
	if err := errIfCancelled(m); err != nil {
		return err
	}
	srcPath, err := stagedBinaryPath("syscgo-tui")
	if err != nil {
		return err
	}
	dstPath := filepath.Join(binDir, "syscgo-tui")

	// Read the source file
	data, err := os.ReadFile(srcPath)
	if err != nil {
		return fmt.Errorf("failed to read binary: %v", err)
	}

	if err := errIfCancelled(m); err != nil {
		return err
	}
	err = installFileAtomic(dstPath, data, binMode)
	if err != nil {
		return fmt.Errorf("failed to install binary: %v", err)
	}

	return nil
}

func removeSyscgoBinary(m *model) error {
	if err := errIfCancelled(m); err != nil {
		return err
	}
	err := os.Remove(filepath.Join(binDir, "syscgo"))
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove binary: %v", err)
	}
	return nil
}

func removeTuiBinary(m *model) error {
	if err := errIfCancelled(m); err != nil {
		return err
	}
	err := os.Remove(filepath.Join(binDir, "syscgo-tui"))
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove binary: %v", err)
	}
	return nil
}

func removeAssets(m *model) error {
	if err := errIfCancelled(m); err != nil {
		return err
	}
	err := os.RemoveAll(shareDir)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove assets: %v", err)
	}
	return nil
}

// copyFile copies a single file. The destination mode is assetMode rather than
// the source's: a checkout made under a restrictive umask carries 0600 files,
// and copying those modes verbatim installs assets nobody else can read.
func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return installFileAtomic(dst, data, assetMode)
}

// buildOutputDir is the temp directory that receives syscgo and syscgo-tui
// before they are copied to /usr/local/bin. Building here avoids leaving
// root-owned binaries in the source tree when the installer runs as root.
var buildOutputDir string

func stagedBinaryPath(name string) (string, error) {
	if buildOutputDir == "" {
		dir, err := os.MkdirTemp("", "syscgo-install-")
		if err != nil {
			return "", fmt.Errorf("failed to create build directory: %v", err)
		}
		buildOutputDir = dir
	}
	return filepath.Join(buildOutputDir, name), nil
}

// commandContext runs name so a cancelled install can stop it and the
// children it spawned (go build launches compile and link).
func commandContext(m *model, name string, args ...string) *exec.Cmd {
	cmd := exec.CommandContext(ctxOf(m), name, args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.WaitDelay = time.Second
	cmd.Cancel = func() error {
		if cmd.Process == nil {
			return nil
		}
		// ponytail: the group id is the child's pid, so this is only safe while
		// that pid is unreaped. os/exec can call Cancel in the same tick the
		// child exits and its pid is recycled; a raw kill cannot detect that the
		// way Process.Kill reports os.ErrProcessDone. Window is a few ns and the
		// child is normally still running, which is why this stays.
		err := syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
		if err != nil && !errors.Is(err, syscall.ESRCH) {
			return err
		}
		return nil
	}
	return cmd
}

func cleanupBuildOutput() {
	if buildOutputDir == "" {
		return
	}
	_ = os.RemoveAll(buildOutputDir)
	buildOutputDir = ""
}

// findModuleRoot walks upward from start until it finds a directory containing go.mod.
func findModuleRoot(start string) (string, bool) {
	if start == "" {
		return "", false
	}
	dir := start
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, true
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", false
		}
		dir = parent
	}
}

// getProjectRoot resolves the module the installer belongs to. install.sh places
// the installer binary in the repo root. `go run` and `go build -o cmd/installer/...`
// use other depths, so the lookup walks until go.mod instead of assuming one.
// The executable is consulted before the cwd; see projectRootFrom.
func getProjectRoot() string {
	execPath, err := os.Executable()
	if err != nil {
		return projectRootFrom("")
	}
	return projectRootFrom(execPath)
}

// projectRootFrom is getProjectRoot with the executable path supplied by the
// caller, so the lookup order is exercisable without relocating the running
// test binary. An empty execPath skips the executable and forces the cwd.
func projectRootFrom(execPath string) string {
	if execPath != "" {
		if root, ok := findModuleRoot(filepath.Dir(execPath)); ok {
			return root
		}
	}

	if dir, err := os.Getwd(); err == nil {
		if root, ok := findModuleRoot(dir); ok {
			return root
		}
	}

	return "."
}

// nonInteractiveRequested reports whether install tasks should run without the TUI.
// args are the program arguments after the executable name. env is the value of
// SYSCGO_INSTALL_NONINTERACTIVE.
func nonInteractiveRequested(args []string, env string) bool {
	if env == "1" {
		return true
	}
	for _, arg := range args {
		if arg == "--yes" || arg == "-y" {
			return true
		}
	}
	return false
}

// runConfiguredTasks runs install or uninstall tasks in order.
// When m.tasks is empty, the standard install task list is used.
func runConfiguredTasks(m *model) error {
	if len(m.tasks) == 0 {
		m.initTasks()
	}
	for i := range m.tasks {
		task := &m.tasks[i]
		fmt.Printf("%s...\n", task.description)
		if err := task.execute(m); err != nil {
			if task.optional {
				fmt.Printf("skip: %s: %v\n", task.name, err)
				continue
			}
			return fmt.Errorf("%s: %w", task.name, err)
		}
	}
	return nil
}

func main() {
	os.Exit(runInstaller())
}

func runInstaller() int {
	// os.Exit skips defers in main, so cleanup lives in this function.
	defer cleanupBuildOutput()

	// Builds run with Setpgid (commandContext), so the compiler group never sees
	// the terminal's SIGINT. Without this the --yes path would die by default
	// SIGINT action, skip the deferred cleanup and orphan `go build` while
	// install.sh's EXIT trap deleted the tree it was compiling from.
	ctx, stopSignals := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stopSignals()

	// Check if go is installed (PATH, else the GOROOT this ran from under sudo)
	if _, err := goToolchain(); err != nil {
		fmt.Println("Error: Go is not installed or not in PATH")
		fmt.Println("Please install Go from https://golang.org/dl/")
		return 1
	}

	if nonInteractiveRequested(os.Args[1:], os.Getenv("SYSCGO_INSTALL_NONINTERACTIVE")) {
		m := newModel(ctx)
		err := runConfiguredTasks(&m)
		switch {
		case ctx.Err() != nil:
			fmt.Fprintln(os.Stderr, "Error: install cancelled")
			return 130
		case err != nil:
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			return 1
		}
		fmt.Println("Installation complete.")
		return 0
	}

	p := tea.NewProgram(newModel(ctx), tea.WithAltScreen())

	final, err := p.Run()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return 1
	}
	// The alt screen is gone, so whatever it showed is gone with it. Print the
	// outcome to the normal terminal and carry it in the exit status.
	if fm, ok := final.(model); ok {
		return fm.report()
	}
	return 0
}

// report prints the install outcome after the alt screen closes and returns the
// exit code. The step alone tells the outcome: stepInstalling means the run
// ended before finishing (Q, Ctrl+C, or a signal), stepComplete means it ran to
// the end, and only then can a task have failed.
func (m model) report() int {
	switch {
	case m.step == stepInstalling:
		fmt.Fprintln(os.Stderr, "Installation cancelled. Nothing further was installed.")
		return 130
	case m.failed():
		fmt.Fprintf(os.Stderr, "Installation failed: %s\n", strings.Join(m.errors, "; "))
		return 1
	default:
		fmt.Println("Installation complete.")
		return 0
	}
}

// failed reports whether a required task failed. Optional skips are recorded in
// errors too, so the task status is what separates them.
func (m model) failed() bool {
	for _, t := range m.tasks {
		if t.status == statusFailed {
			return true
		}
	}
	return false
}
