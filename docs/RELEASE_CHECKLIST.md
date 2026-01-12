# Chauffeur Release Checklist

This checklist ensures consistent, high-quality releases for Chauffeur. Follow these steps for every release.

## Pre-Release Checklist

### 1. Code Quality ✅
- [ ] `go test ./...` passes locally with ≥80% coverage
- [ ] `go vet ./...` shows no issues
- [ ] `go fmt ./...` applied (code is formatted)
- [ ] `go mod tidy` run (dependencies clean)
- [ ] No compiled binaries or build artifacts in git
- [ ] All TODOs and FIXMEs addressed or documented

### 2. Documentation Sync ✅
- [ ] **README.md** reflects current features and installation steps
- [ ] **docs/TODO_STATUS.md** shows accurate project status
- [ ] **docs/README.md** matches actual command behavior and contracts
- [ ] **docs/CONTRIBUTING.md** guidelines are current
- [ ] Changelog entries prepared (following conventional commit format)

### 3. Feature Validation ✅
- [ ] All documented features work as described
- [ ] CLI commands match docs/README.md documentation
- [ ] Error messages are actionable and user-friendly
- [ ] Workspace rules followed (no host system mutations)
- [ ] Service orchestration works (start/stop/restart/status)

### 4. Testing & Validation ✅
- [ ] Fresh installation tested on clean system
- [ ] `chauf init` creates workspace correctly
- [ ] `chauf install php|nginx|composer` works
- [ ] `chauf link` generates proper nginx configs
- [ ] DNS resolution (.test domains) works
- [ ] All example commands from README tested

## Release Process

### 5. Version Bump
```bash
# Update version in CLI (if versioned)
# Update CHANGELOG.md with release notes
# Commit version changes
git add -A
git commit -m "chore(release): prepare v{version} release"
```

### 6. Tag Release
```bash
git tag -a v{version} -m "Release v{version}"
git push origin v{version}
```

### 7. Automated Build
- [ ] GitHub Actions Goreleaser workflow triggers
- [ ] Build completes successfully
- [ ] Release artifacts generated
- [ ] GitHub release created with changelog

## Post-Release Checklist

### 8. Release Verification ✅
- [ ] GitHub release published correctly
- [ ] Download links work
- [ ] Installation instructions tested with new release
- [ ] Release notes are accurate and complete

### 9. Documentation Updates ✅
- [ ] Update README.md with new version info
- [ ] Update docs/TODO_STATUS.md with release progress
- [ ] Move completed items from In Progress to Completed
- [ ] Update roadmap/priorities if needed

### 10. Community Communication ✅
- [ ] Announce release (if applicable)
- [ ] Update project documentation/sites
- [ ] Monitor for issues and feedback

## Build Configuration Details

### Current Setup
- **CI/CD**: GitHub Actions
- **Build Tool**: GoReleaser (currently changelog-only)
- **Platforms**: Linux (primary focus)
- **Architecture**: amd64, arm64 (if applicable)

### GoReleaser Configuration
Located in `.goreleaser.yml`:
- Currently configured for changelog generation only
- Pre-release mode enabled (`prerelease: 'true'`)
- Conventional commit grouping for changelog

### To Enable Binary Releases
Update `.goreleaser.yml` to include:
```yaml
builds:
  - env: [CGO_ENABLED=0]
    goos: [linux]
    goarch: [amd64, arm64]
    binary: chauf

archives:
  - format: tar.gz
    name_template: "{{ .ProjectName }}_{{ .Version }}_{{ .Os }}_{{ .Arch }}"
```

## Emergency Release Process

For critical fixes requiring immediate release:

1. **Skip non-essential steps**: Focus on bug fix and basic testing
2. **Patch version**: Increment patch version (x.y.z → x.y.z+1)
3. **Minimal changelog**: Document only the critical fix
4. **Fast-track**: Push tag directly, trigger build
5. **Post-release**: Complete full checklist within 24 hours

## Release Types

### Major Release (x.0.0)
- Breaking changes
- Significant new features
- Requires comprehensive testing
- Update all documentation

### Minor Release (x.y.0)
- New features (backward compatible)
- Enhancements
- Full testing cycle

### Patch Release (x.y.z)
- Bug fixes only
- Critical issues
- Focused testing on affected areas

## Troubleshooting

### Build Failures
- Check Go version compatibility (requires 1.22+)
- Verify all dependencies available
- Check GitHub Actions secrets

### Test Failures
- Run tests locally first
- Check for workspace isolation issues
- Verify test environment setup

### Documentation Issues
- Use `chauf --help` to verify command accuracy
- Test all examples in README
- Cross-reference with AGENTS.md

## Rolling Back

If a release has critical issues:

1. **Delete GitHub release** (if possible)
2. **Yank tag**: `git tag -d v{version} && git push origin :refs/tags/v{version}`
3. **Fix issues** and create new release
4. **Document rollback** in release notes

## Maintainer Notes

- **Primary Maintainer**: @si-aji
- **Release Cadence**: As-needed based on feature completion
- **Supported Platforms**: Linux (Arch/Ubuntu/Debian primary)
- **Go Version**: 1.22+ required
- **Testing Environment**: Clean Linux workspace preferred

Remember: **AGENTS.md is the single source of truth** - always verify commands and behavior against it during release preparation.