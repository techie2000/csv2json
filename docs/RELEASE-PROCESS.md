# Release Process Quick Reference

This document provides a quick reference for creating releases. For detailed instructions, see [version-management.instructions.md](../.github/instructions/version-management.instructions.md).

## Automated Release Process

### ✨ Just push a tag - everything else is automatic! ✨

```bash
# 1. Update CHANGELOG.md (move [Unreleased] items to new version section)
# 2. Bump version files
echo "0.3.0" > VERSION
# Update internal/version/version.go to match

# 3. Commit and push
git add CHANGELOG.md VERSION internal/version/version.go
git commit -m "chore: bump version to 0.3.0"
git push origin main

# 4. Create and push tag
git tag -a v0.3.0 -m "Release v0.3.0 - Brief description"
git push origin v0.3.0
```

**That's it!** GitHub Actions will automatically:

- ✅ Build binaries for all platforms
- ✅ Generate SHA256 checksums
- ✅ Create GitHub Release with CHANGELOG notes
- ✅ Upload binaries to release
- ✅ Build and push Docker images to ghcr.io
- ✅ Tag Docker images with version and "latest"

## What Gets Built Automatically

### Binaries (5 platforms)

- `csv2json-linux-amd64` + `.sha256`
- `csv2json-linux-arm64` + `.sha256`
- `csv2json-windows-amd64.exe` + `.sha256`
- `csv2json-darwin-amd64` + `.sha256`
- `csv2json-darwin-arm64` + `.sha256`

### Docker Images (multi-arch)

- `ghcr.io/techie2000/csv2json:v0.3.0`
- `ghcr.io/techie2000/csv2json:0.3`
- `ghcr.io/techie2000/csv2json:0`
- `ghcr.io/techie2000/csv2json:latest`
- Architectures: `linux/amd64`, `linux/arm64`

## Monitoring the Release

1. **GitHub Actions**: <https://github.com/techie2000/csv2json/actions>
   - Watch the "Release" workflow
   - Typically completes in 5-10 minutes

2. **GitHub Releases**: <https://github.com/techie2000/csv2json/releases>
   - Release should appear with binaries attached
   - Release notes extracted from CHANGELOG.md

3. **Docker Images**: <https://github.com/techie2000/csv2json/pkgs/container/csv2json>
   - Images should appear with version tags
   - First time: May need to make package public in settings

## Testing the Release

### Test Binaries

```bash
# Download
curl -LO https://github.com/techie2000/csv2json/releases/download/v0.3.0/csv2json-linux-amd64

# Verify checksum
curl -LO https://github.com/techie2000/csv2json/releases/download/v0.3.0/csv2json-linux-amd64.sha256
sha256sum -c csv2json-linux-amd64.sha256

# Test
chmod +x csv2json-linux-amd64
./csv2json-linux-amd64 -version
```

### Test Docker Image

```bash
# Pull
docker pull ghcr.io/techie2000/csv2json:v0.3.0

# Test
docker run --rm ghcr.io/techie2000/csv2json:v0.3.0 ./csv2json -version
```

## Pre-Release Versions

For alpha, beta, or release candidate versions:

```bash
git tag -a v1.0.0-alpha.1 -m "Release v1.0.0-alpha.1"
git push origin v1.0.0-alpha.1
```

GitHub Actions will automatically mark releases containing `-alpha`, `-beta`, or `-rc` as pre-releases.

## Rollback

If you need to delete a release:

```bash
# Delete tag locally
git tag -d v0.3.0

# Delete tag on remote
git push origin :refs/tags/v0.3.0

# Manually delete the GitHub Release in the UI
# https://github.com/techie2000/csv2json/releases
```

## Troubleshooting

### Release workflow failed

- Check GitHub Actions logs for errors
- Common issues:
  - CHANGELOG.md format incorrect (can't extract release notes)
  - Build errors in code
  - Docker build failures

### Docker images not appearing

- First release to ghcr.io? Package may be private by default
- Go to: Settings → Packages → csv2json → Change visibility → Public

### Binaries missing architecture

- Check the build matrix in `.github/workflows/release.yml`
- Ensure GOOS/GOARCH combinations are valid

## Version Numbering

Follow [Semantic Versioning](https://semver.org/):

- **MAJOR** (X.0.0): Breaking changes
- **MINOR** (0.X.0): New features (backward compatible)
- **PATCH** (0.0.X): Bug fixes (backward compatible)

Examples:

- Bug fix: `0.2.0` → `0.2.1`
- New feature: `0.2.1` → `0.3.0`
- Breaking change: `0.3.0` → `1.0.0`
