package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	selectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("170")).
			Bold(true).
			PaddingLeft(2)

	normalStyle = lipgloss.NewStyle().
			PaddingLeft(4)

	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("99")).
			MarginBottom(1)

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Bold(true)

	mutedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Faint(true)

	menuTitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("99"))

	dangerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196"))
)

type worktreeInfo struct {
	path    string
	dirName string
}

type model struct {
	branches    []string
	cursor      int
	err         error
	worktrees   map[string]worktreeInfo
	showingMenu bool
	menuCursor  int
}

func initialModel() (model, error) {
	branches, err := getRecentBranches()
	if err != nil {
		return model{err: err}, err
	}

	worktrees := getWorktrees()

	return model{
		branches:    branches,
		cursor:      0,
		worktrees:   worktrees,
		showingMenu: false,
		menuCursor:  0,
	}, nil
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.showingMenu {
			return m.handleMenuInput(msg)
		}

		switch msg.String() {
		case "ctrl+c", "esc":
			return m, tea.Quit

		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			} else {
				m.cursor = len(m.branches) - 1
			}

		case "down", "j":
			if m.cursor < len(m.branches)-1 {
				m.cursor++
			} else {
				m.cursor = 0
			}

		case "enter":
			if len(m.branches) > 0 {
				branch := m.branches[m.cursor]

				if _, hasWorktree := m.worktrees[branch]; hasWorktree {
					m.showingMenu = true
					m.menuCursor = 0
					return m, nil
				}

				err := checkoutBranch(branch)
				if err != nil {
					m.err = err
					return m, nil
				}
				fmt.Printf("Switched to branch '%s'\n", branch)
				return m, tea.Quit
			}
		}
	}
	return m, nil
}

func (m model) handleMenuInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "esc":
		m.showingMenu = false
		m.menuCursor = 0
		return m, nil

	case "up", "k":
		if m.menuCursor > 0 {
			m.menuCursor--
		} else {
			m.menuCursor = 1
		}

	case "down", "j":
		if m.menuCursor < 1 {
			m.menuCursor++
		} else {
			m.menuCursor = 0
		}

	case "enter":
		branch := m.branches[m.cursor]
		worktree := m.worktrees[branch]

		if m.menuCursor == 0 {
			fmt.Printf("\ncd %s\n\n", worktree.path)

			if err := copyToClipboard(worktree.path); err == nil {
				fmt.Println("(Path copied to clipboard)")
			}

			return m, tea.Quit
		} else {
			err := removeWorktreeAndCheckout(branch, worktree.path)
			if err != nil {
				m.err = err
				m.showingMenu = false
				return m, nil
			}
			fmt.Printf("\nRemoved worktree and switched to branch '%s'\n", branch)
			return m, tea.Quit
		}
	}

	return m, nil
}

func (m model) View() string {
	if m.err != nil {
		return errorStyle.Render(fmt.Sprintf("Error: %v\n", m.err))
	}

	if len(m.branches) == 0 {
		return "No recent branches found.\n"
	}

	if m.showingMenu {
		return m.renderMenu()
	}

	s := titleStyle.Render("Recent Branches") + "\n\n"

	for i, branch := range m.branches {
		cursor := " "
		line := ""

		if i == m.cursor {
			cursor = "▶"
			line = cursor + " " + branch
		} else {
			line = cursor + " " + branch
		}

		if worktree, hasWorktree := m.worktrees[branch]; hasWorktree {
			annotation := mutedStyle.Render(fmt.Sprintf(" (-- checked out at worktree %s)", worktree.dirName))
			line += annotation
		}

		if i == m.cursor {
			s += selectedStyle.Render(line) + "\n"
		} else {
			s += normalStyle.Render(line) + "\n"
		}
	}

	s += "\n" + lipgloss.NewStyle().Faint(true).Render("↑/↓: navigate • enter: checkout • esc: quit")
	return s
}

func (m model) renderMenu() string {
	branch := m.branches[m.cursor]
	worktree := m.worktrees[branch]

	s := menuTitleStyle.Render(fmt.Sprintf("Branch '%s' is checked out in worktree", branch)) + "\n"
	s += mutedStyle.Render(fmt.Sprintf("Location: %s", worktree.path)) + "\n\n"

	options := []string{
		"Navigate to worktree",
		"Remove worktree and checkout branch",
	}

	for i, option := range options {
		cursor := " "
		if i == m.menuCursor {
			cursor = "▶"
			if i == 1 {
				s += dangerStyle.Render(cursor+" "+option) + "\n"
			} else {
				s += selectedStyle.Render(cursor+" "+option) + "\n"
			}
		} else {
			if i == 1 {
				s += dangerStyle.Render(cursor+" "+option) + "\n"
			} else {
				s += normalStyle.Render(cursor+" "+option) + "\n"
			}
		}
	}

	s += "\n" + lipgloss.NewStyle().Faint(true).Render("↑/↓: navigate • enter: select • esc: cancel")
	return s
}

