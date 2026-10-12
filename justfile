# readable-gcl command runner

# Default recipe: list all available recipes
default:
    @just --list

# Format Go source code using golangci-lint and format justfile
fmt:
    @if [ -f "./custom-gcl" ]; then \
        ./custom-gcl fmt; \
    else \
        golangci-lint fmt; \
    fi
    just --fmt

# Check that code and justfile are properly formatted without modifying files
fmt-check:
    @if [ -f "./custom-gcl" ]; then \
        DIFF=$(./custom-gcl fmt --diff); \
        if [ -n "$DIFF" ]; then \
            echo "error: code is not formatted. run 'just fmt' and commit changes before releasing."; \
            echo "$DIFF"; \
            exit 1; \
        fi; \
    elif command -v golangci-lint >/dev/null 2>&1; then \
        DIFF=$(golangci-lint fmt --diff); \
        if [ -n "$DIFF" ]; then \
            echo "error: code is not formatted. run 'just fmt' and commit changes before releasing."; \
            echo "$DIFF"; \
            exit 1; \
        fi; \
    fi
    @just --fmt --check

# Run linters against the codebase
lint:
    @if [ -f "./custom-gcl" ]; then \
        ./custom-gcl run ./...; \
    else \
        golangci-lint run ./...; \
    fi

# Run all automated tests with race detection and statement coverage
test:
    go test -race -coverprofile=coverage.txt -cover ./...

# Block if git working directory is not in a clean state
check-git-clean:
    @if [ -n "$(git status --porcelain 2>/dev/null)" ]; then \
        echo "error: git working directory is dirty. commit or stash changes before releasing."; \
        exit 1; \
    fi

# Build standalone CLI binary
build:
    go build -o bin/readable-gcl ./cmd/readable-gcl

# Dry-run release: preview version bump and changelog without making changes
release-dry-run bump="patch": fmt-check lint test check-git-clean
    ./scripts/release.sh {{ bump }} --dry-run

# Create release: ensure no fmt issues, no lint issues, tests pass, and git is clean
release bump="patch": fmt-check lint test check-git-clean
    ./scripts/release.sh {{ bump }}
