package main

import (
	"bytes"
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
	"unicode"
	"unicode/utf8"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
)

const (
	cyan   = "\x1b[36m"
	green  = "\x1b[32m"
	red    = "\x1b[31m"
	gray   = "\x1b[38;5;247m"
	bold   = "\x1b[1m"
	reset  = "\x1b[0m"
	cursor = "▌"
)

type screen uint8

const (
	stackScreen screen = iota
	packageManagerScreen
	pythonVenvScreen
	folderScreen
	workingScreen
	doneScreen
	errorScreen
)

type stackChoice struct {
	name        string
	value       string
	template    string
	node        bool
	python      bool
	startScript string
}

var stackChoices = []stackChoice{
	{name: "React + JS", value: "Js_react", template: "react-js", node: true, startScript: "dev"},
	{name: "React + TS", value: "Ts_react", template: "react-ts", node: true, startScript: "dev"},
	{name: "Express + Prisma", value: "Express_Prisma", template: "express-prisma", node: true, startScript: "dev"},
	{name: "Express + Mongoose", value: "Express_Mongoose", template: "express-mongoose", node: true, startScript: "dev"},
	{name: "Fiber + SQLC", value: "Fiber_Sqlc", template: "fiber-sqlc"},
	{name: "Go+SQLC", value: "Go_http", template: "go-http"},
	{name: "Django", value: "Django", template: "django", python: true},
	{name: "FastAPI", value: "FastAPI", template: "fastapi", python: true},
	{name: "Expo", value: "Expo", template: "expo", node: true, startScript: "start"},
}

var spinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

var version = "dev"

var errProjectFolderExists = errors.New("project folder already exists")

type operation struct {
	label      string
	liveOutput *synchronizedBuffer
	run        func(context.Context) (string, error)
}

type operationResult struct {
	output string
	err    error
}

type spinnerTick struct{}

type synchronizedBuffer struct {
	mutex  sync.RWMutex
	buffer bytes.Buffer
}

func (b *synchronizedBuffer) Write(value []byte) (int, error) {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	return b.buffer.Write(value)
}

func (b *synchronizedBuffer) String() string {
	b.mutex.RLock()
	defer b.mutex.RUnlock()
	return b.buffer.String()
}

type model struct {
	screen           screen
	cursor           int
	selectedStack    stackChoice
	packageManager   string
	createPythonVenv bool
	folderName       string
	folderError      string
	targetDir        string
	operations       []operation
	operationIndex   int
	operationTimes   []time.Duration
	startedAt        time.Time
	operationStarted time.Time
	spinnerFrame     int
	animationTick    int
	width            int
	pointerActive    bool
	verbose          bool
	commandOutput    string
	setupErr         error
	completed        bool
	cancelled        bool
	ctx              context.Context
	cancel           context.CancelFunc
}

