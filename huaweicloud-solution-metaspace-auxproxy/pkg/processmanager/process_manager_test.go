package processmanager

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestProcessManagerServerSessionRecord(t *testing.T) {
	InitProcessManager()
	// 注入假数据
	pid := 3301
	serverSessionId := "server-session-6890fe60-3b0a-4dca-92d2-77a80619d6ef"
	ProcessMgr.Processes = append(ProcessMgr.Processes, NewProcess("/local/app/fake-server.sh",
		"", pid))
	process := ProcessMgr.GetProcess(pid)

	Convey("record test", t, func() {
		// 第一次不命中
		isStarted := ProcessMgr.IsServerSessionStarted(process, serverSessionId)
		So(isStarted, ShouldBeFalse)

		// 加入记录
		ProcessMgr.RecordServerSessionStarted(process, serverSessionId)

		// 第二次命中
		isStarted = ProcessMgr.IsServerSessionStarted(process, serverSessionId)
		So(isStarted, ShouldBeTrue)

	})
}
