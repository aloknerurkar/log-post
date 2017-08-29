package client

import (
	"github.com/aloknerurkar/log-post/proto_spec"
	utils "github.com/aloknerurkar/backend_utils"
	"google.golang.org/grpc"
	"context"
	"github.com/aloknerurkar/task-runner"
	"os"
	"errors"
)

const (
	DEFAULT_SENDER_QUEUE_SIZE = 100
)

type writeMsg struct {
	buf []byte
	len int
}

type remoteWriter struct {
	msgChan chan writeMsg
}

func newRemoteWriter(msg_chan chan writeMsg) *remoteWriter {
	writer := new(remoteWriter)
	writer.msgChan = msg_chan
	return writer
}

func (r *remoteWriter) Write(msg_buf []byte) (int, error) {
	msg := writeMsg{
		buf: msg_buf,
		len: len(msg_buf),
	}
	r.msgChan <- msg
	return msg.len, nil
}

type sender struct {

	clientID string

	trace 	*remoteWriter
	tchan	chan writeMsg
	traceQ	[]string

	info 	*remoteWriter
	ichan	chan writeMsg
	infoQ	[]string

	error 	*remoteWriter
	echan	chan writeMsg

	fatal 	*remoteWriter
	fchan	chan writeMsg

	clientPool	*utils.RpcClientPool
	runner		*task_runner.TaskRunner
}

var doHeartBeat = func(conn *grpc.ClientConn) error {
	c := LogPost.NewLogPostClient(conn)
	_, err := c.GetHeartBeat(context.Background(), &LogPost.EmptyMessage{})
	return err
}

func startSender(endpoints []utils.ConnEndpointInfo, svc_name, svc_version string) (*sender, error) {

	sndr := new(sender)
	sndr.clientPool = utils.NewRpcClientPool(doHeartBeat, endpoints, 1, os.Stdout)
	sndr.runner = task_runner.StartTaskRunner(2, os.Stdout)

	sndr.tchan = make(chan writeMsg, DEFAULT_SENDER_QUEUE_SIZE)
	sndr.ichan = make(chan writeMsg, DEFAULT_SENDER_QUEUE_SIZE)
	sndr.echan = make(chan writeMsg, DEFAULT_SENDER_QUEUE_SIZE)
	sndr.fchan = make(chan writeMsg, DEFAULT_SENDER_QUEUE_SIZE)

	sndr.trace = newRemoteWriter(sndr.tchan)
	sndr.info = newRemoteWriter(sndr.ichan)
	sndr.error = newRemoteWriter(sndr.echan)
	sndr.fatal = newRemoteWriter(sndr.fchan)

	conn := sndr.clientPool.Get()
	defer sndr.clientPool.Put(conn)

	ip, err := utils.ExternalIP()
	if err != nil {
		return nil, errors.New("Failed to get external IP.")
	}

	client := LogPost.NewLogPostClient(conn)
	epInfo := &LogPost.EpInfo{
		IpAddr: ip,
		ServiceName: svc_name,
		ServiceVersion: svc_version,
	}
	resp, err := client.Register(context.Background(), epInfo)
	if err != nil {
		return nil, errors.New("Failed to register sender.")
	}

	sndr.clientID = resp.ClientId

	go func(sender *sender) {
		for {
			select {
			case msg := <- sender.tchan:
				str_msg := string(msg.buf[:msg.len])
				sender.traceQ = append(sender.traceQ, str_msg)
				if (len(sender.traceQ) == 50) {
					task := newSendListTask(sender, int32(LogPost.LogLevel_TRACE), sender.traceQ)
					sender.runner.EnqueueTask(task)
				}

			case msg := <- sender.ichan:
				str_msg := string(msg.buf[:msg.len])
				sender.infoQ = append(sender.infoQ, str_msg)
				if (len(sender.infoQ) == 50) {
					task := newSendListTask(sender, int32(LogPost.LogLevel_INFO), sender.infoQ)
					sender.runner.EnqueueTask(task)
				}

			case msg := <- sender.echan:
				str_msg := string(msg.buf[:msg.len])
				task := newSendTask(sender, int32(LogPost.LogLevel_ERROR), str_msg)
				sender.runner.EnqueueTask(task)

			case msg := <- sender.fchan:
				str_msg := string(msg.buf[:msg.len])
				task := newSendTask(sender, int32(LogPost.LogLevel_FATAL), str_msg)
				sender.runner.EnqueueTask(task)
			}
		}
	}(sndr)

	return sndr, nil
}


