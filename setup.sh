#!/bin/bash

# ============================================================================
# Setup Script for Dudley Climate Justice Archive
# ============================================================================
#
# This script helps you get the archive running on your Mac by:
# 1. Installing Homebrew (if needed)
# 2. Installing Go (if needed)
# 3. Checking for code editors and offering to install VS Code
# 4. Creating your .env configuration file
# 5. Building and optionally launching the archive
#
# Run this with: bash setup.sh

set -e  # Exit if any command fails

echo ""
echo "============================================"
echo "Dudley Climate Justice Archive Setup"
echo "============================================"
echo ""

# ============================================================================
# Step 1: Check for Homebrew
# ============================================================================

echo "Step 1: Checking for Homebrew..."
echo ""

if command -v brew &> /dev/null; then
    echo "✓ Homebrew is already installed"
    echo ""
else
    echo "Homebrew is not installed. Installing Homebrew now..."
    echo ""
    echo "This will prompt you for your password and may take a few minutes."
    echo ""
    
    # Install Homebrew using the official installation script
    /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
    
    # Add Homebrew to PATH for the current session
    # The location differs on Intel vs Apple Silicon Macs
    if [[ -f "/opt/homebrew/bin/brew" ]]; then
        # Apple Silicon Mac
        eval "$(/opt/homebrew/bin/brew shellenv)"
    elif [[ -f "/usr/local/bin/brew" ]]; then
        # Intel Mac
        eval "$(/usr/local/bin/brew shellenv)"
    fi
    
    echo ""
    echo "✓ Homebrew installed successfully"
    echo ""
fi

# ============================================================================
# Step 2: Check for Go
# ============================================================================

echo "Step 2: Checking for Go..."
echo ""

if command -v go &> /dev/null; then
    GO_VERSION=$(go version)
    echo "✓ Go is already installed: $GO_VERSION"
    echo ""
else
    echo "Go is not installed."
    echo ""
    read -p "Would you like to install Go now? (y/n): " INSTALL_GO
    echo ""
    
    if [[ "$INSTALL_GO" =~ ^[Yy]$ ]]; then
        echo "Installing Go via Homebrew..."
        brew install go
        echo ""
        echo "✓ Go installed successfully"
        echo ""
    else
        echo ""
        echo "Go installation skipped. You'll need to install Go manually to run the archive."
        echo "Visit: https://go.dev/doc/install"
        echo ""
        exit 1
    fi
fi

# ============================================================================
# Step 3: Check for code editors
# ============================================================================

echo "Step 3: Checking for code editors..."
echo ""

# Check which editors are installed
VSCODE_INSTALLED=false
SUBLIME_INSTALLED=false
ATOM_INSTALLED=false

if command -v code &> /dev/null; then
    VSCODE_INSTALLED=true
    echo "✓ Visual Studio Code is installed"
fi

if [ -d "/Applications/Sublime Text.app" ]; then
    SUBLIME_INSTALLED=true
    echo "✓ Sublime Text is installed"
fi

if [ -d "/Applications/Atom.app" ]; then
    ATOM_INSTALLED=true
    echo "✓ Atom is installed"
fi

if [[ "$VSCODE_INSTALLED" == "false" && "$SUBLIME_INSTALLED" == "false" && "$ATOM_INSTALLED" == "false" ]]; then
    echo "No common code editors detected."
    echo ""
    echo "A code editor makes it much easier to work with the archive."
    echo ""
    read -p "Would you like to install Visual Studio Code? (y/n): " INSTALL_VSCODE
    echo ""
    
    if [[ "$INSTALL_VSCODE" =~ ^[Yy]$ ]]; then
        echo "Installing Visual Studio Code..."
        brew install --cask visual-studio-code
        echo ""
        echo "✓ Visual Studio Code installed successfully"
        echo ""
    else
        echo "Skipping editor installation."
        echo ""
    fi
else
    echo ""
fi

# ============================================================================
# Step 4: Create .env file
# ============================================================================

echo "Step 4: Setting up environment variables..."
echo ""
echo "The archive needs some configuration to connect to NocoDB."
echo "I'll ask you for each setting now."
echo ""

# Check if .env already exists
if [[ -f ".env" ]]; then
    echo "Warning: A .env file already exists."
    echo ""
    read -p "Do you want to overwrite it? (y/n): " OVERWRITE
    echo ""
    
    if [[ ! "$OVERWRITE" =~ ^[Yy]$ ]]; then
        echo "Keeping existing .env file."
        echo ""
        SKIP_ENV=true
    fi
fi

if [[ "$SKIP_ENV" != "true" ]]; then
    # Create .env file with user input
    > .env  # Create/clear the file
    
    echo "Enter your NocoDB endpoint URL"
    echo "(e.g., https://nocodb-r87d.onrender.com)"
    read -p "NOCODB_ENDPOINT: " NOCODB_ENDPOINT
    echo "NOCODB_ENDPOINT=$NOCODB_ENDPOINT" >> .env
    echo ""
    
    echo "Enter your NocoDB API key"
    echo "(This is a long string like: noco-abc123...)"
    read -p "NOCODB_API_KEY: " NOCODB_API_KEY
    echo "NOCODB_API_KEY=$NOCODB_API_KEY" >> .env
    echo ""
    
    echo "Enter your NocoDB table ID"
    echo "(This is the ID of your Stories table)"
    read -p "NOCODB_TABLE_ID: " NOCODB_TABLE_ID
    echo "NOCODB_TABLE_ID=$NOCODB_TABLE_ID" >> .env
    echo ""
    
    echo "✓ Configuration saved to .env"
    echo ""
fi

# ============================================================================
# Step 5: Build and optionally run the archive
# ============================================================================

echo "Step 5: Building the archive..."
echo ""

# Build the archive binary
echo "Compiling the archive..."
go build -o archive ./cmd/archive

if [[ $? -eq 0 ]]; then
    echo ""
    echo "✓ Archive built successfully"
    echo ""
else
    echo ""
    echo "Build failed. Check the error messages above."
    exit 1
fi

echo "============================================"
echo "Setup Complete!"
echo "============================================"
echo ""
echo "The archive is ready to run."
echo ""
read -p "Would you like to launch the archive in development mode now? (y/n): " LAUNCH
echo ""

if [[ "$LAUNCH" =~ ^[Yy]$ ]]; then
    echo "Starting the archive in development mode (skipping image processing for faster startup)..."
    echo "This will open a server at http://localhost:8080"
    echo ""
    echo "Press Enter to regenerate pages when you change templates."
    echo "Press Ctrl+C to stop the server."
    echo ""
    echo "============================================"
    echo ""
    
    # Run the archive in development mode, skipping image processing for faster startup
    ./archive -d -s
else
    echo "You can start the archive anytime with:"
    echo ""
    echo "  Development mode (with live reload):"
    echo "    ./archive -d"
    echo ""
    echo "  Production mode (build once):"
    echo "    ./archive"
    echo ""
    echo "  Skip image processing (faster for testing):"
    echo "    ./archive -s"
    echo ""
fi