func newModel(verbose bool) model {
	ctx, cancel := context.WithCancel(context.Background())
	return model{
		screen:     stackScreen,
		folderName: "letsinit-project",
		verbose:    verbose,
		ctx:        ctx,
		cancel:     cancel,
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(message tea.Msg) (tea.Model, tea.Cmd) {
	switch message := message.(type) {
	case tea.KeyPressMsg:
		key := message.String()
		if key == "ctrl+c" || key == "esc" || (key == "q" && m.screen != folderScreen) {
			m.cancel()
			m.cancelled = true
			m.pointerActive = false
			return m, tea.Sequence(tea.Raw(ansi.SetPointerShape("default")), tea.Quit)
		}

		switch m.screen {
		case stackScreen:
			return m.updateStackSelection(key)
		case packageManagerScreen:
			return m.updatePackageManagerSelection(key)
		case pythonVenvScreen:
			return m.updatePythonVenvSelection(key)
		case folderScreen:
			return m.updateFolderInput(message)
		}

	case tea.PasteMsg:
		if m.screen == folderScreen {
			m.folderName += printableText(message.Content)
			m.folderError = ""
		}

	case tea.MouseMotionMsg:
		return m.updateMouseHover(message.Mouse())

	case tea.MouseClickMsg:
		if message.Mouse().Button == tea.MouseLeft {
			return m.updateMouseClick(message.Mouse())
		}

	case tea.MouseWheelMsg:
		return m.updateMouseWheel(message.Mouse())

	case tea.WindowSizeMsg:
		m.width = message.Width

	case spinnerTick:
		if m.screen == workingScreen {
			m.spinnerFrame = (m.spinnerFrame + 1) % len(spinnerFrames)
			m.animationTick++
			if output := m.currentLiveOutput(); output != "" && m.verbose {
				m.commandOutput = output
			}
			return m, tickSpinner()
		}

	case operationResult:
		if m.screen != workingScreen {
			return m, nil
		}

		m.operationTimes = append(m.operationTimes, time.Since(m.operationStarted))
		if message.output != "" && (m.verbose || message.err != nil) {
			m.commandOutput = message.output
		}
		if message.err != nil {
			m.setupErr = message.err
			m.screen = errorScreen
			m.cancel()
			return m, tea.Quit
		}

		m.operationIndex++
		if m.operationIndex == len(m.operations) {
			m.completed = true
			m.screen = doneScreen
			m.cancel()
			return m, tea.Quit
		}
		m.commandOutput = ""
		m.operationStarted = time.Now()
		return m, runOperation(m.ctx, m.operations[m.operationIndex])
	}

	return m, nil
}

func (m model) updateStackSelection(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "up", "k":
		m.cursor = wrapCursor(m.cursor-1, len(stackChoices))
	case "down", "j":
		m.cursor = wrapCursor(m.cursor+1, len(stackChoices))
	case "enter":
		m.selectedStack = stackChoices[m.cursor]
		m.cursor = 0
		if m.selectedStack.node {
			m.screen = packageManagerScreen
		} else if m.selectedStack.python {
			m.screen = pythonVenvScreen
		} else {
			m.screen = folderScreen
		}
		if m.screen == packageManagerScreen || m.screen == pythonVenvScreen {
			return m, nil
		}
		m.pointerActive = false
		return m, tea.Raw(ansi.SetPointerShape("default"))
	}
	return m, nil
}

func (m model) updatePackageManagerSelection(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "up", "k", "left", "h", "down", "j", "right", "l":
		m.cursor = wrapCursor(m.cursor+1, 2)
	case "enter":
		if m.cursor == 0 {
			m.packageManager = "npm"
		} else {
			m.packageManager = "pnpm"
		}
		m.cursor = 0
		m.screen = folderScreen
		m.pointerActive = false
		return m, tea.Raw(ansi.SetPointerShape("default"))
	}
	return m, nil
}

func (m model) updatePythonVenvSelection(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "up", "k", "left", "h", "down", "j", "right", "l":
		m.cursor = wrapCursor(m.cursor+1, 2)
	case "enter":
		m.createPythonVenv = m.cursor == 0
		m.cursor = 0
		m.screen = folderScreen
		m.pointerActive = false
		return m, tea.Raw(ansi.SetPointerShape("default"))
	}
	return m, nil
}

func (m model) updateMouseHover(mouse tea.Mouse) (tea.Model, tea.Cmd) {
	choice, selectable := m.choiceAt(mouse.X, mouse.Y)
	if selectable {
		m.cursor = choice
	}
	if selectable == m.pointerActive {
		return m, nil
	}

	m.pointerActive = selectable
	shape := "default"
	if selectable {
		shape = "pointer"
	}
	return m, tea.Raw(ansi.SetPointerShape(shape))
}

func (m model) updateMouseClick(mouse tea.Mouse) (tea.Model, tea.Cmd) {
	choice, selectable := m.choiceAt(mouse.X, mouse.Y)
	if !selectable {
		return m.updateMouseHover(mouse)
	}
	m.cursor = choice

	switch m.screen {
	case stackScreen:
		return m.updateStackSelection("enter")
	case packageManagerScreen:
		return m.updatePackageManagerSelection("enter")
	case pythonVenvScreen:
		return m.updatePythonVenvSelection("enter")
	default:
		return m, nil
	}
}

