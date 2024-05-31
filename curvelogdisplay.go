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

func index_handler(w http.ResponseWriter, r *http.Request) {
	// MAIN SECTION HTML CODE
	fmt.Fprintf(w, "<h1>Whoa, Go is neat!</h1>")
	fmt.Fprintf(w, "<title>Go</title>")
	fmt.Fprintf(w, "<img src='assets/plotter.png' alt='gopher' style='width:1024px;height:800px;'>")
}

func about_handler(w http.ResponseWriter, r *http.Request) {
	// ABOUT SECTION HTML CODE
	fmt.Fprintf(w, "<title>Go/about/</title>")
	fmt.Fprintf(w, "Plotting data from signal broker")
}

func dynamicHandler2(w http.ResponseWriter, r *http.Request) {
	html := `<!DOCTYPE html>
				<html lang="en">
				<head>
					<meta charset="UTF-8">
					<meta name="viewport" content="width=device-width, initial-scale=1.0">
					<title>Refresh PNG Image</title>
				<script>
						// Function to refresh the image
						function refreshImage() {
							var img = document.getElementById('image');
							img.src = img.src.split('?')[0] + '?' + new Date().getTime();
						}

						// Function to refresh the image every second
						function startRefreshing() {
							setInterval(refreshImage, 500);
						}
				</script>
				</head>
				<body onload="startRefreshing()">
					<!-- Replace 'image.png' with your image file path -->
					<img id="image" src="image.png" alt="Image">
				</body>
		</html>`
	fmt.Fprintf(w, html)
}

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
	http.HandleFunc("/mupp", index_handler)
	http.HandleFunc("/about/", about_handler)
	http.HandleFunc("/plotter", dynamicHandler2)

	http.Handle("/", http.FileServer(http.Dir("./assets")))

	http.ListenAndServe(":9000", nil)
}
