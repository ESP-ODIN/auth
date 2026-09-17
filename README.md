# auth

### Prerequisites
- **Go** (version 1.22 or later required for standard `net/http` method-based routing):
  - **Linux (Ubuntu/Debian):**
    ```bash
    sudo apt update && sudo apt install -y golang-go
    ```
  - **Windows (PowerShell / Winget):**
    ```powershell
    winget install GoLang.Go
    ```
  - **macOS (Homebrew):**
    ```bash
    brew install go
    ```
- **VS Code** with the official **Go** extension (by *Go Team at Google*)
- **Make** (optional, for running Makefile targets):
  - **Windows (PowerShell / Winget):**
    ```powershell
    winget install GnuWin32.Make
    ```
    *(Alternatively via Chocolatey: `choco install make`)*
  - **macOS:** Check with `make --version`. If missing, run:
    ```bash
    xcode-select --install
    ```
  - **Linux (Ubuntu/Debian):**
    ```bash
    sudo apt install -y make
    ```

---

### Environment Setup

1. At the root of the project, duplicate the example file to create your local `.env` file[cite: 1]:
   ```bash
   cp .env.exemple .env

  
## Docker

A development `Dockerfile` runs the API with hot-reload (via [air](https://github.com/air-verse/air)),
meant to be plugged as-is into the project's docker-compose. Fill in `DATABASE_URL` in `.env` first
(see `.env.exemple`).

Build the image:

```sh
make docker-build
```

Run it locally with hot-reload (mounts the source code and exposes the app on port 8090):

```sh
make docker-dev
```

Then check it with `curl http://127.0.0.1:8090/health`. Editing any `.go` file rebuilds and restarts
the server automatically inside the container. The container itself listens on `0.0.0.0:8080`
(overridden from `.env`'s `HTTP_ADDR` by the `docker-dev` target); only the host-side port is `8090`.
