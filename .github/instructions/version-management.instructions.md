---
applyTo: '**/VERSION,**/version.go,**/.github/**'
description: 'Version management and release preparation instructions for csv2json project'
---

# Version Management Instructions

## Core Principle
**Version tracking is critical for release management, git tagging, and production deployments.**

## Semantic Versioning (SemVer)

The project follows [Semantic Versioning 2.0.0](https://semver.org/):

**Format:** `MAJOR.MINOR.PATCH`

- **MAJOR**: Incompatible API changes or breaking changes
- **MINOR**: New features, backward-compatible functionality
- **PATCH**: Bug fixes, backward-compatible patches

### Version Increment Guidelines

#### When to Bump MAJOR (X.0.0)
- Breaking changes to CLI arguments or flags
- Incompatible changes to configuration format (.env variables)
- Breaking changes to output JSON structure
- Removal of deprecated features
- Changes that require users to modify their integration

#### When to Bump MINOR (0.X.0)
- New features (new configuration options, output formats)
- New command-line flags (backward compatible)
- Performance improvements (significant)
- New archiving categories or behaviors
- Deprecation notices (without removal)

#### When to Bump PATCH (0.0.X)
- Bug fixes
- Security patches
- Documentation updates (significant)
- Internal refactoring (no external impact)
- Performance improvements (minor)
- Dependency updates

## Version Update Workflow

### Files That MUST Be Updated Together

1. **`VERSION`** (root file)
   - Plain text file containing only the version number
   - Example: `0.1.0`

2. **`internal/version/version.go`**
   - Update the `Version` constant
   - Must match the `VERSION` file exactly
   ```go
   const Version = "0.1.0"
   ```

### Automated Version Injection

The following are automatically injected at build time via `-ldflags`:
- `internal/version.GitCommit` - Git commit hash
- `internal/version.BuildDate` - Build timestamp

**DO NOT** manually edit these in the source code.

## Release Preparation Checklist

When incrementing the version for a release:

1. **Update Version Files**
   - [ ] Update `VERSION` file with new version number
   - [ ] Update `internal/version/version.go` constant to match
   - [ ] Commit these changes with message: `chore: bump version to X.Y.Z`

2. **Create Git Tag**
   ```bash
   git tag -a vX.Y.Z -m "Release vX.Y.Z"
   git push origin vX.Y.Z
   ```

3. **Build Release Binaries**
   ```bash
   make build-all  # Cross-compile for all platforms
   ```

4. **Verify Version Information**
   ```bash
   ./csv2json -version  # Should show correct version
   ```

5. **Update CHANGELOG** (if exists)
   - Document changes, features, and fixes

## Version Checking

### During Development
```bash
# Check version without building
cat VERSION

# Check version in code
grep 'const Version' internal/version/version.go
```

### In Built Binary
```bash
./csv2json -version
# Output: csv2json v0.1.0 (commit: abc1234) (built: 2026-01-22T12:34:56Z)
```

### In Running Service
- Version is logged on service startup
- Check logs: `logs/csv2json.log`

## Docker Build with Version

When building Docker images, pass version information as build args:

```bash
docker build \
  --build-arg VERSION=$(cat VERSION) \
  --build-arg GIT_COMMIT=$(git rev-parse --short HEAD) \
  --build-arg BUILD_DATE=$(date -u +"%Y-%m-%dT%H:%M:%SZ") \
  -t csv2json:$(cat VERSION) \
  -t csv2json:latest \
  .
```

## Critical Rules

1. **NEVER** commit version updates without updating BOTH files
2. **ALWAYS** use semantic versioning rules
3. **ALWAYS** create a git tag for releases
4. **ALWAYS** document breaking changes in commit messages
5. **ALWAYS** verify version with `-version` flag after building

## Common Mistakes to Avoid

- ❌ Updating VERSION file but forgetting version.go
- ❌ Updating version.go but forgetting VERSION file
- ❌ Using incorrect SemVer format (e.g., `v1.0` instead of `1.0.0`)
- ❌ Forgetting to tag releases in git
- ❌ Bumping MAJOR for non-breaking changes
- ❌ Forgetting to update CHANGELOG

## Pre-Release Versions

For alpha, beta, or release candidate versions:

```
1.0.0-alpha.1
1.0.0-beta.1
1.0.0-rc.1
```

## Integration with CI/CD

When setting up automated releases:

1. Read version from `VERSION` file
2. Inject via ldflags during build
3. Tag Docker images with version
4. Create GitHub releases with version tag

## Quick Reference

| Scenario | Example | Next Version |
|----------|---------|--------------|
| Bug fix | Fix delimiter parsing | 0.1.0 → 0.1.1 |
| New feature | Add Kafka output | 0.1.1 → 0.2.0 |
| Breaking change | Change CLI flags | 0.2.0 → 1.0.0 |
| Security patch | Fix CVE | 1.0.0 → 1.0.1 |
| New minor feature | Add --quiet flag | 1.0.1 → 1.1.0 |

## Version History Strategy

- Keep a CHANGELOG.md following [Keep a Changelog](https://keepachangelog.com/)
- Document all changes under appropriate categories:
  - Added
  - Changed
  - Deprecated
  - Removed
  - Fixed
  - Security
