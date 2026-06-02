# Contributing to CURSE

First off, thank you for contributing. CURSE is built to evolve quickly, and high-signal contributions make that possible.

## How to Contribute

### 1. Report Bugs or Suggest Features

Open an issue at:
[https://github.com/M523zappin/Curse-Core/issues](https://github.com/M523zappin/Curse-Core/issues)

When possible, include:
- exact reproduction steps
- expected vs actual behavior
- platform details and version

### 2. Submit Pull Requests

- Fork the repository and branch from `master`.
- Keep changes focused and reviewable.
- Add or update tests for behavioral changes.
- Update documentation when behavior changes.
- Ensure CI passes before requesting review.

Recommended pre-PR checks:

```bash
go fmt ./...
go vet ./...
go test ./...
golangci-lint run ./...
```

If `golangci-lint` is not installed locally, submit anyway and use CI feedback to iterate.

## Branch and Commit Strategy

- Use focused branches (`feat/...`, `fix/...`, `chore/...`).
- Keep commits atomic and descriptive.
- Prefer conventional prefixes (`feat:`, `fix:`, `chore:`, `docs:`) for cleaner release notes.

## Coding Standards

- **Language:** Go (core runtime and TUI).
- **Formatting:** Standard Go formatting (`go fmt` / `gofmt`).
- **Reliability:** Prioritize deterministic behavior and recoverable failure modes.
- **Documentation:** Any non-obvious behavior needs docs.

## Code of Conduct

Please keep interactions respectful and professional. See `CODE_OF_CONDUCT.md`.

## License

By contributing, you agree your contributions are licensed under the MIT License.
