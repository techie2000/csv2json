# Install Git hooks for the csv2json project (PowerShell version)
# Run this script after cloning the repository on Windows

$ErrorActionPreference = "Stop"

$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$ProjectRoot = Split-Path -Parent $ScriptDir
$HooksDir = Join-Path $ProjectRoot ".git\hooks"

Write-Host "📦 Installing Git hooks..." -ForegroundColor Cyan

# Check if .git directory exists
if (-not (Test-Path (Join-Path $ProjectRoot ".git"))) {
    Write-Host "❌ Not a Git repository. Run this from the project root." -ForegroundColor Red
    exit 1
}

# Create pre-commit hook
$PreCommitContent = @'
#!/bin/sh
#
# Pre-commit hook that runs linting before allowing commit
# This ensures code quality standards are maintained
#

echo "🔍 Running pre-commit checks..."

# Check if golangci-lint is installed
if ! command -v golangci-lint &> /dev/null; then
    echo "❌ golangci-lint is not installed"
    echo "   Installing golangci-lint..."
    go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
fi

# Run fast linting (skips slow checks for quick commits)
echo "Running golangci-lint (fast mode)..."
if ! make lint-fast 2>/dev/null; then
    # If make is not available, run golangci-lint directly
    if ! golangci-lint run ./... --fast; then
        echo ""
        echo "❌ Linting failed! Please fix the issues above before committing."
        echo "   Run 'make lint' or 'golangci-lint run ./...' for detailed analysis"
        echo "   Run 'golangci-lint run --fix' to auto-fix some issues"
        echo ""
        echo "To commit anyway (not recommended), use: git commit --no-verify"
        exit 1
    fi
fi

# Check for Go formatting
echo "Checking Go formatting..."
UNFORMATTED=$(gofmt -l . | grep -v '^vendor/')
if [ -n "$UNFORMATTED" ]; then
    echo "❌ The following files are not formatted:"
    echo "$UNFORMATTED"
    echo ""
    echo "Run 'go fmt ./...' or 'make fmt' to fix formatting"
    exit 1
fi

echo "✅ All pre-commit checks passed!"
exit 0
'@

$PreCommitPath = Join-Path $HooksDir "pre-commit"
Set-Content -Path $PreCommitPath -Value $PreCommitContent -NoNewline

Write-Host "✅ Git hooks installed successfully!" -ForegroundColor Green
Write-Host ""
Write-Host "Installed hooks:"
Write-Host "  - pre-commit: Runs linting and formatting checks"
Write-Host ""
Write-Host "To bypass hooks temporarily, use: git commit --no-verify"
