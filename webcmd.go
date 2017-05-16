package main

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
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
	"golang.org/x/net/websocket"
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
var ws http.Handler  // WebSocket Handler

type as struct{} // 自定义Handler

func (*as) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.String() //获得访问的路径
	switch true {
	case strings.HasPrefix(strings.ToLower(path), "/script"):
		fallthrough
	case strings.HasPrefix(strings.ToLower(path), "/image"):
		fallthrough
	case strings.HasPrefix(strings.ToLower(path), "/static"):
		fs.ServeHTTP(w, r) // 静态路径下的传递给FileServer Handler
	case strings.HasPrefix(strings.ToLower(path), "/ws"):
		ws.ServeHTTP(w, r)
	default:
		ms.ServeHTTP(w, r) // 其他路径传递给 ServerMux
	}
}

var ch chan int

var globalSessions *session.Manager
var cmdSessions *cmdmgr.CmdSessionManager
var sss *cmdmgr.CmdSession

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
	sss = cmdSessions.GetCmdSession("")
	sss.Start()

	fmt.Println("dir:", getCurrentDirectory())

	fs = http.FileServer(http.Dir("./"))
	ms.HandleFunc("/", dealHTTPFunc)
	ms.HandleFunc("/OutCmd", dealOutCmd)
	ms.HandleFunc("/DoCmd", dealDoCmd)
	ws = websocket.Handler(dealWebSocket)

	fmt.Println("Start listen..")

	cmd := exec.Command("cmd", " /c start "+"http://"+listenIP+":"+listenPort+"/")
	cmd.Run()

	//cmd1.Process.Kill()
	go globalSessions.GC()

	go http.ListenAndServe(listenIP+":"+listenPort, &as{})
	<-ch
}

/// HTTP部分

type infoStru struct {
	Title   string
	TimeTag time.Time
	SKey    string
}

// 首页

func dealHTTPFunc(w http.ResponseWriter, req *http.Request) {
	req.ParseForm()
	sess, _ := globalSessions.SessionStart(w, req)
	defer sess.SessionRelease(w)
	fmt.Println("get:", req.RequestURI)
	sess.Set("CmdLoaded", "true")

	strKey := "whatareyou123456"
	byteAes, _ := aesEncrypt([]byte(sess.SessionID()), []byte(strKey))
	strAes, _ := toHex(byteAes)

	info := infoStru{"WebCmd  v0.1 by wwcMonkey", time.Now(), strAes}
	tmpl, _ := template.ParseFiles("index.html")
	tmpl.Execute(w, info)
}

type myResp struct {
	header map[string][]string
}

func (resp myResp) Header() http.Header {
	return resp.header
}
func (resp myResp) Write([]byte) (int, error) {
	return 0, nil
}
func (resp myResp) WriteHeader(int) {

}

func dealWebSocket(ws *websocket.Conn) {
	defer ws.Close()
	fmt.Println("ws: build")
	var err error
	var str string
	w := myResp{make(map[string][]string)}
	sess, _ := globalSessions.SessionStart(w, ws.Request())
	defer sess.SessionRelease(w)

	fmt.Print("ws: wait for input..")
	for {
		if err = websocket.Message.Receive(ws, &str); err != nil {
			fmt.Println("recv fail..")
			break
		} else {
			if checkstr(&sess, str) == false {
				fmt.Println("check fail..", str)
				return
			}
			fmt.Println("check success..")
		}

		for {
			fmt.Println("ws: wait for Outcmd..")
			stro := <-sss.Outcmd
			fmt.Print("ws: # ", stro)
			if err = websocket.Message.Send(ws, stro); err != nil {
				fmt.Println("send fail..")
				break
			} else {
				//fmt.Println("向客户端发送：", str)
			}
		}
	}
}

func check(sess *session.Store, req *http.Request) bool {
	return checkstr(sess, req.Form.Get("sKey"))
}
func checkstr(sess *session.Store, sKeyReq string) bool {
	if (*sess).Get("CmdLoaded") != "true" || sKeyReq == "" {
		return false
	}

	strKey := "whatareyou123456"
	byteAes, _ := aesEncrypt([]byte((*sess).SessionID()), []byte(strKey))
	strAes, _ := toHex(byteAes)

	h := md5.New()
	h.Write([]byte(strAes))
	strAes = hex.EncodeToString(h.Sum(nil))

	if sKeyReq != strAes {
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
	fmt.Println("get:", req.RequestURI)
	sess, _ := globalSessions.SessionStart(w, req)
	defer sess.SessionRelease(w)

	if check(&sess, req) == false {
		return
	}
	cmd := req.Form.Get("cmd")
	if cmd != "" {
		sss.Incmd <- cmd
	}

	// 执行参数
}

type outCmdStru struct {
	outcmd  string
	timetag time.Time
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
		case block := <-sss.Outcmd:
			buffer.WriteString(block)
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
