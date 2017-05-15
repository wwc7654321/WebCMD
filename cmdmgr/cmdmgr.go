package cmdmgr

import (
	"fmt"
	"os/exec"
	"time"
)

// CmdSessionManager ...
type CmdSessionManager struct {
	sessions map[string]*CmdSession
}

// CmdSession ...
type CmdSession struct {
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
		0,
		time.Now(),
		time.Now(),
		time.Now(),
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
		if funcCheck(v.sessionID) != true {
			cm.DeleteCmdSession(v.sessionID)
		}
	}
}

// Start ...
func (cs *CmdSession) Start() {
	if cs.running {
		return
	}
	//ctx, cancel := context.WithCancel(context.Background())
	fmt.Println("Cmd Session Start..", cs.sessionID)
}

// End ...
func (cs *CmdSession) End(force bool) {
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

// Run ...
func (cs *CmdSession) Run() {

}