func getRecentBranches() ([]string, error) {
	if !isGitRepo() {
		return nil, fmt.Errorf("not a git repository")
	}

	cmd := exec.Command("git", "reflog", "show", "--all", "--format=%gs")
	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("git reflog failed: %s", string(exitErr.Stderr))
		}
		return nil, fmt.Errorf("git reflog failed: %w", err)
	}

	currentBranch, err := getCurrentBranch()
	if err != nil {
		currentBranch = ""
	}

	seen := make(map[string]bool)
	var branches []string

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)

		if strings.HasPrefix(line, "checkout: moving from ") {
			parts := strings.Split(line, " to ")
			if len(parts) == 2 {
				branch := strings.TrimSpace(parts[1])

				if branch == "" || branch == currentBranch {
					continue
				}

				if strings.Contains(branch, " ") {
					continue
				}

				if !seen[branch] {
					seen[branch] = true
					branches = append(branches, branch)

					if len(branches) >= 20 {
						break
					}
				}
			}
		}
	}

	return branches, nil
}

func getCurrentBranch() (string, error) {
	cmd := exec.Command("git", "branch", "--show-current")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

func isGitRepo() bool {
	cmd := exec.Command("git", "rev-parse", "--git-dir")
	err := cmd.Run()
	return err == nil
}

func checkoutBranch(branch string) error {
	statusCmd := exec.Command("git", "status", "--porcelain")
	statusOutput, err := statusCmd.Output()
	if err != nil {
		return fmt.Errorf("failed to check git status: %w", err)
	}

	if len(strings.TrimSpace(string(statusOutput))) > 0 {
		modified := []string{}
		lines := strings.Split(strings.TrimSpace(string(statusOutput)), "\n")
		for i, line := range lines {
			if i >= 3 {
				modified = append(modified, "...")
				break
			}
			if len(line) > 2 {
				modified = append(modified, strings.TrimSpace(line[3:]))
			}
		}
		return fmt.Errorf("uncommitted changes would be overwritten:\n  %s\n\nCommit or stash your changes first", strings.Join(modified, "\n  "))
	}

	cmd := exec.Command("git", "checkout", branch)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("checkout failed: %s", strings.TrimSpace(string(output)))
	}
	return nil
}

func getWorktrees() map[string]worktreeInfo {
	worktrees := make(map[string]worktreeInfo)

	cmd := exec.Command("git", "worktree", "list", "--porcelain")
	output, err := cmd.Output()
	if err != nil {
		return worktrees
	}

	lines := strings.Split(string(output), "\n")
	var currentPath string

	for _, line := range lines {
		line = strings.TrimSpace(line)

		if strings.HasPrefix(line, "worktree ") {
			currentPath = strings.TrimPrefix(line, "worktree ")
		} else if strings.HasPrefix(line, "branch ") {
			branchRef := strings.TrimPrefix(line, "branch ")
			branchName := strings.TrimPrefix(branchRef, "refs/heads/")

			if currentPath != "" {
				dirName := filepath.Base(currentPath)
				worktrees[branchName] = worktreeInfo{
					path:    currentPath,
					dirName: dirName,
				}
				currentPath = ""
			}
		}
	}

	return worktrees
}

func copyToClipboard(text string) error {
	cmd := exec.Command("pbcopy")
	pipe, err := cmd.StdinPipe()
	if err != nil {
		return err
	}

	if err := cmd.Start(); err != nil {
		return err
	}

	if _, err := pipe.Write([]byte(text)); err != nil {
		return err
	}

	if err := pipe.Close(); err != nil {
		return err
	}

	return cmd.Wait()
}

func removeWorktreeAndCheckout(branch, worktreePath string) error {
	statusCmd := exec.Command("git", "-C", worktreePath, "status", "--porcelain")
	statusOutput, err := statusCmd.Output()
	if err != nil {
		return fmt.Errorf("failed to check worktree status: %w", err)
	}

	if len(strings.TrimSpace(string(statusOutput))) > 0 {
		return fmt.Errorf("worktree has uncommitted changes - commit or stash them first")
	}

	removeCmd := exec.Command("git", "worktree", "remove", worktreePath)
	output, err := removeCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to remove worktree: %s", strings.TrimSpace(string(output)))
	}

	return checkoutBranch(branch)
}

func main() {
	m, err := initialModel()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	p := tea.NewProgram(m)
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
