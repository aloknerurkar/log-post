package server

import (
	"os"
	"github.com/aloknerurkar/log-post/proto_spec"
	"fmt"
)

type LogStore interface {
	PrintMsg(s string)
}

type StdOutStore struct {
}

func (p *StdOutStore) PrintMsg(s string) {
	os.Stdout.WriteString(s)
}

type LogTask struct {
	log *LogPost.LogMsg
	storage LogStore
}

func generateLogMsg(prefix, id, msg string) string {
	return fmt.Sprintf("%s Client:%s %s", prefix, id, msg)
}

func (t *LogTask) Execute() {

	switch (t.log.LogLevel) {
	case int32(LogPost.LogLevel_TRACE):
		for idx := range t.log.Message {
			msg := generateLogMsg(TRACE_PREFIX, t.log.ClientId, t.log.Message[idx])
			t.storage.PrintMsg(msg)
		}
		break
	case int32(LogPost.LogLevel_INFO):
		for idx := range t.log.Message {
			msg := generateLogMsg(INFO_PREFIX, t.log.ClientId, t.log.Message[idx])
			t.storage.PrintMsg(msg)
		}
		break
	case int32(LogPost.LogLevel_ERROR):
		for idx := range t.log.Message {
			msg := generateLogMsg(ERROR_PREFIX, t.log.ClientId, t.log.Message[idx])
			t.storage.PrintMsg(msg)
		}
		break
	case int32(LogPost.LogLevel_FATAL):
		// FATAL msg expected to be max 1. No point sending multiple FATAL msgs.
		// ERROR should be used to send multiple fault cases.
		msg := generateLogMsg(FATAL_PREFIX, t.log.ClientId, t.log.Message[0])
		t.storage.PrintMsg(msg)
		break
	}
}
