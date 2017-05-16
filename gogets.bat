if "%GOPATH%"=="" (
echo You Should Set GoPath var
exit
)
go get github.com/astaxie/beego/session
go get github.com/axgle/mahonia
go get golang.org/x/net/websocket