package main

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	//"os/signal" 	// 哪个沙雕规定不用的库报错的
	"os"
	//"bytes"
	"log"
	"strings"
	//"bufio"
	//"github.com/axgle/mahonia"
	"path/filepath"
	"time"

	"html/template"

	"./cmdmgr"
	"github.com/astaxie/beego/session"
)

const (
	listenIP   = "127.0.0.1"
	listenPort = "8011"
)

func getCurrentDirectory() string {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal(err)
	}
	return strings.Replace(dir, "\\", "/", -1)
}

var fs http.Handler  // 文件Handler
var ms http.ServeMux // 多路由Handler

type as struct{}

func (*as) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.String() //获得访问的路径
	switch true {
	case strings.HasPrefix(strings.ToLower(path), "/script"):
		fallthrough
	case strings.HasPrefix(strings.ToLower(path), "/image"):
		fallthrough
	case strings.HasPrefix(strings.ToLower(path), "/static"):
		fs.ServeHTTP(w, r) // 静态路径下的传递给FileServer Handler
	default:
		ms.ServeHTTP(w, r) // 其他路径传递给 ServerMux
	}
}

var ch chan int
var globalSessions *session.Manager
var cmdSessions *cmdmgr.CmdSessionManager

func main() {
	sessionConfig := &session.ManagerConfig{
		CookieName:      "gosessionid",
		EnableSetCookie: true,
		Gclifetime:      3600,
		Maxlifetime:     3600,
		Secure:          false,
		CookieLifeTime:  3600,
		ProviderConfig:  listenIP + ":" + listenPort,
	}
	globalSessions, _ = session.NewManager("memory", sessionConfig)
	cmdSessions = cmdmgr.NewCmdManager()
	sss := cmdSessions.GetCmdSession("123")
	sss.Start()

	fmt.Println("dir:", getCurrentDirectory())

	fs = http.FileServer(http.Dir("./"))
	ms.HandleFunc("/", dealHTTPFunc)
	ms.HandleFunc("/OutCmd", dealOutCmd)
	ms.HandleFunc("/DoCmd", dealDoCmd)

	fmt.Println("Start listen..")

	cmd := exec.Command("cmd", " /c start "+"http://"+listenIP+":"+listenPort+"/")
	cmd.Run()
	go globalSessions.GC()

	go http.ListenAndServe(listenIP+":"+listenPort, &as{})
	<-ch
}

// CMD部分
/*
type cmdSessionManager struct {
	sessions map[string]*cmdSession
}
type cmdSession struct {
	running     bool
	sessionID   string
	cmd         *exec.Cmd
	outcmd      chan string
	incmd       chan string
	end         bool
	state       int
	starttime   time.Time
	lastcmdtime time.Time
	endtime     time.Time
}
type checkSessionExist func(session string) bool

func (cm *cmdSessionManager) GetCmdSession(sessionID string) *cmdSession {
	if cs, ok := cm.sessions[sessionID]; ok {
		return cs
	}
	cs := &cmdSession{
		false,
		sessionID,
		nil,
		make(chan string, 1024),
		make(chan string, 1024),
		false,
		0,
		time.Now(),
		time.Now(),
		time.Now(),
	}
	cm.sessions[sessionID] = cs
	return cs
}
func (cm *cmdSessionManager) DeleteCmdSession(sessionID string) {
	cs, ok := cm.sessions[sessionID]
	if !ok {
		return
	}
	cs.End(true)
	delete(cm.sessions, sessionID)
}
func (cm *cmdSessionManager) CheckSessions(funcCheck checkSessionExist) {
	for k, v := range cm.sessions {
		if funcCheck(v.sessionID) != true {
			cm.DeleteCmdSession(v.sessionID)
		}
	}
}

func (cs *cmdSession) Start() {
	if cs.running {
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
}
func (cs *cmdSession) End(force bool) {
	if !cs.running {
		return
	}
	if cs.cmd != nil && cs.cmd.Process != nil {
		if !cs.cmd.ProcessState.Exited() {
			if force {
				cs.cmd.Process.Kill()
			} else {
				cs.cmd.Wait()
			}
		}
	}
	cs.cmd = nil
	cs.running = false
	close(cs.outcmd)
	close(cs.incmd)
}
func (cs *cmdSession) Run() {

}
*/
/// HTTP部分

type infoStru struct {
	Title   string
	Cmd     string
	TimeTag time.Time
	SKey    string
}

