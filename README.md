# agentExample

## Linting

- **From this project:** run `./go-lint` to lint the whole project.
- **As `go lint`:** install once so the `go` command finds it, then run `go lint` from any Go module directory:

  ```bash
  cp go-lint $(go env GOPATH)/bin/
  ```

  Ensure `$(go env GOPATH)/bin` is in your `PATH`. After that, from this (or any) project directory:

  ```bash
  go lint
  ```

  The linter runs in the background (via the script) and lints the whole project.
