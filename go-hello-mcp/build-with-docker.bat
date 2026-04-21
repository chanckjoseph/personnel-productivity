@echo off
setlocal

if not exist bin mkdir bin

echo Building Go MCP server in disposable container...
docker run --rm -v "%cd%":/src -w /src golang:1.22 sh -c "CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o bin/hello-mcp.exe ."
if errorlevel 1 (
  echo Build failed.
  exit /b %errorlevel%
)

echo Build complete: bin/hello-mcp.exe
echo Container removed automatically via --rm.
