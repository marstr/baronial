perl .\get-version.pl > version.txt
perl .\get-revision.pl > revision.txt

set /p version=<version.txt
set /p revision=<revision.txt

if "%1" == "install" (
    go install -ldflags "-X github.com/marstr/baronial/cmd.revision=%revision% -X github.com/marstr/baronial/cmd.version=%version%"
    exit
)

if "%1" == "" (
    go build -ldflags "-X github.com/marstr/baronial/cmd.revision=%revision% -X github.com/marstr/baronial/cmd.version=%version%" -o bin\windows\baronial.exe
)

if "%1" == "darwin" (
    setlocal
    set GOOS=darwin
    go build -ldflags "-X github.com/marstr/baronial/cmd.revision=%revision% -X github.com/marstr/baronial/cmd.version=%version%" -o bin\darwin\baronial
    endlocal
)

if "%1" == "linux" (
    setlocal
    set GOOS=linux
    go build -ldflags "-X github.com/marstr/baronial/cmd.revision=%revision% -X github.com/marstr/baronial/cmd.version=%version%" -o bin\linux\baronial
    endlocal
)

