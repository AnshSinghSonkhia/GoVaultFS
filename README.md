# Go VaultFS

## Development

### Windows Setup

This project uses a Makefile for build automation. On Windows, you'll need to install GNU Make:

```powershell
winget install GnuWin32.Make
```

After installation, run the setup script to add make to your PATH:

```powershell
.\setup-make.ps1
```

### Available Commands

- `make build` - Build the application
- `make run` - Build and run the application
- `make test` - Run tests

### Alternative (PowerShell Scripts)

If you prefer not to use make, PowerShell scripts are also available:

- `.\build.ps1` - Build the application
- `.\run.ps1` - Build and run the application
- `.\test.ps1` - Run tests