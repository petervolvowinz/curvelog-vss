/******** Peter Winzell (c), 6/4/24 *********************************************/

package main

import (
	"encoding/json"
	"fmt"
	"os"
)

type Settings struct {
	VssName     string `json:"vss-name"`
	SubPeriod   string `json:"sub-period-ms"`
	CurveLogErr string `json:"curve-log-err"`
	CurveLogBuf string `json:"curve-log-buf"`
	Adress      string `json:"adress"`
	PortNo      int    `json:"port"`
}

var JsonSettings *Settings
var commandList []string

func InitCommandList() {
	commandList = make([]string, 2)
	commandList[0], commandList[1] = GenerateCommands()
}

func (settings *Settings) GetSettings() {
	jsonFile, err := os.Open("settings.json")
	if err != nil {
		fmt.Println("Error opening JSON file:", err)
		return
	}
	defer jsonFile.Close()
	decoder := json.NewDecoder(jsonFile)
	err = decoder.Decode(settings)
	if err != nil {
		fmt.Println("Error decoding JSON data:", err)
		return
	}
}

type Command struct {
	Action    string `json:"action"`
	Path      string `json:"path"`
	Filter    Filter `json:"filter"`
	RequestID string `json:"requestId"`
}

type Filter struct {
	Type      string      `json:"type"`
	Parameter interface{} `json:"parameter"`
}

type CurveLogParameter struct {
	MaxErr  string `json:"maxerr"`
	BufSize string `json:"bufsize"`
}

type TimeFilterParameter struct {
	Period string `json:"period"`
}

func GenerateCommands() (string, string) {
	JsonSettings = &Settings{}
	JsonSettings.GetSettings()

	command_1 := Command{
		Action: "subscribe",
		Path:   JsonSettings.VssName,
		Filter: Filter{
			Type:      "curvelog",
			Parameter: CurveLogParameter{MaxErr: JsonSettings.CurveLogErr, BufSize: JsonSettings.CurveLogBuf},
		},
		RequestID: "300",
	}

	command_2 := Command{
		Action: "subscribe",
		Path:   JsonSettings.VssName,
		Filter: Filter{
			Type:      "timebased",
			Parameter: TimeFilterParameter{Period: JsonSettings.SubPeriod},
		},
		RequestID: "301",
	}

	curveCommand, err := json.Marshal(command_1)
	if err != nil {
		fmt.Println(err)
		panic(err)
	}

	timeCommand, err := json.Marshal(command_2)
	if err != nil {
		fmt.Println(err)
		panic(err)
	}

	return string(curveCommand), string(timeCommand)
}
