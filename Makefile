# CURSE Development Makefile

## Commands

### Build
Build the dashboard binary.
```bash
make build
```

### Test
Run all tests.
```bash
make test
```

### Lint
Run linting (requires `golangci-lint`).
```bash
make lint
```

### Clean
Remove build artifacts and temporary files.
```bash
make clean
```

### Help
Show this help message.
```bash
make help
```

## Targets

### build
Builds the `curse` binary in the `bin/` directory.

### test
Runs `go test ./...`.

### lint
Runs `golangci-lint run`.

### clean
Removes the `bin/` directory and build artifacts.

### help
Prints the help message.
