@rem	Get the contents of the VERSION file and increment the most minor digit
for /f "tokens=1,2,3 delims=." %%a in (VERSION) do (
	set var1=%%a
	set var2=%%b
	set /a var3=%%c+1
)

@rem	And put it back in the file, this outputs to file with no newline
@rem	(echo puts an unavoidable newline so we use set instead)
<NUL set /p ="%var1%.%var2%.%var3%" > VERSION

@rem	Now build the program with the new version
for /f "tokens=* delims=" %%a in (VERSION) do (
  go build -ldflags="-w -s -X main.version=%%a -H windowsgui" -o "GUIEtcdClient.exe" ./src
)

@rem and move them to the correct dir
move .\bin\GUIEtcdClient.exe .\bin\GUIEtcdClient.exe.bak
move .\GUIEtcdClient.exe .\bin\GUIEtcdClient.exe