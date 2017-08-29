package server

import (
	"flag"
	"net"
	"fmt"
	"google.golang.org/grpc/grpclog"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc"
	"github.com/aloknerurkar/log-post/proto_spec"
	"golang.org/x/net/context"
	"github.com/aloknerurkar/dumbDB"
	"github.com/aloknerurkar/task-runner"
	"os"
	"github.com/aloknerurkar/backend_utils"
	"github.com/golang/protobuf/proto"
	"log"
)

var (
	tls      = flag.Bool("tls", false, "Connection uses TLS if true, else plain TCP")
	certFile = flag.String("cert_file", "testdata/server1.pem", "The TLS cert file")
	keyFile  = flag.String("key_file", "testdata/server1.key", "The TLS key file")
	port     = flag.Int("port", 10060, "The server port")
)

type logServer struct {
	dbP *dumbDatabase.DumbDB
	tr *task_runner.TaskRunner
}

const (
	DB_NAME string = "log_post"
	BUCKET string = "log_post_clients"
	TRACE_PREFIX = "TRACE: "
	INFO_PREFIX = "INFO: "
	ERROR_PREFIX = "ERROR: "
	FATAL_PREFIX = "FATAL: "
)

func newServer() *logServer {
	s := new(logServer)
	s.dbP = dumbDatabase.NewDumbDB(".", DB_NAME, os.Stdout)
	s.tr = task_runner.StartTaskRunner(1, os.Stdout)
	return s
}

func (l *logServer) GetHeartBeat(c context.Context,
				 emp *LogPost.EmptyMessage) (ret_emp *LogPost.EmptyMessage, ret_err error) {
	ret_emp = emp
	ret_err = nil
	return
}

func (l *logServer) Register(c context.Context, ep *LogPost.EpInfo) (ret_conf *LogPost.ConfigResp, ret_err error) {

	ret_conf = new(LogPost.ConfigResp)
	eps, err := l.dbP.GetAll(BUCKET)

	// Use internal "err" to set return error before returning.
	defer func(e error) {
		if e != nil {
			ret_err = e
			log.Printf("Error registering log client. Error:%s", ret_err.Error())
		}
	}(err)

	if err == nil {
		for idx := range eps {
			epInfo := new(LogPost.EpInfo)
			err = proto.Unmarshal(eps[idx], epInfo)
			if err != nil {
				return
			}
			if epInfo.IpAddr == ep.IpAddr {
				ret_conf.ClientId = epInfo.ClientId
				return
			}
		}
	}

	ep.ClientId, err = backend_utils.NewUUID()
	if err != nil {
		return
	}

	record := make([][]byte, 2)
	record[0] = []byte(ep.ClientId)
	if err != nil {
		return
	}

	record[1], err = proto.Marshal(ep)
	if err != nil {
		return
	}

	err = l.dbP.Store(record, BUCKET)
	if err == nil {
		ret_conf.ClientId = ep.ClientId
		log.Printf("Assigned client ID %s to Ep: %+v", ret_conf.ClientId, ep)
	}
	return
}

// This is async so that we respond immediately.
func (l *logServer) LogReq(c context.Context, log *LogPost.LogMsg) (ret_emp *LogPost.EmptyMessage, ret_err error) {

	logTask := &LogTask {
		log: log,
		storage: &StdOutStore{},
	}

	l.tr.EnqueueTask(logTask)
	ret_emp = &LogPost.EmptyMessage{}
	ret_err = nil
	return
}

func main() {
	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		grpclog.Fatalf("failed to listen: %v", err)
	}
	var opts []grpc.ServerOption
	if *tls {
		creds, err := credentials.NewServerTLSFromFile(*certFile, *keyFile)
		if err != nil {
			grpclog.Fatalf("Failed to generate credentials %v", err)
		}
		opts = []grpc.ServerOption{grpc.Creds(creds)}
	}
	grpcServer := grpc.NewServer(opts...)
	LogPost.RegisterLogPostServer(grpcServer, newServer())
	grpcServer.Serve(lis)
}
