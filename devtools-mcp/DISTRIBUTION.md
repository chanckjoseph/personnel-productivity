# Devtools MCP - Distribution & Installation Guide

## Quick Start for End Users

### Windows

If you received `install.exe`:

1. **Run the installer**
   ```
   install.exe
   ```
   This installs devtools-mcp to your Documents folder.

2. **In your project, create `.vscode/mcp.json`:**
   ```json
   {
     "mcpServers": {
       "devtools-mcp": {
         "type": "stdio",
         "command": "C:\\Users\\YourUsername\\Documents\\devtools-mcp\\devtools-mcp.exe"
       }
     }
   }
   ```

3. **Create credential files in your project root:**
   - `.pat` - Your GitHub Personal Access Token
   - `.username` - Your GitHub username

4. **Restart VS Code** - the MCP will auto-connect

### Linux/macOS

If you received `install.sh`:

1. **Make the script executable and run it**
   ```bash
   chmod +x install.sh
   ./install.sh
   ```
   This installs devtools-mcp to `~/.local/devtools-mcp`

2. **In your project, create `.vscode/mcp.json`:**
   ```json
   {
     "mcpServers": {
       "devtools-mcp": {
         "type": "stdio",
         "command": "~/.local/devtools-mcp/devtools-mcp"
       }
     }
   }
   ```

3. **Create credential files in your project root:**
   ```bash
   echo "your_github_token" > .pat
   echo "your_username" > .username
   chmod 600 .pat .username
   ```

4. **Restart VS Code** - the MCP will auto-connect

---

## For Developers: Building the Installers

### ⚠️ IMPORTANT: Always Build Fresh Before Distribution

**NEVER distribute a pre-built exe directly.** Always rebuild from source to ensure you're distributing the latest version.

**Before each release, run ONE of these:**

**Windows:**
```batch
distribute.bat
```

**Linux/macOS:**
```bash
chmod +x distribute.sh
./distribute.sh
```

These scripts:
- ✓ Rebuild the binary from the latest source code
- ✓ Verify the build succeeded
- ✓ Stamp the build with current date/time
- ✓ Ensure the `bin/` directory is always fresh

**Then proceed with installer creation:**
- Windows: `build-installer.bat` → creates `install.exe`
- Linux: `build-linux.sh` → creates `install.sh`

---

### Prerequisites
- **Docker** - Required for cross-platform binary compilation
- **InnoSetup 6** (Windows only) - Download from https://jrsoftware.org/isdl.php (optional)

### Build All Binaries (Windows, Linux, macOS)

```bash
./build-all.sh
```

This compiles binaries for:
- Windows 64-bit (.exe)
- Linux 64-bit
- macOS Intel & Apple Silicon

### Build Windows Installer

#### Option A: Automatic (if InnoSetup is installed on Windows)
```batch
build-installer.bat
```

This will:
1. ✓ Compile Go binary (Windows x64)
2. ✓ Build InnoSetup installer
3. ✓ Generate `install.exe`

#### Option B: Manual

**Step 1:** Build the binary
```batch
setup.bat
```

**Step 2:** Compile the installer
- Right-click `installer.iss` → "Compile with InnoSetup"
- OR use: `"C:\Program Files (x86)\Inno Setup 6\ISCC.exe" installer.iss`

### Build Linux (User-Local Installation)

```bash
./build-linux.sh
./install.sh
```

This installs to `~/.local/devtools-mcp`

### Build Linux (.deb Package - Optional)

For Debian/Ubuntu package distribution:

```bash
sudo apt-get install ruby-dev
sudo gem install fpm
./build-linux.sh
```

The script will offer to create a `.deb` package if `fpm` is available.

---

## Distribution Package Contents

### Windows Distribution
```
devtools-mcp/
├── install.exe                 ← Single-file installer (MAIN DELIVERABLE)
├── bin/
│   └── devtools-mcp.exe       ← Binary (included in installer)
└── README.md                  ← User documentation
```

### Linux Distribution
```
devtools-mcp/
├── install.sh                 ← Installation script (MAIN DELIVERABLE)
├── bin/
│   └── devtools-mcp          ← Binary (used by install.sh)
└── README.md                 ← User documentation
```

### macOS Distribution
```
devtools-mcp/
├── install.sh                 ← Installation script (MAIN DELIVERABLE)
├── bin/
│   ├── devtools-mcp-darwin-amd64  ← Intel 64-bit
│   └── devtools-mcp-darwin-arm64  ← Apple Silicon
└── README.md                 ← User documentation
```

