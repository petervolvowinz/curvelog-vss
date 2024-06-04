/******** Peter Winzell (c), 6/4/24 *********************************************/

package main

import (
	"encoding/json"
	"fmt"
	"testing"
)

var _commandList []string

func _initCommandList() {
	_commandList = make([]string, 2)

	_commandList[0] = `{"action":"subscribe","path":"Vehicle.Speed","filter":{"type":"curvelog","parameter":{"maxerr":"1","bufsize":"30"}},"requestId":"285"}`
	_commandList[1] = `{"action":"subscribe","path":"Vehicle.Speed","filter":{"type":"timebased","parameter":{"period":"100"}},"requestId":"286"}`
}

func getStructCommands()(Command,Command){
	command_1 := Command{
		Action: "subscribe",
		Path:   "vehicle.speed",
		Filter: Filter{
			Type:      "curvelog",
			Parameter: CurveLogParameter{MaxErr: "1.0", BufSize: "20"},
		},
		RequestID: "300",
	}

	command_2 := Command{
		Action: "subscribe",
		Path:   "vehicle.speed",
		Filter: Filter{
			Type:      "timebased",
			Parameter: TimeFilterParameter{Period: "100"},
		},
		RequestID: "301",
	}

	return command_1,command_2
}

func TestCommand(t *testing.T) {
	command_1,command_2 := GenerateCommands()
	_initCommandList()
	var test_cmd_1 Command
	err := json.Unmarshal([]byte(_commandList[0]), &test_cmd_1)
	if err != nil {
		fmt.Println("Error:", err)
		t.Error(" could not unmarshal test ", err)
	}

	var test_cmd_2 Command
	err = json.Unmarshal([]byte(_commandList[1]), &test_cmd_2)
	if err != nil {
		fmt.Println("Error:", err)
		t.Error(" could not unmarshal test ", err)
	}

	if command_1
}
