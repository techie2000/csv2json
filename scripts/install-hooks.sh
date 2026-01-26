#!/bin/bash
#
# Install Git hooks for the csv2json project
# Run this script after cloning the repository
#

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
HOOKS_DIR="$PROJECT_ROOT/.git/hooks"

echo "📦 Installing Git hooks..."

# Check if .git directory exists
if [ ! -d "$PROJECT_ROOT/.git" ]; then
    echo "❌ Not a Git repository. Run this from the project root."
    exit 1
fi

# Create pre-commit hook
cat > "$HOOKS_DIR/pre-commit" << 'EOF'
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
EOF

# Make hook executable
chmod +x "$HOOKS_DIR/pre-commit"

echo "✅ Git hooks installed successfully!"
echo ""
echo "Installed hooks:"
echo "  - pre-commit: Runs linting and formatting checks"
echo ""
echo "To bypass hooks temporarily, use: git commit --no-verify"
