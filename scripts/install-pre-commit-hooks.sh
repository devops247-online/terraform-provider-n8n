#!/bin/bash

# install-pre-commit-hooks.sh
# Script to install pre-commit hooks for terraform-provider-n8n

set -e

echo "🚀 Installing pre-commit hooks for terraform-provider-n8n..."

# Check if pre-commit is installed
if ! command -v pre-commit &> /dev/null; then
    echo "❌ pre-commit is not installed."
    echo "📦 Please install pre-commit first:"
    echo "   - Via pip: pip install pre-commit"
    echo "   - Via brew (macOS): brew install pre-commit"
    echo "   - Via apt (Ubuntu): sudo apt install pre-commit"
    echo "   - Via conda: conda install -c conda-forge pre-commit"
    exit 1
fi

# Check if we're in a git repository
if [ ! -d ".git" ]; then
    echo "❌ This script must be run from the root of a git repository."
    exit 1
fi

# Install the git hook scripts
echo "🔧 Installing pre-commit git hooks..."
pre-commit install

# Install commit-msg hook for conventional commits (optional)
pre-commit install --hook-type commit-msg

# Install pre-push hook (optional)
pre-commit install --hook-type pre-push

# Run pre-commit on all files to check setup
echo "🧪 Running pre-commit on all files to verify setup..."
if pre-commit run --all-files; then
    echo "✅ Pre-commit hooks installed successfully!"
    echo ""
    echo "🎉 Setup complete! Pre-commit hooks will now run automatically on:"
    echo "   - git commit (pre-commit hooks)"
    echo "   - git push (pre-push hooks)"
    echo ""
    echo "📝 To manually run pre-commit on all files:"
    echo "   pre-commit run --all-files"
    echo ""
    echo "🔧 To update hooks to latest versions:"
    echo "   pre-commit autoupdate"
else
    echo "⚠️  Pre-commit hooks installed, but some checks failed."
    echo "   Please fix the issues above and run: pre-commit run --all-files"
fi