### Developer Repository
```
devtools-mcp/
├── main.go                    ← Source code
├── go.mod                     ← Go module definition
├── installer.iss              ← Windows installer script
├── setup.bat                  ← Windows binary builder
├── setup.sh                   ← Unix binary builder
├── distribute.bat             ← ⭐ Use for Windows distribution (rebuilds fresh)
├── distribute.sh              ← ⭐ Use for Linux/macOS distribution (rebuilds fresh)
├── build-installer.bat        ← Windows installer builder (run after distribute.bat)
├── build-linux.sh             ← Linux installer builder
├── build-all.sh               ← Cross-platform builder
├── install.bat                ← Lightweight Windows installer
├── install.sh                 ← Lightweight Unix installer
└── Dockerfile                 ← For build environments
```

---

## What the Installers Do

### Windows (install.exe)
✓ **Installs** devtools-mcp.exe to user's Documents/devtools-mcp  
✓ **Creates** easy-to-reference deployment guides  
✓ **Supports** uninstall via Windows "Add/Remove Programs"  
✓ **Project-agnostic** - works with any project by editing mcp.json  

### Linux/macOS (install.sh)
✓ **Installs** devtools-mcp to ~/.local/devtools-mcp  
✓ **Creates** easy-to-reference deployment guides  
✓ **Allows** system-wide installation via `sudo cp`  
✓ **Project-agnostic** - works with any project by editing mcp.json  

---

## Deploying to Another Project

### Windows: Share install.exe
```
- Simplest for end users
- They download and run install.exe
- Configure each project with mcp.json
```

### Linux/macOS: Share install.sh
```bash
chmod +x install.sh
./install.sh
# Then configure each project with mcp.json
```

### Multi-Platform (GitHub Releases)

```yaml
# .github/workflows/release.yml
name: Build Release

on:
  push:
    tags: ['v*']

jobs:
  build:
    strategy:
      matrix:
        os: [ubuntu-latest, windows-latest]
    runs-on: ${{ matrix.os }}
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.22'
      
      - name: Build Windows
        if: runner.os == 'Windows'
        run: .\build-installer.bat
      
      - name: Build Linux
        if: runner.os == 'Linux'
        run: |
          chmod +x build-all.sh
          ./build-all.sh
      
      - name: Upload Release
        uses: softprops/action-gh-release@v1
        with:
          files: |
            install.exe
            bin/devtools-mcp
            bin/devtools-mcp-darwin-*
```

### Package Manager Distribution

#### Linux: Debian Repository
```bash
./build-linux.sh
# Creates devtools-mcp_1.0.0_amd64.deb
```

#### macOS: Homebrew
```ruby
# devtools-mcp/formula/devtools-mcp.rb
class DevtoolsMcp < Formula
  desc "MCP server for Git workflows"
  homepage "https://github.com/..."
  url "https://github.com/.../releases/download/v1.0.0/devtools-mcp-darwin-arm64"
  sha256 "..."
  
  def install
    bin.install "devtools-mcp-darwin-arm64" => "devtools-mcp"
  end
end
```

---

## Troubleshooting

**Q: InnoSetup not found during build**  
A: Download from https://jrsoftware.org/isdl.php and install, then re-run `build-installer.bat`

**Q: Binary not working in project**  
A: Edit `.vscode/mcp.json` to use the full path to devtools-mcp.exe shown in DEPLOYMENT_GUIDE.txt

**Q: Docker build failing**  
A: Ensure Docker Desktop is running: `docker --version`

---

## Version Management

When releasing updates:

1. **Make code changes** to `main.go`, `types.go`, etc.

2. **Rebuild the binary FROM SOURCE:**
   - Windows: `distribute.bat`
   - Linux: `./distribute.sh`

3. **Build the installer:**
   - Windows: `build-installer.bat` → creates `install.exe`
   - Linux: `build-linux.sh` → creates `install.sh`

4. **Update version in `installer.iss` (Windows only):**
   ```
   AppVersion=1.1.0
   ```

5. **Distribute** the installer to end users

6. **Tag the release** in Git:
   ```bash
   git tag v1.1.0
   git push origin v1.1.0
   ```

This workflow ensures `bin/devtools-mcp.exe` always reflects the latest source code.

---

## Technical Details

- **Language**: Go 1.22
- **Compiler**: InnoSetup 6 (creates standalone .exe)
- **Binary Size**: ~10-12 MB (stripped static binary)
- **Dependencies**: None (fully self-contained)
- **Platforms**: Windows 64-bit
- **Installation**: User account (no admin required)