func (m model) updateMouseWheel(mouse tea.Mouse) (tea.Model, tea.Cmd) {
	choices, _ := m.selectableChoices()
	if len(choices) == 0 {
		return m, nil
	}

	switch mouse.Button {
	case tea.MouseWheelUp:
		m.cursor = wrapCursor(m.cursor-1, len(choices))
	case tea.MouseWheelDown:
		m.cursor = wrapCursor(m.cursor+1, len(choices))
	}
	return m, nil
}

func (m model) choiceAt(x, y int) (int, bool) {
	choices, firstRow := m.selectableChoices()
	choice := y - firstRow
	if choice < 0 || choice >= len(choices) {
		return 0, false
	}

	choiceEnd := 4 + ansi.StringWidth(choices[choice])
	if x < 2 || x > choiceEnd {
		return 0, false
	}
	return choice, true
}

func (m model) selectableChoices() ([]string, int) {
	// renderLayout adds one empty line above the content.
	switch m.screen {
	case stackScreen:
		return stackNames(), 6
	case packageManagerScreen:
		return []string{"npm", "pnpm"}, 8
	case pythonVenvScreen:
		return []string{"Yes", "No"}, 8
	default:
		return nil, 0
	}
}

func (m model) updateFolderInput(key tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch key.String() {
	case "enter":
		if strings.TrimSpace(m.folderName) == "" {
			m.folderError = "Enter a project folder name."
			return m, nil
		}
		operations, targetDir, err := prepareOperations(m)
		if err != nil {
			if errors.Is(err, errProjectFolderExists) {
				m.folderError = fmt.Sprintf("%q already exists. Enter a different folder name.", m.folderName)
				m.folderName = ""
				return m, nil
			}
			m.setupErr = err
			m.screen = errorScreen
			m.cancel()
			return m, tea.Quit
		}
		m.operations = operations
		m.targetDir = targetDir
		m.screen = workingScreen
		m.startedAt = time.Now()
		m.operationStarted = m.startedAt
		m.operationTimes = make([]time.Duration, 0, len(operations))
		return m, tea.Batch(runOperation(m.ctx, m.operations[0]), tickSpinner())

	case "backspace":
		if len(m.folderName) > 0 {
			_, size := utf8.DecodeLastRuneInString(m.folderName)
			m.folderName = m.folderName[:len(m.folderName)-size]
		}
		m.folderError = ""

	default:
		m.folderName += printableText(key.Text)
		m.folderError = ""
	}
	return m, nil
}

