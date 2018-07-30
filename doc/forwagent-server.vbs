'Sample script to automatically start the server, listening on a specific
'interface and port. To use, place this file in the Windows startup directory:
'WIN+Run -> "shell:startup" -> OK.
'Make sure to adjust the path and interface/port below as desired.
Dim WinScriptHost
Set WinScriptHost = CreateObject("WScript.Shell")
WinScriptHost.Run Chr(34) & "C:\Applications\forwagent-server.exe" & Chr(34) & "192.168.137.1:4711", 0
Set WindScriptHost = Nothing
