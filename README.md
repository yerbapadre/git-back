# git-back

A fast, interactive CLI tool to navigate your git branch history and quickly switch between recently used branches.

## Features

- 📜 **Branch History**: Parses git reflog to show your recent branch checkouts
- ⌨️ **Interactive Navigation**: Arrow keys to navigate, Enter to checkout
- 🌳 **Worktree Support**: Detects branches checked out in worktrees with smart handling
- 🎨 **Clean UI**: Minimal, keyboard-driven interface using Bubble Tea
- ⚡ **Fast**: Loads in <100ms for typical repos, handles 1000+ reflog entries
- 🛡️ **Safe**: Respects git checkout constraints (dirty working tree, etc.)

## Installation

### Build from Source

```bash
go install github.com/jakeevans/git-back@latest
```

Or clone and build locally:

```bash
git clone https://github.com/jakeevans/git-back.git
cd git-back
go build -o git-back
mv git-back /usr/local/bin/  # or anywhere in your PATH
```

## Usage

Navigate to any git repository and run:

```bash
git-back
```

### Keyboard Controls

- `↑` / `↓` or `k` / `j`: Navigate up/down
- `Enter`: Checkout selected branch
- `Esc` or `Ctrl+C`: Cancel and exit

### What It Shows

- Shows up to 20 most recently checked-out branches
- Each branch appears only once (deduplicated)
- Branches are ordered by most recent checkout first
- Current branch is excluded from the list
- Empty list shows: "No recent branches found"

### Worktree Support

When a branch is checked out in a git worktree, `git-back` displays it with a muted annotation showing the worktree directory name:

```
▶ feature/new-ui (-- checked out at worktree feature-work)
```

Pressing Enter on a worktree branch shows an options menu:
- **Navigate to worktree**: Prints a `cd` command you can copy/paste
- **Remove worktree and checkout branch**: Safely removes the worktree and checks out the branch
  - Checks for uncommitted changes first
  - Fails safely if worktree has unsaved work
  - Shown in red text as a warning

## How It Works

`git-back` parses your git reflog to find branch checkout events, deduplicates them, and presents them in an interactive list. When you press Enter, it runs `git checkout <branch>` with the selected branch.

## Error Handling

- **Not a git repository**: Shows error if run outside a git repo
- **Git not installed**: Shows error if git command is not found
- **Checkout failures**: Preserves and displays git's error messages

## Requirements

- Git installed and in PATH
- Go 1.21+ (for building from source)

## Examples

```bash
# Navigate to your repo
cd my-project

# Run git-back
git-back

# Select a branch with arrows and press Enter
# You're now on that branch!
```

## Roadmap

Future versions may include:
- Branch metadata (last checkout time, last commit message)
- Fuzzy search filtering
- Configuration options (limit, show all branches)
- Preview pane with git log

## License

MIT

## Contributing

Contributions welcome! Please open an issue or PR.