func (m model) View() tea.View {
	var view strings.Builder
	view.WriteString(bold + cyan + "  LetsInit" + reset + "\n")
	view.WriteString(gray + "  Scaffold a ready-to-use project" + reset + "\n\n")

	switch m.screen {
	case stackScreen:
		view.WriteString(bold + "  Choose your stack:" + reset + "\n\n")
		writeChoices(&view, stackNames(), m.cursor)

	case packageManagerScreen:
		view.WriteString(selectionSummary(m))
		view.WriteString(bold + "  Choose your package manager:" + reset + "\n\n")
		writeChoices(&view, []string{"npm", "pnpm"}, m.cursor)

	case pythonVenvScreen:
		view.WriteString(selectionSummary(m))
		view.WriteString(bold + "  Create a Python virtual environment?" + reset + "\n\n")
		writeChoices(&view, []string{"Yes", "No"}, m.cursor)

	case folderScreen:
		view.WriteString(selectionSummary(m))
		view.WriteString(bold + "  Enter your project folder name:" + reset + "\n\n")
		view.WriteString("  " + cyan + "> " + reset + m.folderName + cyan + cursor + reset + "\n")
		if m.folderError != "" {
			view.WriteString("\n  " + red + "! " + m.folderError + reset + "\n")
		}

	case workingScreen:
		view.WriteString(selectionSummary(m))
		currentOperation := m.operations[m.operationIndex]
		fmt.Fprintf(
			&view,
			"  %s%s%s %s%s%s\n",
			cyan,
			spinnerFrames[m.spinnerFrame],
			reset,
			bold,
			currentOperation.label,
			reset,
		)
		fmt.Fprintf(
			&view,
			"  %s%s  •  elapsed %s%s\n\n",
			gray,
			activityDetail(currentOperation.label, m.animationTick),
			formatDuration(time.Since(m.startedAt)),
			reset,
		)
		fmt.Fprintf(&view, "  %sStep %d of %d%s\n\n", gray, m.operationIndex+1, len(m.operations), reset)

		for index, operation := range m.operations {
			switch {
			case index < m.operationIndex:
				fmt.Fprintf(
					&view,
					"  %s✓%s %s %s(%s)%s\n",
					green,
					reset,
					operation.label,
					gray,
					formatDuration(m.operationTimes[index]),
					reset,
				)
			case index == m.operationIndex:
				fmt.Fprintf(&view, "  %s◆%s %s\n", cyan, reset, operation.label)
			default:
				fmt.Fprintf(&view, "  %s○%s %s\n", gray, reset, operation.label)
			}
		}
		if m.verbose && m.commandOutput != "" {
			view.WriteString("\n" + gray + indent(tailLines(m.commandOutput, 10), "  ") + reset + "\n")
		}

	case doneScreen:
		view.WriteString(green + bold + "  Project setup completed!" + reset + "\n")

	case errorScreen:
		view.WriteString(red + bold + "  Project setup failed." + reset + "\n")
	}

	view.WriteString("\n")
	if m.screen == folderScreen {
		view.WriteString(gray + "  enter continue • esc cancel" + reset)
	} else if m.screen == stackScreen || m.screen == packageManagerScreen || m.screen == pythonVenvScreen {
		view.WriteString(gray + "  mouse click • ↑/↓ move • enter select • q quit" + reset)
	} else if m.screen == workingScreen {
		view.WriteString(gray + "  esc cancel" + reset)
	}
	view.WriteString("\n")

	rendered := tea.NewView(renderLayout(view.String(), m.width))
	rendered.AltScreen = true
	rendered.WindowTitle = m.windowTitle()
	if m.screen == stackScreen || m.screen == packageManagerScreen || m.screen == pythonVenvScreen {
		rendered.MouseMode = tea.MouseModeAllMotion
	}
	return rendered
}

func (m model) windowTitle() string {
	switch m.screen {
	case stackScreen:
		return "LetsInit - Choose a stack"
	case packageManagerScreen:
		return "LetsInit - " + m.selectedStack.name + " - Package manager"
	case pythonVenvScreen:
		return "LetsInit - " + m.selectedStack.name + " - Python environment"
	case folderScreen:
		return "LetsInit - New " + m.selectedStack.name + " project"
	case workingScreen:
		if m.operationIndex < len(m.operations) {
			return "LetsInit - " + m.operations[m.operationIndex].label
		}
		return "LetsInit - Setting up project"
	case doneScreen:
		return "LetsInit - Project ready"
	case errorScreen:
		return "LetsInit - Setup failed"
	default:
		return "LetsInit"
	}
}

