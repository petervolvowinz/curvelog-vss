/******** Peter Winzell (c), 4/15/24 *********************************************/

package main

import (
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"sync"
	"time"
)

type DataPoint struct {
	Value string `json:"value"`
	Ts    string `json:"ts"`
}

type Data struct {
	Path string      `json:"path"`
	Dp   interface{} `json:"dp"`
}

type Message struct {
	Action         string    `json:"action"`
	Data           Data      `json:"data"`
	Ts             time.Time `json:"ts"`
	SubscriptionId string    `json:"subscriptionId"`
}

type TimeSeriesDataPoint struct {
	Timestamp int64
	Value     float64
}

type TimeSeriesDataPoints struct {
	TSeries   []TimeSeriesDataPoint
	Maxlength int
}

type GraphData struct {
	maxX float64
	minX float64
}

var gMutex sync.Mutex
var graphMetaData = GraphData{
	0,
	math.MaxFloat64,
}

var timeSeriesDataBufferNoCLog = TimeSeriesDataPoints{
	TSeries:   make([]TimeSeriesDataPoint, 1),
	Maxlength: 1000,
}

var cLogMutex sync.Mutex
var timeSeriesDataBufferCLog = TimeSeriesDataPoints{
	TSeries:   make([]TimeSeriesDataPoint, 1),
	Maxlength: 1000,
}

func (tsb *TimeSeriesDataPoints) AddPoint(tspoint TimeSeriesDataPoint) {
	// Check if buffer exceeds maximum length, if so, remove oldest data points

	timeinmillis := float64(tspoint.Timestamp)
	gMutex.Lock()
	if timeinmillis > graphMetaData.maxX {
		graphMetaData.maxX = timeinmillis
	}
	if timeinmillis <= graphMetaData.minX {
		graphMetaData.minX = timeinmillis
	}
	gMutex.Unlock()

	if len(tsb.TSeries) > tsb.Maxlength {
		numToRemove := len(tsb.TSeries) - tsb.Maxlength
		tsb.TSeries = tsb.TSeries[numToRemove:]
		graphMetaData.minX = float64(tsb.TSeries[0].Timestamp)
		removePointsOutsideOfGraph(&timeSeriesDataBufferCLog)
	}
	tspoint.Timestamp = tspoint.Timestamp
	tsb.TSeries = append(tsb.TSeries, tspoint)
}

func removePointsOutsideOfGraph(series *TimeSeriesDataPoints) *TimeSeriesDataPoints {
	result := []TimeSeriesDataPoint{}
	skip := make(map[int]bool)

	slice := series.TSeries

	for index, elem := range slice {
		if float64(elem.Timestamp) < graphMetaData.minX {
			skip[index] = true
		}
	}

	for i, v := range slice {
		if !skip[i] {
			result = append(result, v)
		}
	}
	series.TSeries = result
	return series
}

func getTimeSeriesDataPoint(dp map[string]interface{}) *TimeSeriesDataPoint {
	// parsedTime := time.Now().UnixMilli() // ,
	// tdp, err := time.Parse(time.RFC3339Nano, dp["ts"].(string))
	tdpstr := dp["ts"].(string)
	tdp, err := strconv.ParseInt(tdpstr, 10, 64)
	if err != nil {
		fmt.Println("cannot covert ts to int64")
	} else {
		seconds := tdp / 1000
		nanoseconds := (tdp % 1000) * 1000000

		// Convert to time.Time
		t := time.Unix(seconds, nanoseconds)

		// Format time as RFC3339Nano
		rfc3339NanoTime := t.Format(time.RFC3339Nano)

		fmt.Println("RFC3339Nano time:", rfc3339NanoTime)
	}

	f, err := strconv.ParseFloat(dp["value"].(string), 64)
	if err != nil {
		fmt.Println("Error:", err)
		return nil
	}
	return &TimeSeriesDataPoint{
		tdp,
		f,
	}
}

func getMessage(vssjson string, timeSeriesDataBuffer *TimeSeriesDataPoints) *TimeSeriesDataPoints {
	var msg Message
	err := json.Unmarshal([]byte(vssjson), &msg)
	if err != nil {
		fmt.Println("Error:", err)
	}
	switch dp := msg.Data.Dp.(type) {
	case []interface{}:
		//fmt.Println("Multiple Data Points:")
		for _, item := range dp {
			dataPoint := item.(map[string]interface{})
			/*value := dataPoint["value"].(string)
			ts := dataPoint["ts"].(string)
			fmt.Printf("  Value: %s, Timestamp: %s\n", value, ts)*/
			timeSeriesDataBuffer.AddPoint(*getTimeSeriesDataPoint(dataPoint))
		}
		return timeSeriesDataBuffer
	default:
		if dp != nil {
			//fmt.Println("Single Data Point:")
			dataPoint := dp.(map[string]interface{})
			/*value := dataPoint["value"].(string)
			ts := dataPoint["ts"].(string)
			fmt.Printf("  Value: %s, Timestamp: %s\n", value, ts)*/
			timeSeriesDataBuffer.AddPoint(*getTimeSeriesDataPoint(dataPoint))
			return timeSeriesDataBuffer
		}
	}
	return nil
}
