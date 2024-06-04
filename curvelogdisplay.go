/******** Peter Winzell (c), 4/15/24 *********************************************/

package main

import (
	"context"
	"fmt"
	"github.com/akamensky/argparse"
	pb "github.com/covesa/vissr/grpc_pb"
	"github.com/covesa/vissr/utils"
	"net/http"
)

var grpcCompression utils.Compression

func getVISSStream(command string, ctx context.Context) pb.VISSv2_SubscribeRequestClient {

	vssRequest := command
	grpcCompression = utils.PB_LEVEL1
	pbRequest := utils.SubscribeRequestJsonToPb(vssRequest, grpcCompression)

	client := getVISSClient(getGRPCServerConnection())
	stream, _ := client.SubscribeRequest(ctx, pbRequest)
	return stream
}

func retrieveValue(rpcvalue_1 chan string) {

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	vssRequest := commandList[1]
	grpcCompression = utils.PB_LEVEL1
	pbRequest := utils.SubscribeRequestJsonToPb(vssRequest, grpcCompression)

	client := getVISSClient(getGRPCServerConnection())
	stream, _ := client.SubscribeRequest(ctx, pbRequest)

	for {
		pbResponse, err := stream.Recv()
		if err != nil {
			fmt.Printf("Error=%v when issuing request=:%s", err, vssRequest)
			break
		}
		vssResponse := utils.SubscribeStreamPbToJson(pbResponse, grpcCompression)
		// fmt.Printf("Received response:%s\n", vssResponse)
		rpcvalue_1 <- vssResponse
		// time.Sleep(1000)
	}
}

func getCurveLogValues() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	vissrequest := commandList[0]
	stream := getVISSStream(commandList[0], ctx)
	for {
		pbResponse, err := stream.Recv()
		if err != nil {
			fmt.Printf("Error=%v when issuing request=:%s", err, vissrequest)
			break
		}
		vissResponse := utils.SubscribeStreamPbToJson(pbResponse, utils.PB_LEVEL1)
		cLogMutex.Lock()
		getMessage(vissResponse, &timeSeriesDataBufferCLog)
		cLogMutex.Unlock()
		fmt.Println("curvelog response ", vissResponse)
	}
}

func main() {

	parser := argparse.NewParser("print", "curve log display server ") // The root node name Vehicle must be synched with the feeder-registration.json file.

	logFile := parser.Flag("", "logfile", &argparse.Options{Required: false, Help: "outputs to logfile in ./logs folder"})
	logLevel := parser.Selector("", "loglevel", []string{"trace", "debug", "info", "warn", "error", "fatal", "panic"}, &argparse.Options{
		Required: false,
		Help:     "changes log output level",
		Default:  "info"})

	utils.InitLog("feeder-log.txt", "./logs", *logFile, *logLevel)
	initCommandList()
	grpcvalue_1 := make(chan string, 1)

	go getCurveLogValues()
	go retrieveValue(grpcvalue_1)
	go drawPNG(grpcvalue_1)

	//web server
	http.HandleFunc("/about/", about_handler)
	http.HandleFunc("/plotter", dynamicHandler)
	http.Handle("/", http.FileServer(http.Dir("./assets")))

	http.ListenAndServe(":9000", nil)
}
