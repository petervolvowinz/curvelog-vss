/******** Peter Winzell (c), 6/12/24 *********************************************/

package main

import (
	"fmt"
	"math"
	"sync"
)

type Counter interface {
	IncSampleCounter() int
	IncDownSampleCounter() int
	Ratio() int
}
type sampleCounters struct {
	sample_counter     int
	downsample_counter int
}

// increase sample_counter with 1
func (counter *sampleCounters) IncSampleCounter() int {
	counter.sample_counter++
	return counter.sample_counter
}

// increase downsample_cpunter with 1
func (counter *sampleCounters) IncDownSampleCounter() int {
	counter.downsample_counter++
	return counter.downsample_counter
}

// the percentage of samples not used
func (counter sampleCounters) Ratio() int {
	return int(math.Round((1.0 - float64(counter.downsample_counter)/float64(counter.sample_counter)) * 100))
}

var counterlock = &sync.Mutex{}
var counters *sampleCounters

// there can be only one counter of counters
func GetCountersInstance() Counter {
	if counters == nil {
		counterlock.Lock()
		defer counterlock.Unlock()
		if counters == nil {
			counters = &sampleCounters{}
		} else {
			fmt.Println("Single instance Counter already created.")
		}
	} else {
		fmt.Println("Single instance Counter already created.")
	}
	return counters
}
