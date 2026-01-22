# Security Policy

## Supported Versions

| Version | Supported          |
| ------- | ------------------ |
| latest  | :white_check_mark: |

## Current Security Baseline

**Last Updated:** 2026-01-21

### Runtime Environment

- **Go Version:** 1.25.6 (golang:1.25-alpine)
- **Alpine Linux:** 3.23.2
- **Busybox:** 1.37.0-r9

### Known CVEs - RESOLVED ✅

- CVE-2025-60876 (busybox) - **RESOLVED** by Alpine 3.21 → 3.23 upgrade
- Multiple golang/stdlib CVEs - **RESOLVED** by Go 1.21 → 1.23 upgrade

### Go Vulnerability Status

```bash
$ govulncheck ./...
No vulnerabilities found.
```

**Last Scan:** 2026-01-21

## Security Scanning

### Automated Scanning

The repository uses automated security scanning via GitHub Actions:

- **Schedule:** Weekly scans every Monday at 9 AM UTC
- **Triggers:** Push to main/master, pull requests, manual workflow dispatch
- **Tools:** Trivy (Docker + dependencies), govulncheck (Go code), GitHub Dependency Review

### Scan Coverage

1. **Filesystem Scan** - Scans repository for vulnerabilities in dependencies
2. **Docker Image Scan** - Scans built Docker image for OS-level and application vulnerabilities
3. **Go Code Scan** - Uses `govulncheck` to detect vulnerable Go packages
4. **Dependency Review** - Reviews dependency changes in pull requests

### Issue Creation

When critical or high severity vulnerabilities are detected with available fixes, the workflow automatically creates GitHub issues with:

- **Title:** `[Security] Upgrade [package] from [current] to [fixed]`
- **Labels:** `security-critical` or `security-high`, `security`, `dependencies`
- **Details:** CVE IDs, severity levels, descriptions, fix instructions

### Manual Scanning

#### Go Vulnerabilities

```bash
# Install govulncheck
go install golang.org/x/vuln/cmd/govulncheck@latest

# Scan codebase
govulncheck ./...
```

#### Docker Image (requires Trivy in CI/CD due to corporate proxy)

```bash
# In GitHub Actions or environments without proxy issues
docker run --rm -v /var/run/docker.sock:/var/run/docker.sock \
  aquasec/trivy image txt2json:latest
```

## Reporting a Vulnerability

If you discover a security vulnerability, please:

1. **DO NOT** open a public GitHub issue
2. Email security concerns to: [your-security-email@example.com]
3. Include:
   - Description of the vulnerability
   - Steps to reproduce
   - Potential impact
   - Suggested fix (if any)

We will respond within 48 hours and provide a timeline for fixes.

## Security Best Practices

### For Contributors

1. **Dependencies:**
   - Keep dependencies up to date
   - Review dependency changes in pull requests
   - Use `go mod tidy` to clean unused dependencies

2. **Secrets:**
   - Never commit secrets, API keys, or passwords
   - Use environment variables for sensitive configuration
   - Review `.env.example` for required variables

3. **Docker:**
   - Always use specific version tags (not `latest`)
   - Use multi-stage builds to minimize attack surface
   - Run containers as non-root users when possible

4. **Code:**
   - Follow OWASP secure coding guidelines
   - Validate all input data
   - Use parameterized queries (if applicable)
   - Handle errors securely (no sensitive data in error messages)

### For Deployment

1. **Environment Variables:**
   - Set strong credentials for RabbitMQ (`RABBITMQ_DEFAULT_USER`, `RABBITMQ_DEFAULT_PASS`)
   - Restrict file system permissions on mounted volumes
   - Use read-only mounts where possible

2. **Network:**
   - Restrict exposed ports to necessary services only
   - Use firewall rules to limit access
   - Enable TLS for RabbitMQ in production

3. **Monitoring:**
   - Monitor container logs for suspicious activity
   - Set up alerts for security-related events
   - Regularly review access logs

## Security Update Process

When a security issue is identified:

1. **Assessment** - Evaluate severity and impact
2. **Fix Development** - Develop and test fix
3. **Testing** - Run full test suite + security scans
4. **Documentation** - Update SECURITY.md with resolved CVEs
5. **Release** - Tag release with security notes
6. **Notification** - Update GitHub security advisory if applicable

## Version Upgrade Guidelines

### Go Version Upgrades

When upgrading Go:

1. Update `go.mod`: `go 1.XX`
2. Update `Dockerfile` builder stage: `FROM golang:1.XX.Y-alpineZ.ZZ`
3. Run tests: `go test ./... -v`
4. Rebuild Docker image: `docker-compose build --no-cache`
5. Scan for vulnerabilities: `govulncheck ./...`

### Alpine Version Upgrades

When upgrading Alpine:

1. Update `Dockerfile` runtime stage: `FROM alpine:Z.ZZ`
2. Rebuild: `docker-compose build --no-cache`
3. Test services: `docker-compose up -d && docker-compose logs`
4. Verify Alpine version: `docker exec txt2json-service cat /etc/alpine-release`

### Recent Upgrades

| Date       | Component | From      | To       | Reason                   |
|------------|-----------|-----------|----------|-------------------==-----|
| 2026-01-21 | Alpine    | 3.21      | 3.23.2   | CVE-2025-60876 (busybox) |
| 2026-01-21 | Go        | 1.21.13   | 1.25.6   | Multiple stdlib CVEs     |

## Compliance

This project follows:

- OWASP Secure Coding Practices
- CIS Docker Benchmarks (where applicable)
- Go Security Best Practices

## Security Contacts

- **Maintainer:** [Your Name/Team]
- **Security Email:** [security@example.com]
- **GitHub:** [@techie2000](https://github.com/techie2000)
