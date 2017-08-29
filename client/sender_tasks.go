package client

import (
	"github.com/aloknerurkar/log-post/proto_spec"
	"context"
)

type sendListTask struct {
	sndr *sender
	lvl int32
	msgs []string
}

func newSendListTask(sendr *sender, level int32, msgList []string) *sendListTask {
	task := new(sendListTask)
	task.sndr = sendr
	task.lvl = level
	task.msgs = msgList
	return task
}

func (task *sendListTask) Execute() {
	msg := new(LogPost.LogMsg)
	msg.ClientId = task.sndr.clientID
	copy(msg.Message, task.msgs)
	msg.LogLevel = task.lvl

	sendProtoMsg(task.sndr, msg)
}

type sendTask struct {
	sndr *sender
	lvl int32
	msg string
}

func newSendTask(sendr *sender, level int32, message string) *sendTask {
	task := new(sendTask)
	task.sndr = sendr
	task.lvl = level
	task.msg = message
	return task
}

func (task *sendTask) Execute() {
	msg := new(LogPost.LogMsg)
	msg.ClientId = task.sndr.clientID
	msg.Message = make([]string, 1)
	msg.Message[0] = task.msg
	msg.LogLevel = task.lvl

	sendProtoMsg(task.sndr, msg)
}

func sendProtoMsg(s *sender, msg *LogPost.LogMsg) {
	conn := s.clientPool.Get()
	defer s.clientPool.Put(conn)

	client := LogPost.NewLogPostClient(conn)
	_, _ = client.LogReq(context.Background(), msg)
}
