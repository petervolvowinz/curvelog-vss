/******** Peter Winzell (c), 6/5/24 *********************************************/

package main

import (
	"fmt"
	"github.com/fogleman/gg"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/goregular"
	"image"
	"image/color"
	"log"
	"os"
	"time"
)

func DoFoglemanPNG() {
	font, err := truetype.Parse(goregular.TTF)
	if err != nil {
		log.Fatal(err)
	}

	face := truetype.NewFace(font, &truetype.Options{Size: 48})

	dc := gg.NewContext(1024, 1024)
	dc.SetFontFace(face)
	dc.SetRGB(1, 1, 1)
	dc.Clear()
	dc.SetRGB(0, 0, 0)
	dc.DrawStringAnchored("Hello, world!", 512, 512, 0.5, 0.5)
	dc.SavePNG("out.png")
}

func GetFontFace(size float64) font.Face {
	font, err := truetype.Parse(goregular.TTF)
	if err != nil {
		log.Fatal(err)
	}

	return truetype.NewFace(font, &truetype.Options{Size: size})
}

func _addLabel(dc *gg.Context, x, y float64, label string) {
	face := GetFontFace(18)
	dc.SetFontFace(face)
	dc.DrawString(label, x-40, y)
	dc.Stroke()
}

func _getXY(dp TimeSeriesDataPoint, width, height, woffset, hoffset, maxX, minX float64) (float64, float64) {

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

func _drawGraphData(dc *gg.Context, currentsaveratio float64) {

	dc.SetColor(color.Black)
	_addLabel(dc, 840, 20, JsonSettings.VssName)
	sampleRateStr := fmt.Sprintf("Sample rate %s%s", JsonSettings.SubPeriod, " ms")
	_addLabel(dc, 840, 35, sampleRateStr)
	curveMaxError := fmt.Sprintf("Cl max error = %s", JsonSettings.CurveLogErr)
	_addLabel(dc, 840, 50, curveMaxError)
	currentsaveratio = currentsaveratio * 100
	outStr := fmt.Sprintf("Discarded samples = %d", uint64(currentsaveratio))
	_addLabel(dc, 840, 75, outStr)

}

// draws the y and x axis and put it in an image for reuse
var backgroundImage image.Image = nil

func getBackgroundImage(width, height, woffset, hoffset int) image.Image {
	if backgroundImage == nil {
		dc := gg.NewContext(width, height)
		dc.SetRGBA255(255, 255, 255, 128)
		dc.Clear() // fill image with this color.
		dc.SetLineWidth(1)
		dc.SetRGB255(0, 0, 0)
		dc.DrawLine(float64(woffset), float64(height-hoffset), float64(width-woffset), float64(height-hoffset))
		dc.DrawLine(float64(width-woffset), float64(height-hoffset)-3, float64(width-woffset), float64(height-hoffset)+3)
		dc.DrawLine(float64(woffset), float64(hoffset), float64(woffset), float64(height-hoffset))
		dc.DrawLine(float64(woffset)-3, float64(hoffset), float64(woffset)+3, float64(hoffset))
		dc.Stroke()
		face := GetFontFace(18)
		dc.SetFontFace(face)
		dc.DrawStringAnchored("t", float64(width-woffset)+3, float64(height-hoffset), 0.5, 0.5)
		dc.Stroke()
		dc.SetRGB255(96, 96, 96)
		dc.DrawStringAnchored("t", float64(width-woffset)+4, float64(height-hoffset)+1, 0.5, 0.5)
		dc.Stroke()
		backgroundImage = dc.Image()
	}

	return backgroundImage
}

// show a rectangle for each curvelogged point picked
var curvelogPickerDot image.Image = nil

func getcurvelogPickerDot() image.Image {

	if curvelogPickerDot == nil {
		dc := gg.NewContext(5, 5)
		dc.SetRGB255(255, 255, 255)
		dc.Clear()
		dc.SetRGB255(0, 0, 0)
		dc.DrawRectangle(0, 0, 4, 4)
		dc.Stroke()
		curvelogPickerDot = dc.Image()
	}
	return curvelogPickerDot

}

func DrawPNGgg(grpcvalue_1 chan string) {
	for {
		width := 1040
		height := 800
		woffset := 10
		hoffset := 10

		dc := gg.NewContext(width, height)
		dc.DrawImage(getBackgroundImage(width, height, woffset, hoffset), 0, 0)
		label := <-grpcvalue_1 // waiting for data
		series := getMessage(label, &timeSeriesDataBufferNoCLog)
		//plot time series data points
		if series != nil {
			oldx, oldy := -1000.0, -1000.0 // make sure values initiates to the first x,y values.
			lastdp := &TimeSeriesDataPoint{}
			dc.SetRGB255(255, 0, 0)
			for i := 0; i < len(series.TSeries); i++ {
				dp := series.TSeries[i]
				x, y := _getXY(dp, float64(width), float64(height), float64(woffset), float64(hoffset), graphMetaData.maxX, graphMetaData.minX)
				if oldx == -1000 {
					oldx = x
					oldy = y
				}
				dc.DrawLine(x, y, oldx, oldy)
				oldx = x
				oldy = y
				lastdp = &dp
			}
			dc.Stroke()
			// dc.Clear()
			dc.SetColor(color.Black)
			_addLabel(dc, oldx, oldy-10, fmt.Sprintf("%dm/s", int64(lastdp.Value)))
			dc.Stroke()
			cLogMutex.Lock()
			oldx, oldy = -1000.0, -1000.0
			dc.SetRGB255(0, 0, 255)
			for i := 0; i < len(timeSeriesDataBufferCLog.TSeries); i++ {
				dc.SetRGB255(0, 0, 255)
				dp := timeSeriesDataBufferCLog.TSeries[i]
				// fmt.Println("CL", dp.Value, " ", dp.Timestamp, " ", graphMetaData.maxX, " ", graphMetaData.minX)
				x, y := _getXY(dp, float64(width), float64(height), float64(woffset), float64(hoffset), graphMetaData.maxX, graphMetaData.minX)
				if oldx == -1000 {
					oldx = x
					oldy = y
				}
				dc.SetLineWidth(1)
				dc.DrawLine(x, y, oldx, oldy)
				dc.Stroke()
				dc.DrawImage(getcurvelogPickerDot(), int(x-2), int(y-2))

				oldx = x
				oldy = y
			}
			dc.Stroke()
			ratio := float64(len(timeSeriesDataBufferCLog.TSeries)) / float64(len(timeSeriesDataBufferNoCLog.TSeries))
			fmt.Println("CURRENT SAVE RATIO = ", 1.0-ratio, "%")
			cLogMutex.Unlock()

			_drawGraphData(dc, 1.0-ratio)

			dc.SavePNG("assets/image1.png")
			time.Sleep(20 * time.Millisecond)
			os.Rename("assets/image1.png", "assets/image.png")
		}

	}
}
