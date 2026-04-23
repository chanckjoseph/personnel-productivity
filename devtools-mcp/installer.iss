; Devtools MCP Installer Script
; This script creates a professional Windows installer for devtools-mcp
; Requires InnoSetup 6.0+ (free download: https://jrsoftware.org/isdl.php)

[Setup]
AppName=Devtools MCP
AppVersion=1.0.0
AppPublisher=DevTools
AppPublisherURL=https://github.com/chanckjoseph/personnel-productivity
DefaultDirName={userdocs}\devtools-mcp
DefaultGroupName=Devtools MCP
OutputDir=.\
OutputBaseFilename=install
UninstallDisplayName=Devtools MCP
UninstallDisplayIcon={app}\devtools-mcp.exe
Compression=lzma2
SolidCompression=yes
ArchitecturesInstallIn64BitMode=x64

[Files]
; Copy the compiled binary
Source: "bin\devtools-mcp.exe"; DestDir: "{app}"; Flags: ignoreversion

[Tasks]
Name: "addToVSCode"; Description: "Add to VS Code MCP Configuration"; GroupDescription: "Integration:"; Flags: checkablealone; Check: IsVSCodeInstalled

[Run]
Filename: "{app}\devtools-mcp.exe"; Description: "Run devtools-mcp"; Flags: nowait postinstall unchecked

[Code]
function IsVSCodeInstalled: Boolean;
begin
  Result := FileExists('C:\Program Files\Microsoft VS Code\Code.exe') or
            FileExists('C:\Program Files (x86)\Microsoft VS Code\Code.exe') or
            FileExists(ExpandConstant('{localappdata}\Programs\Microsoft VS Code\Code.exe'));
end;

procedure CurStepChanged(CurStep: TSetupStep);
var
  MCPConfigPath: string;
  MCPConfig: string;
begin
  if CurStep = ssPostInstall then
  begin
    // Create a helper batch file for easy reference
    SaveStringToFile(ExpandConstant('{app}\DEPLOYMENT_GUIDE.txt'),
      'DEVTOOLS MCP INSTALLATION COMPLETE' + #13#10 +
      '===================================' + #13#10 + #13#10 +
      'Binary Location: ' + ExpandConstant('{app}\devtools-mcp.exe') + #13#10 + #13#10 +
      'SETUP IN YOUR PROJECT:' + #13#10 +
      '1. Navigate to your project directory' + #13#10 +
      '2. Create or edit .vscode/mcp.json:' + #13#10 + #13#10 +
      '{' + #13#10 +
      '  "mcpServers": {' + #13#10 +
      '    "devtools-mcp": {' + #13#10 +
      '      "type": "stdio",' + #13#10 +
      '      "command": "' + ExpandConstant('{app}\devtools-mcp.exe') + '"' + #13#10 +
      '    }' + #13#10 +
      '  }' + #13#10 +
      '}' + #13#10 + #13#10 +
      '3. Create credential files in your project root:' + #13#10 +
      '   - .pat (GitHub Personal Access Token)' + #13#10 +
      '   - .username (GitHub username)' + #13#10 + #13#10 +
      '4. Restart VS Code' + #13#10 + #13#10 +
      'For more info, see README.md in the installation directory.',
      false);
  end;
end;

procedure CurUninstallStepChanged(CurUninstallStep: TUninstallStep);
begin
  // Uninstaller messages
  if CurUninstallStep = usUninstall then
    MsgBox('Devtools MCP will be removed. Make sure to remove the MCP config from your .vscode/mcp.json files.', mbInformation, MB_OK);
end;
