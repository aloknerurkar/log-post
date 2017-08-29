package client

import (
	"log"
	"fmt"
	utils "github.com/aloknerurkar/backend_utils"
	"os"
	"github.com/aloknerurkar/log-post/proto_spec"
	"io/ioutil"
)

const LOG_FLAGS = log.Ldate | log.Ltime | log.Lmicroseconds | log.Llongfile

type LogPostClient struct {
	PkgName string
	PkgVersion string

	remoteSender bool
	sndr *sender

	trace *log.Logger
	info *log.Logger
	error *log.Logger
	fatal *log.Logger
}

func InitLogger(pkgName, pkgVersion string, logLevel int32, remote_sender bool) *LogPostClient {

	client := new(LogPostClient)
	client.PkgName = pkgName
	client.PkgVersion = pkgVersion
	client.remoteSender = remote_sender

	endpoints := make([]utils.ConnEndpointInfo, 1)
	endpoints[0] = utils.ConnEndpointInfo{
		Tls: false,
		CertFile: "",
		ServerHostOverride: "",
		ServerAddr: "localhost:10000",

	}

	if client.remoteSender {
		var err error
		client.sndr, err = startSender(endpoints, pkgName, pkgVersion)
		if err != nil {
			log.Fatalf("Failed starting sender Err:%v", err)
		}
	}

	if logLevel & int32(LogPost.LogLevel_TRACE) > 0 {
		if client.remoteSender {
			client.trace = log.New(client.sndr.trace, client.PkgId(), LOG_FLAGS)
		} else {
			client.trace = log.New(os.Stdout, "TRACE\t" + client.PkgId(), LOG_FLAGS)
		}
	} else {
		client.trace = log.New(ioutil.Discard, "", LOG_FLAGS)
	}

	if logLevel & int32(LogPost.LogLevel_INFO) > 0 {
		if client.remoteSender {
			client.info = log.New(client.sndr.info, client.PkgId(), LOG_FLAGS)
		} else {
			client.info = log.New(os.Stdout, "INFO\t" + client.PkgId(), LOG_FLAGS)
		}
	} else {
		client.info = log.New(ioutil.Discard, "", LOG_FLAGS)
	}

	if logLevel & int32(LogPost.LogLevel_ERROR) > 0 {
		if client.remoteSender {
			client.error = log.New(client.sndr.error, client.PkgId(), LOG_FLAGS)
		} else {
			client.error = log.New(os.Stdout, "ERROR\t" + client.PkgId(), LOG_FLAGS)
		}
	} else {
		client.error = log.New(ioutil.Discard, "", LOG_FLAGS)
	}

	if logLevel & int32(LogPost.LogLevel_FATAL) > 0 {
		if client.remoteSender {
			client.fatal = log.New(client.sndr.fatal, client.PkgId(), LOG_FLAGS)
		} else {
			client.fatal = log.New(os.Stdout, "FATAL\t" + client.PkgId(), LOG_FLAGS)
		}
	} else {
		client.fatal = log.New(ioutil.Discard, "", LOG_FLAGS)
	}

	return client
}

func (l *LogPostClient) PkgId() string {
	return l.PkgName + ":" + l.PkgVersion
}

func (l *LogPostClient) TraceStart(format string, args... interface{}) {
	msg := fmt.Sprintf(format, args...)
	msg_wfuncname := fmt.Sprintf("Started FUNC:%s %s", utils.MyCaller(), msg)
	_ = l.trace.Output(3, msg_wfuncname)
}

func (l *LogPostClient) TraceEnd(format string, args... interface{}) {
	msg := fmt.Sprintf(format, args...)
	msg_wfuncname := fmt.Sprintf("End FUNC:%s %s", utils.MyCaller(), msg)
	_ = l.trace.Output(3, msg_wfuncname)
}

func (l *LogPostClient) Info(format string, args... interface{}) {
	msg := fmt.Sprintf(format, args...)
	msg_wfuncname := fmt.Sprintf("FUNC:%s %s", utils.MyCaller(), msg)
	_ = l.info.Output(3, msg_wfuncname)
}

func (l *LogPostClient) Error(e error, format string, args... interface{}) error {
	msg := fmt.Sprintf(format, args...)
	msg_wfuncname := fmt.Sprintf("ERR:%s FUNC:%s %s", e.Error(), utils.MyCaller(), msg)
	_ = l.error.Output(3, msg_wfuncname)
	return e
}

func (l *LogPostClient) Fatal(e error, format string, args... interface{}) {
	msg := fmt.Sprintf(format, args...)
	msg_wfuncname := fmt.Sprintf("ERR_FATAL:%s FUNC:%s %s", e.Error(), utils.MyCaller(), msg)
	_ = l.fatal.Output(3, msg_wfuncname)
	// Check handling of fatal stuff. Panic/Recover? Currently process exit may cause loss of msg.
	os.Exit(1)
}