func dealHTTPFunc(w http.ResponseWriter, req *http.Request) {
	req.ParseForm()
	sess, _ := globalSessions.SessionStart(w, req)
	defer sess.SessionRelease(w)
	fmt.Println("get:", req.RequestURI)

	cmd := req.Form.Get("cmd")

	strKey := "whatareyou123456"
	byteAes, _ := aesEncrypt([]byte(sess.SessionID()), []byte(strKey))
	strAes, _ := toHex(byteAes)

	info := infoStru{"WebCmd  v0.1 by wwcMonkey", cmd, time.Now(), strAes}
	tmpl, _ := template.ParseFiles("index.html")
	tmpl.Execute(w, info)
	/*
		fmt.Fprint(w, "<html><head><title>WebCmd  v0.1 by wwcMonkey</title></head><body>", time.Now())
		fmt.Fprint(w, "<script src=\"/Script/webc.js\"></script>")
		fmt.Fprint(w, "\n<br>cmd:", cmd)
		if cmd != "" {

		}

		fmt.Fprint(w, "\n<br>Output:<textarea style=\"width:500px;height:500px\"></textarea>")

		fmt.Fprint(w, "<form method=\"post\" target=\"iframe1\" action=\"DoCmd\" id=\"form1\"><textarea name=\"cmd\" style=\"width:500px;height:20px\" id=\"cmd\"></textarea><iframe name=\"iframe1\" style=\"display:none\" id=\"iframe1\"></iframe>   <input type=\"submit\" value=\"DoCmd\"/></form>")

		fmt.Fprint(w, "<script>", "var sKey=\""+strAes+"\";init(sKey);", "</script>")
		fmt.Fprint(w, "</body></html>")*/
}

type outCmdStru struct {
	outcmd  string
	timetag time.Time
}

func check(sess *session.Store, req *http.Request) bool {
	if (*sess).Get("CmdLoaded") != "true" {
		return false
	}

	strKey := "whatareyou123456"
	byteAes, _ := aesEncrypt([]byte((*sess).SessionID()), []byte(strKey))
	strAes, _ := toHex(byteAes)
	if req.Form.Get("sKey") != strAes {
		return false
	}
	return true
}
func dealDoCmd(w http.ResponseWriter, req *http.Request) {
	// 参数校验
	req.ParseForm()
	if req.Form.Get("sKey") == "" {
		return
	}
	sess, _ := globalSessions.SessionStart(w, req)
	defer sess.SessionRelease(w)

	if check(&sess, req) == false {
		return
	}

	// 执行参数
}
func dealOutCmd(w http.ResponseWriter, req *http.Request) {
	// 参数校验
	req.ParseForm()
	if req.Form.Get("sKey") == "" {
		return
	}
	sess, _ := globalSessions.SessionStart(w, req)
	defer sess.SessionRelease(w)

	if check(&sess, req) == false {
		return
	}

	// 获取输出
	var buffer bytes.Buffer
	for {
		select {
		//case block := <-outcmd:
		//	buffer.WriteString(block)
		default:
			goto finished
		}
	}
finished:
	outstr := buffer.String()

	t := time.Now()
	outcmd := &outCmdStru{outstr, t}
	bytes, _ := json.Marshal(outcmd)
	w.Write(bytes)
}

//// 加解密相关

func toHex(data []byte) (string, bool) {
	return hex.EncodeToString(data), true
}
func fromHex(str string, expectlen int) ([]byte, bool) {
	bytes := []byte{}
	str = strings.TrimSpace(str)
	if len(str) != expectlen && expectlen != -1 {
		return nil, false
	}
	for i := 0; i < len(str); i++ {
		abl := str[i : i+2]
		bts, err := hex.DecodeString(abl)
		if err != nil {
			return nil, false
		}
		bytes = append(bytes, bts[0])
	}

	return bytes, true
}

func aesEncrypt(origData, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	blockSize := block.BlockSize()
	origData = pKCS5Padding(origData, blockSize)
	// origData = ZeroPadding(origData, block.BlockSize())
	blockMode := cipher.NewCBCEncrypter(block, key[:blockSize])
	crypted := make([]byte, len(origData))
	// 根据CryptBlocks方法的说明，如下方式初始化crypted也可以
	// crypted := origData
	blockMode.CryptBlocks(crypted, origData)
	return crypted, nil
}

func aesDecrypt(crypted, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	blockSize := block.BlockSize()
	blockMode := cipher.NewCBCDecrypter(block, key[:blockSize])
	origData := make([]byte, len(crypted))
	// origData := crypted
	blockMode.CryptBlocks(origData, crypted)
	origData = pKCS5UnPadding(origData)
	// origData = ZeroUnPadding(origData)
	return origData, nil
}

func pKCS5Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

func pKCS5UnPadding(origData []byte) []byte {
	length := len(origData)
	// 去掉最后一个字节 unpadding 次
	unpadding := int(origData[length-1])
	return origData[:(length - unpadding)]
}
