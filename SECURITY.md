# Security

## Reporting Security Issues

If you discover a security vulnerability, please email: [your-email] (DO NOT open a public issue)

## Security Measures

### Installation Security

**Checksum Verification**
- All release binaries include SHA-256 checksums in `checksums.txt`
- The install script automatically verifies checksums if `shasum` is available
- Manual verification:
  ```bash
  shasum -a 256 -c checksums.txt
  ```

**Safer Installation**
Instead of piping curl to bash, review the script first:
```bash
# Download and inspect
curl -sL https://raw.githubusercontent.com/yerbapadre/git-back/main/install.sh -o install.sh
cat install.sh  # Review the script
bash install.sh
```

### Code Security

**Git Command Execution**
- All git commands use `exec.Command()` with separate arguments (no shell interpretation)
- Branch names come from git's reflog output, not user input
- Worktree paths come from `git worktree list`, a trusted source

**Path Handling**
- Uses `filepath.Base()` for extracting directory names (no path traversal)
- Temporary directories use `mktemp -d` for secure temp file creation
- Install script uses `trap` to cleanup temp files even on errors

**Input Validation**
- Branch names are validated by git itself before checkout
- Uncommitted changes are checked before destructive operations
- All git errors are captured and displayed safely (no raw stderr)

**Clipboard Access**
- Supports multiple clipboard utilities: pbcopy (macOS), xclip, xsel, wl-copy (Linux)
- Fails gracefully if no clipboard utility is available
- Only copies worktree paths (user-controlled, non-sensitive data)

### Binary Distribution

**Cross-Platform Builds**
- Built with official Go compiler
- No CGO dependencies (pure Go = no native code vulnerabilities)
- Reproducible builds from source

**Release Process**
1. Build binaries for all platforms using `build-release.sh`
2. Generate SHA-256 checksums
3. Sign and upload to GitHub Releases
4. Users verify checksums during installation

## Threat Model

### What git-back protects against:
- ✅ Command injection (uses exec.Command safely)
- ✅ Path traversal (validates paths)
- ✅ Binary tampering (checksum verification)
- ✅ Uncommitted changes loss (checks before checkout)

### What git-back does NOT protect against:
- ❌ Compromised git executable on your system
- ❌ Malicious git repository contents
- ❌ Man-in-the-middle attacks on git remotes (use git over SSH/HTTPS)

### Known Limitations

**Install Script Trust**
- Running `curl | bash` requires trusting GitHub and this repository
- Consider reviewing the script before execution
- Use `go install` instead for reproducible builds from source

**Clipboard Security**
- Worktree paths are copied to clipboard (visible to other apps)
- Only copies paths when explicitly requested (not automatic)

## Dependencies

git-back uses minimal, well-vetted dependencies:
- `github.com/charmbracelet/bubbletea` - TUI framework (audited, popular)
- `github.com/charmbracelet/lipgloss` - Styling library (same maintainer)
- Standard Go library only (no other third-party deps)

Run `go mod verify` to verify dependency checksums.

## Best Practices

1. **Always review scripts before running them**
2. **Verify checksums of downloaded binaries**
3. **Keep git-back updated** for security patches
4. **Use SSH for git remotes** to prevent MITM attacks
5. **Don't run git-back with elevated privileges** (no sudo needed)

## Audit History

- 2026-04-02: Initial security audit, added checksum verification
