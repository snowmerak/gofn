# GoFn Development Commands

## Installation and Setup
```fish
# Install the CLI tool
go install ./cmd/gofn

# Or run without installing
go run ./cmd/gofn -src=. -out=.
```

## Code Generation
```fish
# Generate code for current directory
gofn -src . -out .

# Generate for specific source and output directories
gofn -src=./example -out=./example

# Development run (no install required)
go run ./cmd/gofn -src=. -out=.
```

## Building and Testing
```fish
# Build all packages
go build ./...

# Run example
cd example && go run .

# Verify code (no tests found in project)
go vet ./...

# Format code
gofmt -l .  # List unformatted files
gofmt -w .  # Format all files
```

## Development Workflow
```fish
# 1. Add //gofn: directives to your code
# 2. Generate helper code
go run ./cmd/gofn -src=. -out=.

# 3. Build and verify
go build ./...
go vet ./...

# 4. Test functionality (if example exists)
cd example && go run .
```

## System Commands (Linux/Fish Shell)
```fish
# File operations
ls -la          # List files with details
find . -name "*.go"  # Find Go files
grep -r "gofn:" .    # Search for directives

# Git operations
git status
git add .
git commit -m "message"
git push

# Directory navigation
cd path/to/dir
pwd            # Print working directory
```

## Go-specific Commands
```fish
# Module management
go mod tidy    # Clean up dependencies
go mod download # Download dependencies

# Code quality
go vet ./...   # Static analysis
gofmt -s -w .  # Format with simplifications

# Documentation
go doc package/path
godoc -http=:6060  # Local documentation server
```