func prepareOperations(m model) ([]operation, string, error) {
	templateRoot, err := findTemplateRoot()
	if err != nil {
		return nil, "", err
	}

	templateDir := filepath.Join(templateRoot, m.selectedStack.template)
	if info, statErr := os.Stat(templateDir); statErr != nil || !info.IsDir() {
		if statErr == nil {
			statErr = errors.New("not a directory")
		}
		return nil, "", fmt.Errorf("template not found for %s: %w", m.selectedStack.value, statErr)
	}

	targetDir, err := filepath.Abs(m.folderName)
	if err != nil {
		return nil, "", fmt.Errorf("resolve project folder: %w", err)
	}
	if err := validateTargetDirectory(targetDir); err != nil {
		return nil, "", err
	}

	operations := []operation{
		{
			label: "Creating project from template",
			run: func(ctx context.Context) (string, error) {
				if err := copyTemplate(ctx, templateDir, targetDir, m.packageManager); err != nil {
					return "", fmt.Errorf("create project: %w", err)
				}
				return "", nil
			},
		},
	}

	if m.createPythonVenv {
		python := systemPythonCommand()
		operations = append(operations,
			commandOperation("Creating Python virtual environment", targetDir, python[0], append(python[1:], "-m", "venv", ".venv")...),
			commandOperation("Installing Python dependencies", targetDir, virtualEnvironmentPython(), "-m", "pip", "install", "-r", "requirements.txt"),
		)
	} else if m.selectedStack.node {
		operations = append(operations,
			commandOperation("Installing dependencies with "+m.packageManager, targetDir, m.packageManager, "install"),
		)
	} else if m.selectedStack.value == "Fiber_Sqlc" {
		operations = append(operations,
			commandOperation("Installing Go dependencies", targetDir, "go", "mod", "download"),
		)
	}

	operations = append(operations, operation{
		label: "Preparing .gitignore",
		run: func(context.Context) (string, error) {
			return "", ensureEnvIgnored(targetDir)
		},
	})

	if m.selectedStack.value == "Express_Prisma" {
		if m.packageManager == "pnpm" {
			operations = append(operations,
				commandOperation("Generating Prisma Client", targetDir, "pnpm", "exec", "prisma", "generate"),
			)
		} else {
			operations = append(operations,
				commandOperation("Generating Prisma Client", targetDir, "npm", "exec", "--", "prisma", "generate"),
			)
		}
	}

	return operations, targetDir, nil
}

func commandOperation(label, directory, name string, arguments ...string) operation {
	liveOutput := &synchronizedBuffer{}
	return operation{
		label:      label,
		liveOutput: liveOutput,
		run: func(ctx context.Context) (string, error) {
			command := exec.CommandContext(ctx, name, arguments...)
			command.Dir = directory
			command.Stdout = liveOutput
			command.Stderr = liveOutput
			err := command.Run()
			cleanOutput := strings.TrimSpace(liveOutput.String())
			if err != nil {
				return cleanOutput, fmt.Errorf("run %s: %w", strings.Join(append([]string{name}, arguments...), " "), err)
			}
			return cleanOutput, nil
		},
	}
}

func runOperation(ctx context.Context, operation operation) tea.Cmd {
	return func() tea.Msg {
		output, err := operation.run(ctx)
		return operationResult{output: output, err: err}
	}
}

func tickSpinner() tea.Cmd {
	return tea.Tick(90*time.Millisecond, func(time.Time) tea.Msg {
		return spinnerTick{}
	})
}

func (m model) currentLiveOutput() string {
	if m.operationIndex >= len(m.operations) {
		return ""
	}
	output := m.operations[m.operationIndex].liveOutput
	if output == nil {
		return ""
	}
	return strings.TrimSpace(output.String())
}

