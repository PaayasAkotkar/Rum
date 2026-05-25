@echo off
echo [Updating] just a sec or a nano....
go mod tidy
if %ERRORLEVEL% neq 0 (
    echo [Error] go mod tidy failed!
    exit /b %ERRORLEVEL%
)

echo [Building] app.exe soon or in a nano...
go build
if %ERRORLEVEL% neq 0 (
    echo [Error] Build failed!
    exit /b %ERRORLEVEL%
)

echo [Launching] app.exe running ....
app.exe
