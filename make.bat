perl .\get-version.pl > version.txt
perl .\get-revision.pl > revision.txt

set /p version=<version.txt
set /p revision=<revision.txt

go test ./... || exit /b

go build -ldflags "-X github.com/marstr/baronial/cmd.revision=%revision% -X github.com/marstr/baronial/cmd.version=%version%" -o bin\windows\baronial.exe

setlocal
set GOOS=darwin
go build -ldflags "-X github.com/marstr/baronial/cmd.revision=%revision% -X github.com/marstr/baronial/cmd.version=%version%" -o bin\darwin\baronial
endlocal

setlocal
set GOOS=linux
go build -ldflags "-X github.com/marstr/baronial/cmd.revision=%revision% -X github.com/marstr/baronial/cmd.version=%version%" -o bin\linux\baronial
endlocal

go install -ldflags "-X github.com/marstr/baronial/cmd.revision=%revision% -X github.com/marstr/baronial/cmd.version=%version%"