func findTemplateRoot() (string, error) {
	var candidates []string
	if configured := os.Getenv("LETSINIT_TEMPLATES_DIR"); configured != "" {
		candidates = append(candidates, configured)
	}

	if workingDirectory, err := os.Getwd(); err == nil {
		candidates = append(candidates,
			filepath.Join(workingDirectory, "cli", "stackforge-cli", "templates"),
			filepath.Join(workingDirectory, "..", "cli", "stackforge-cli", "templates"),
			filepath.Join(workingDirectory, "templates"),
		)
	}

	if _, sourceFile, _, ok := runtime.Caller(0); ok {
		candidates = append(candidates,
			filepath.Join(filepath.Dir(sourceFile), "..", "templates"),
			filepath.Join(filepath.Dir(sourceFile), "..", "cli", "stackforge-cli", "templates"),
		)
	}

	if executable, err := os.Executable(); err == nil {
		executableDirectory := filepath.Dir(executable)
		candidates = append(candidates,
			filepath.Join(executableDirectory, "cli", "stackforge-cli", "templates"),
			filepath.Join(executableDirectory, "..", "cli", "stackforge-cli", "templates"),
			filepath.Join(executableDirectory, "templates"),
		)
		if resolvedExecutable, resolveErr := filepath.EvalSymlinks(executable); resolveErr == nil {
			candidates = append(candidates, filepath.Join(filepath.Dir(resolvedExecutable), "templates"))
		}
	}

	seen := make(map[string]struct{}, len(candidates))
	for _, candidate := range candidates {
		absoluteCandidate, err := filepath.Abs(candidate)
		if err != nil {
			continue
		}
		absoluteCandidate = filepath.Clean(absoluteCandidate)
		if _, exists := seen[absoluteCandidate]; exists {
			continue
		}
		seen[absoluteCandidate] = struct{}{}

		if info, err := os.Stat(absoluteCandidate); err == nil && info.IsDir() {
			return absoluteCandidate, nil
		}
	}

	return "", errors.New("templates directory not found; set LETSINIT_TEMPLATES_DIR to cli/stackforge-cli/templates")
}

func validateTargetDirectory(targetDir string) error {
	_, err := os.Stat(targetDir)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("inspect project folder: %w", err)
	}
	return fmt.Errorf("%w: %q", errProjectFolderExists, targetDir)
}

func copyTemplate(ctx context.Context, sourceDir, targetDir, packageManager string) error {
	if err := validateTargetDirectory(targetDir); err != nil {
		return err
	}

	sourceInfo, err := os.Stat(sourceDir)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(targetDir, sourceInfo.Mode().Perm()); err != nil {
		return err
	}

	err = filepath.WalkDir(sourceDir, func(sourcePath string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if err := ctx.Err(); err != nil {
			return err
		}

		relativePath, err := filepath.Rel(sourceDir, sourcePath)
		if err != nil {
			return err
		}
		if relativePath == "." {
			return nil
		}
		targetPath := filepath.Join(targetDir, relativePath)

		info, err := entry.Info()
		if err != nil {
			return err
		}
		switch {
		case entry.Type()&os.ModeSymlink != 0:
			linkTarget, err := os.Readlink(sourcePath)
			if err != nil {
				return err
			}
			return os.Symlink(linkTarget, targetPath)

		case entry.IsDir():
			return os.MkdirAll(targetPath, info.Mode().Perm())

		case info.Mode().IsRegular():
			return copyFile(sourcePath, targetPath, info.Mode())

		default:
			return fmt.Errorf("unsupported template file %q", sourcePath)
		}
	})
	if err != nil {
		return err
	}

	for _, names := range [][2]string{{"_gitignore", ".gitignore"}, {".env.example", ".env"}} {
		sourcePath := filepath.Join(targetDir, names[0])
		targetPath := filepath.Join(targetDir, names[1])
		if err := os.Rename(sourcePath, targetPath); err != nil && !errors.Is(err, os.ErrNotExist) {
			return err
		}
	}

	if packageManager == "pnpm" {
		if err := removeIfExists(filepath.Join(targetDir, "package-lock.json")); err != nil {
			return err
		}
	} else if packageManager == "npm" {
		if err := removeIfExists(filepath.Join(targetDir, "pnpm-lock.yaml")); err != nil {
			return err
		}
	}

	return nil
}

func copyFile(sourcePath, targetPath string, mode os.FileMode) error {
	source, err := os.Open(sourcePath)
	if err != nil {
		return err
	}
	defer source.Close()

	target, err := os.OpenFile(targetPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, mode.Perm())
	if err != nil {
		return err
	}

	_, copyErr := io.Copy(target, source)
	closeErr := target.Close()
	if copyErr != nil {
		return copyErr
	}
	return closeErr
}

