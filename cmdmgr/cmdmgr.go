package cmdmgr

import (
	"bufio"
	"fmt"
	"io"
	"os/exec"
	"time"

	"github.com/axgle/mahonia"
)

// CmdSessionManager ...
type CmdSessionManager struct {
	sessions map[string]*CmdSession
}

// CmdSession ...
type CmdSession struct {
	Running     bool
	SessionID   string
	Cmd         *exec.Cmd
	Outcmd      chan string
	Incmd       chan string
	Stopped     bool
	StartTime   time.Time
	LastCmdTime time.Time
	EndTime     time.Time
	outpipe     io.ReadCloser
	inpipe      io.WriteCloser
}

// NewCmdManager ...
func NewCmdManager() *CmdSessionManager {
	return &CmdSessionManager{make(map[string]*CmdSession)}
}

// CheckSessionExist ...
type CheckSessionExist func(session string) bool

// GetCmdSession ...
func (cm *CmdSessionManager) GetCmdSession(sessionID string) *CmdSession {
	if cs, ok := cm.sessions[sessionID]; ok {
		return cs
	}
	cs := &CmdSession{
		false,
		sessionID,
		nil,
		make(chan string, 1024),
		make(chan string, 1024),
		false,
		time.Now(),
		time.Now(),
		time.Now(),
		nil,
		nil,
	}
	cm.sessions[sessionID] = cs
	return cs
}

// DeleteCmdSession ...
func (cm *CmdSessionManager) DeleteCmdSession(sessionID string) {
	cs, ok := cm.sessions[sessionID]
	if !ok {
		return
	}
	cs.End(true)
	delete(cm.sessions, sessionID)
}

// CheckSessions ...
func (cm *CmdSessionManager) CheckSessions(funcCheck CheckSessionExist) {
	for _, v := range cm.sessions {
		if funcCheck(v.SessionID) != true {
			cm.DeleteCmdSession(v.SessionID)
		}
	}
}

// Start ...
func (cs *CmdSession) Start() bool {
	if cs.Running {
		return false
	}
	//ctx, cancel := context.WithCancel(context.Background())
	cs.Cmd = exec.Command("cmd", "") //  " /k echo 1&c:\\sleep 2000&echo 2&c:\\sleep 2000&echo 3&c:\\sleep 2000&echo 4&c:\\sleep 2000&echo 5"
	cs.outpipe, _ = cs.Cmd.StdoutPipe()
	cs.inpipe, _ = cs.Cmd.StdinPipe()

	if cs.Cmd.Start() == nil {

		cs.Running = true
		go cs.run(1)
		go cs.run(2)
		fmt.Println("Cmd Session Start..", cs.SessionID)
		return true
	}
	return false

}

// End ...
func (cs *CmdSession) End(force bool) {
	if !cs.Running {
		return
	}
	cs.Running = false
	close(cs.Outcmd)
	close(cs.Incmd)
	if cs.Cmd != nil && cs.Cmd.Process != nil {
		//if !cs.cmd.ProcessState.Exited() {
		if force {
			cs.Cmd.Process.Kill()
		} else {
			cs.Cmd.Wait()
		}
		//}
	}
	cs.Cmd = nil
	cs.Stopped = true
}

func (cs *CmdSession) run(ioro int) {
	if ioro == 1 {
		reader := bufio.NewReader(cs.outpipe)
		pid := cs.Cmd.Process.Pid
		fmt.Println("pid :", pid)
		for {
			line, err2 := reader.ReadString('\n')
			if err2 != nil || io.EOF == err2 || cs.Cmd == nil || !cs.Running {
				break
			}

			line = ConvertToString(line, "gbk", "utf-8")
			fmt.Println(" * ", line)
			select {
			case cs.Outcmd <- line:
			default:
			}
		}
		fmt.Println("fini", ioro)
		cs.End(false)
	} else if ioro == 2 {
		//writer := bufio.NewWriter(cs.inpipe)
		for {
			cmdstr := <-cs.Incmd
			if cs.Cmd == nil || !cs.Running {
				break
			}
			fmt.Println(" # ", cmdstr)
			fmt.Fprintln(cs.inpipe, cmdstr)
			//writer.WriteString(cmdstr)
			//writer.Flush()
		}
		fmt.Println("fini", ioro)
	}
}

// ConvertToString 编码转换
//  eg. ConvertToString(line, "gbk", "utf-8")
func ConvertToString(src string, srcCode string, tagCode string) string {
	srcCoder := mahonia.NewDecoder(srcCode)
	srcResult := srcCoder.ConvertString(src)
	tagCoder := mahonia.NewDecoder(tagCode)
	_, cdata, _ := tagCoder.Translate([]byte(srcResult), true)
	result := string(cdata)
	return result
}
