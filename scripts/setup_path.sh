#!/bin/bash

# Get the user's shell configuration file
SHELL_CONFIG="${HOME}/.zshrc"
if [ ! -f "$SHELL_CONFIG" ]; then
    SHELL_CONFIG="${HOME}/.bashrc"
fi

# Add GOPATH/bin to PATH if not already present
GOPATH=$(go env GOPATH)
if ! grep -q "${GOPATH}/bin" "$SHELL_CONFIG"; then
    echo "" >> "$SHELL_CONFIG"
    echo "# Add Go binaries to PATH" >> "$SHELL_CONFIG"
    echo "export PATH=\"\${PATH}:${GOPATH}/bin\"" >> "$SHELL_CONFIG"
    echo "Added ${GOPATH}/bin to PATH in $SHELL_CONFIG"
    echo "Please run: source $SHELL_CONFIG"
else
    echo "${GOPATH}/bin is already in PATH"
fi