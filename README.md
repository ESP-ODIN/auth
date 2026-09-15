# auth

### Prerequisites
- **Go** (version 1.20 or later):
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