func removeIfExists(path string) error {
	err := os.Remove(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	return err
}

func ensureEnvIgnored(targetDir string) error {
	ignorePath := filepath.Join(targetDir, ".gitignore")
	content, err := os.ReadFile(ignorePath)
	if errors.Is(err, os.ErrNotExist) {
		return os.WriteFile(ignorePath, []byte(".env\n"), 0o644)
	}
	if err != nil {
		return err
	}
	if strings.Contains(string(content), ".env") {
		return nil
	}

	file, err := os.OpenFile(ignorePath, os.O_APPEND|os.O_WRONLY, 0)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = file.WriteString("\n.env\n")
	return err
}

func systemPythonCommand() []string {
	if runtime.GOOS == "windows" {
		return []string{"py", "-3"}
	}
	return []string{"python3"}
}

func virtualEnvironmentPython() string {
	if runtime.GOOS == "windows" {
		return filepath.Join(".venv", "Scripts", "python.exe")
	}
	return filepath.Join(".venv", "bin", "python")
}

func printResult(m model) int {
	if m.cancelled && !m.completed && m.setupErr == nil {
		fmt.Println(gray + "Setup cancelled." + reset)
		return 0
	}
	if m.setupErr != nil {
		fmt.Fprintln(os.Stderr, red+"Error during setup:"+reset, m.setupErr)
		if m.commandOutput != "" {
			fmt.Fprintln(os.Stderr, m.commandOutput)
		}
		return 1
	}
	if !m.completed {
		return 0
	}

	fmt.Printf("\n%s%s✓ Project initialized%s\n\n", green, bold, reset)
	fmt.Println(bold + "Next steps" + reset)
	for _, command := range nextStepCommands(m) {
		fmt.Printf("  %s❯%s %s\n", cyan, reset, command)
	}
	fmt.Println()

	return 0
}

func nextStepCommands(m model) []string {
	commands := []string{fmt.Sprintf("cd %q", m.folderName)}

	pythonCommand := strings.Join(systemPythonCommand(), " ")
	if m.createPythonVenv {
		pythonCommand = virtualEnvironmentPython()
	}

	switch {
	case m.selectedStack.python && !m.createPythonVenv:
		commands = append(commands, pythonCommand+" -m pip install -r requirements.txt")
		if m.selectedStack.value == "Django" {
			commands = append(commands, pythonCommand+" manage.py migrate", pythonCommand+" manage.py runserver")
		} else {
			commands = append(commands, pythonCommand+" main.py")
		}

	case m.selectedStack.node:
		commands = append(commands, fmt.Sprintf("%s run %s", m.packageManager, m.selectedStack.startScript))

	case m.selectedStack.value == "Go_http":
		commands = append(commands, "go run ./cmd/api")

	case m.selectedStack.value == "Fiber_Sqlc":
		commands = append(commands, "go run .")

	case m.selectedStack.value == "Django":
		commands = append(commands, pythonCommand+" manage.py migrate", pythonCommand+" manage.py runserver")

	case m.selectedStack.value == "FastAPI":
		commands = append(commands, pythonCommand+" main.py")
	}

	return commands
}

func renderLayout(content string, terminalWidth int) string {
	contentLines := strings.Split(strings.TrimRight(content, "\n"), "\n")
	if terminalWidth > 4 {
		for index, line := range contentLines {
			contentLines[index] = ansi.Truncate(line, terminalWidth-2, "…")
		}
	}
	return "\n" + strings.Join(contentLines, "\n") + "\n"
}

func activityDetail(label string, animation int) string {
	details := []string{"Working", "Still working"}
	switch {
	case strings.Contains(label, "project from template"):
		details = []string{"Copying bundled starter files", "Preserving the template structure"}
	case strings.Contains(label, "Python virtual environment"):
		details = []string{"Preparing an isolated Python environment", "Setting up the Python interpreter"}
	case strings.Contains(label, "Python dependencies"):
		details = []string{"Resolving Python requirements", "Downloading and installing Python packages"}
	case strings.Contains(label, "dependencies with"):
		details = []string{"Resolving the package graph", "Downloading and linking packages", "The package manager is still working"}
	case strings.Contains(label, "Go dependencies"):
		details = []string{"Fetching Go modules", "Verifying downloaded Go modules"}
	case strings.Contains(label, ".gitignore"):
		details = []string{"Finalizing environment-safe defaults"}
	case strings.Contains(label, "Prisma"):
		details = []string{"Generating the typed Prisma client", "Finalizing Prisma artifacts"}
	}

	detail := details[(animation/18)%len(details)]
	dots := strings.Repeat(".", (animation/4)%4)
	return detail + dots
}

func formatDuration(duration time.Duration) string {
	if duration < time.Second {
		return duration.Round(100 * time.Millisecond).String()
	}
	if duration < time.Minute {
		return duration.Round(100 * time.Millisecond).String()
	}
	return duration.Round(time.Second).String()
}

func stackNames() []string {
	names := make([]string, len(stackChoices))
	for index, stack := range stackChoices {
		names[index] = stack.name
	}
	return names
}

func writeChoices(view *strings.Builder, choices []string, selected int) {
	for index, choice := range choices {
		if index == selected {
			fmt.Fprintf(view, "  %s› %s%s\n", cyan+bold, choice, reset)
		} else {
			fmt.Fprintf(view, "    %s\n", choice)
		}
	}
}

func selectionSummary(m model) string {
	var summary strings.Builder
	if m.selectedStack.name != "" {
		fmt.Fprintf(&summary, "  %sStack:%s %s\n", gray, reset, m.selectedStack.name)
	}
	if m.packageManager != "" {
		fmt.Fprintf(&summary, "  %sPackage manager:%s %s\n", gray, reset, m.packageManager)
	}
	if m.selectedStack.python && m.screen != pythonVenvScreen {
		venv := "No"
		if m.createPythonVenv {
			venv = "Yes"
		}
		fmt.Fprintf(&summary, "  %sPython virtual environment:%s %s\n", gray, reset, venv)
	}
	if m.targetDir != "" {
		fmt.Fprintf(&summary, "  %sProject:%s %s\n", gray, reset, m.targetDir)
	}
	if summary.Len() > 0 {
		summary.WriteString("\n")
	}
	return summary.String()
}

func wrapCursor(position, length int) int {
	if position < 0 {
		return length - 1
	}
	if position >= length {
		return 0
	}
	return position
}

func printableText(value string) string {
	return strings.Map(func(character rune) rune {
		if unicode.IsPrint(character) {
			return character
		}
		return -1
	}, value)
}

func tailLines(value string, maximum int) string {
	lines := strings.Split(value, "\n")
	if len(lines) > maximum {
		lines = lines[len(lines)-maximum:]
	}
	return strings.Join(lines, "\n")
}

func indent(value, prefix string) string {
	return prefix + strings.ReplaceAll(value, "\n", "\n"+prefix)
}

func main() {
	verbose := flag.Bool("verbose", false, "show command output during setup")
	flag.BoolVar(verbose, "v", false, "show command output during setup")
	showVersion := flag.Bool("version", false, "print the LetsInit version")
	flag.Parse()
	if *showVersion {
		fmt.Println("letsinit", version)
		return
	}

	initialModel := newModel(*verbose)
	finalModel, err := tea.NewProgram(initialModel).Run()
	if err != nil {
		fmt.Fprintln(os.Stderr, red+"Error starting LetsInit:"+reset, err)
		os.Exit(1)
	}

	result, ok := finalModel.(model)
	if !ok {
		fmt.Fprintln(os.Stderr, red+"Error during setup:"+reset, "unexpected application state")
		os.Exit(1)
	}
	if exitCode := printResult(result); exitCode != 0 {
		os.Exit(exitCode)
	}
}
