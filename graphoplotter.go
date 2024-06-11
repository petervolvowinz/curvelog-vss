/******** Peter Winzell (c), 5/22/24 *********************************************/

package main

import (
	"fmt"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
	"image"
	"image/color"
	"image/png"
	"log"
	"os"
	"time"
)

func drawLine(img *image.RGBA, x0, y0, x1, y1 int, clr color.RGBA) {
	// Bresenham's line algorithm
	dx := abs(x1 - x0)
	dy := abs(y1 - y0)
	sx := 1
	if x0 > x1 {
		sx = -1
	}
	sy := 1
	if y0 > y1 {
		sy = -1
	}
	err := dx - dy

	for {
		img.Set(x0, y0, clr)
		if x0 == x1 && y0 == y1 {
			break
		}
		e2 := 2 * err
		if e2 > -dy {
			if x0 == x1 {
				break
			}
			err -= dy
			x0 += sx
		}
		if e2 < dx {
			if y0 == y1 {
				break
			}
			err += dx
			y0 += sy
		}
	}
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func addLabel(img *image.RGBA, x, y int, label string) {
	col := color.RGBA{0, 0, 0, 255}
	point := fixed.Point26_6{fixed.I(x), fixed.I(y)}

	d := &font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(col),
		Face: basicfont.Face7x13,
		Dot:  point,
	}
	d.DrawString(label)
}

func mapRange(value, fromLow, fromHigh, toLow, toHigh float64) float64 {
	return (value-fromLow)*(toHigh-toLow)/(fromHigh-fromLow) + toLow
}

func xPlotWithOffset(val float64, woffset float64) float64 {
	return woffset + val
}

func getXY(dp TimeSeriesDataPoint, img *image.RGBA, width, height, woffset, hoffset, maxX, minX float64) (float64, float64) {

	xval := float64(dp.Timestamp)
	if xval == 0 {
		xval = minX
	}
	if maxX == minX {
		return 0, 0
	}
	x := mapRange(xval, minX, maxX, woffset, width-woffset) //woffset/2 + (width-woffset)*(xval/1000)
	y := height - mapRange(dp.Value, 0, 50, hoffset, height-hoffset)

	return x, y
}

func drawGraphData(img *image.RGBA, signalname string, currentsaveratio float64) {

	addLabel(img, 880, 20, signalname)
	addLabel(img, 880, 35, "Sample rate 100 ms")
	addLabel(img, 880, 50, "Cl max error = 1.0 ")
	currentsaveratio = currentsaveratio * 100
	outStr := fmt.Sprintf("Discarded samples = %d", uint64(currentsaveratio))
	addLabel(img, 880, 85, outStr)

}

func drawPNG(grpcvalue_1 chan string) {
	for {
		width := 1040
		height := 800
		woffset := 10
		hoffset := 10

		upLeft := image.Point{0, 0}
		lowRight := image.Point{width, height}

		img := image.NewRGBA(image.Rectangle{upLeft, lowRight})

		// Colors are defined by Red, Green, Blue, Alpha uint8 values.
		//cyan := color.RGBA{uint8(rand.Intn(100)), uint8(rand.Intn(200)), uint8(rand.Intn(200)), 0xff}

		// Set color for each pixel.
		axisColor := color.RGBA{100, 200, 200, 255}

		// Draw X-axis
		for x := 0; x < width-woffset; x++ {
			img.Set(x+woffset, height-10, axisColor)
		}

		// Draw Y-axis
		for y := hoffset; y < height-hoffset; y++ {
			img.Set(10, y, axisColor)
		}

		label := <-grpcvalue_1
		series := getMessage(label, &timeSeriesDataBufferNoCLog)

		if series != nil {
			oldx, oldy := -1000.0, -1000.0 // make sure values initiates to the first x,y values.
			lastdp := &TimeSeriesDataPoint{}

			for i := 0; i < len(series.TSeries); i++ {
				dp := series.TSeries[i]
				//gMutex.Lock()
				//fmt.Println("NONC", dp.Value, " ", dp.Timestamp, graphMetaData.maxX, " ", graphMetaData.minX)
				x, y := getXY(dp, img, float64(width), float64(height), float64(woffset), float64(hoffset), graphMetaData.maxX, graphMetaData.minX)
				//gMutex.Unlock()
				if oldx == -1000 {
					oldx = x
					oldy = y
				}
				drawLine(img, int(x), int(y), int(oldx), int(oldy), color.RGBA{255, 0, 0, 255})
				oldx = x
				oldy = y
				lastdp = &dp
			}
			addLabel(img, int(oldx-25), int(oldy-10), fmt.Sprintf("%dm/s", int64(lastdp.Value)))

			cLogMutex.Lock()
			oldx, oldy = -1000.0, -1000.0
			for i := 0; i < len(timeSeriesDataBufferCLog.TSeries); i++ {
				dp := timeSeriesDataBufferCLog.TSeries[i]
				// fmt.Println("CL", dp.Value, " ", dp.Timestamp, " ", graphMetaData.maxX, " ", graphMetaData.minX)
				x, y := getXY(dp, img, float64(width), float64(height), float64(woffset), float64(hoffset), graphMetaData.maxX, graphMetaData.minX)
				if oldx == -1000 {
					oldx = x
					oldy = y
				}
				drawLine(img, int(x), int(y), int(oldx), int(oldy), color.RGBA{0, 0, 0, 255})
				img.Set(int(x)-1, int(y), color.RGBA{0, 0, 0, 255})
				img.Set(int(x)+1, int(y), color.RGBA{0, 0, 0, 255})
				img.Set(int(x), int(y+1), color.RGBA{0, 0, 0, 255})
				img.Set(int(x), int(y-1), color.RGBA{0, 0, 0, 255})
				oldx = x
				oldy = y
			}
			ratio := float64(len(timeSeriesDataBufferCLog.TSeries)) / float64(len(timeSeriesDataBufferNoCLog.TSeries))
			fmt.Println("CURRENT SAVE RATIO = ", 1.0-ratio, "%")
			cLogMutex.Unlock()
			drawGraphData(img, "VSS: Vehicle.Speed", 1.0-ratio)
			// Encode as PNG.

			f, err := os.Create("assets/image1.png")
			if err != nil {
				log.Fatal(err)
			}
			if err := png.Encode(f, img); err != nil {
				f.Close()
				log.Fatal(err)
			}
			if err := f.Close(); err != nil {
				log.Fatal(err)
			}
			os.Rename("assets/image1.png", "assets/image.png")
			time.Sleep(20 * time.Millisecond)
		}
	}
}
