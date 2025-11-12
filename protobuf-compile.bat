@ECHO OFF
SET MISSING_FILENAME=protoc.exe
WHERE protoc.exe >NUL 2>&1
IF ERRORLEVEL 1 GOTO missing

SET MISSING_FILENAME=protoc-gen-go.exe
WHERE protoc-gen-go.exe >NUL 2>&1
IF ERRORLEVEL 1 SET PATH=%USERPROFILE%\go\bin;%PATH%

WHERE protoc-gen-go.exe >NUL 2>&1
IF ERRORLEVEL 1 (
    go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
)
IF ERRORLEVEL 1 GOTO error

WHERE protoc-gen-go.exe >NUL 2>&1
IF ERRORLEVEL 1 GOTO error

GOTO compile

:error
ECHO ERROR: Please try to install protoc-gen-go and make sure it is in your PATH.
ECHO        C:\^>go install google.golang.org/protobuf/cmd/protoc-gen-go@latest

:missing
ECHO %MISSING_FILENAME% cannot be found
GOTO eof

:compile
protoc.exe --go_out=. drm\widevine\cdm\widevine.proto
ECHO SUCCESS: Compiling is done.

:eof
PAUSE