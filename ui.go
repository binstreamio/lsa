package main

import (
	"fmt"
	"time"

	ui "gopkg.in/gizak/termui.v2"
)

const (
	uiDataListSize = 120
)

var (
	videoFrameRateList    = make([]float64, uiDataListSize)
	videoKeyFrameRateList = make([]float64, uiDataListSize)
	videoKeyFrameTagList  = make([]string, uiDataListSize)
	audioFrameRateList    = make([]float64, uiDataListSize)
	streamBandwidth       = make([]float64, uiDataListSize)

	processInfoPar *ui.Par
	streamInfoPar  *ui.Par
	processLogPar  *ui.Par
)

func initStreamInfoBox() (*ui.Par, *ui.Par) {
	parProcess := ui.NewPar("")
	parProcess.Height = 6
	parProcess.BorderLabel = "Process Info"

	parStream := ui.NewPar("")
	parStream.Height = 6
	parStream.BorderLabel = "Stream Info"

	ui.Body.AddRows(
		ui.NewRow(
			ui.NewCol(9, 0, parStream),
			ui.NewCol(3, 0, parProcess),
		),
	)

	return parProcess, parStream
}

func initStreamLogBox() *ui.Par {
	parProcess := ui.NewPar("")
	parProcess.Height = 8
	parProcess.BorderLabel = "Process Log"

	ui.Body.AddRows(
		ui.NewRow(
			ui.NewCol(12, 0, parProcess),
		),
	)

	return parProcess
}

func initAudioInfoBox() {
	lcFrames := ui.NewLineChart()
	lcFrames.BorderLabel = "Audio: frames"
	lcFrames.Data = audioFrameRateList
	lcFrames.Mode = "dot"
	lcFrames.Height = 11
	lcFrames.AxesColor = ui.ColorWhite
	lcFrames.LineColor = ui.ColorYellow | ui.AttrBold

	lcKeyFrames := ui.NewLineChart()
	lcKeyFrames.BorderLabel = "Bandwidth: Mbps"
	lcKeyFrames.Data = streamBandwidth
	lcKeyFrames.Mode = "dot"
	lcKeyFrames.Height = 11
	lcKeyFrames.AxesColor = ui.ColorWhite
	lcKeyFrames.LineColor = ui.ColorYellow | ui.AttrBold

	ui.Body.AddRows(
		ui.NewRow(
			ui.NewCol(6, 0, lcFrames),
			ui.NewCol(6, 0, lcKeyFrames),
		),
	)
}

func initVideoInfoBox() {
	lcFrames := ui.NewLineChart()
	lcFrames.BorderLabel = "Video: frames"
	lcFrames.Data = videoFrameRateList
	lcFrames.Mode = "dot"
	lcFrames.Height = 11
	lcFrames.AxesColor = ui.ColorWhite
	lcFrames.LineColor = ui.ColorYellow | ui.AttrBold

	lcKeyFrames := ui.NewLineChart()
	lcKeyFrames.BorderLabel = "Video: key frames"
	lcKeyFrames.Data = videoKeyFrameRateList
	lcKeyFrames.Height = 11
	lcKeyFrames.AxesColor = ui.ColorWhite
	lcKeyFrames.LineColor = ui.ColorYellow

	ui.Body.AddRows(
		ui.NewRow(
			ui.NewCol(6, 0, lcFrames),
			ui.NewCol(6, 0, lcKeyFrames),
		),
	)
}

func updateUI(curInfo, lastInfo *streamInfo) {
	// process info
	if !curInfo.ProcessStartTime.IsZero() {
		processInfoPar.Text = fmt.Sprintf("%s\n\nStart on: %s\nDuration: %s",
			time.Now().Format(time.Stamp),
			curInfo.ProcessStartTime.Format(time.Stamp),
			time.Now().Sub(curInfo.ProcessStartTime))
	}

	streamInfoPar.Text = curInfo.StreamInfo

	processLogPar.Text = curInfo.AnalyzingLog

	vFPS := float64(curInfo.VideoFrames - lastInfo.VideoFrames)
	vKFPS := float64(curInfo.VideoKeyFrames - lastInfo.VideoKeyFrames)
	aFPS := float64(curInfo.AudioFrames - lastInfo.AudioFrames)
	streamBW := float64(curInfo.ReceivedSize-lastInfo.ReceivedSize) / 1024 / 1024 * 8

	for i := uiDataListSize - 1; i >= 2; i-- {
		videoFrameRateList[i] = videoFrameRateList[i-1]
		videoKeyFrameRateList[i] = videoKeyFrameRateList[i-1]
		audioFrameRateList[i] = audioFrameRateList[i-1]
		streamBandwidth[i] = streamBandwidth[i-1]
	}
	videoFrameRateList[1] = vFPS
	videoKeyFrameRateList[1] = vKFPS
	audioFrameRateList[1] = aFPS
	streamBandwidth[1] = streamBW

	ui.Render(ui.Body)
}

func streamInfoRecv(streamInfoChan chan *streamInfo) {
	lastInfo := streamInfo{}
	curInfo := streamInfo{}

	ticker := time.NewTicker(time.Second)
	for {
		select {
		case sInfo := <-streamInfoChan:
			curInfo = *sInfo
		case <-ticker.C:
			updateUI(&curInfo, &lastInfo)

			lastInfo = curInfo
		}
	}
}

func drawUI(streamInfoChan chan *streamInfo) {
	go streamInfoRecv(streamInfoChan)

	err := ui.Init()
	if err != nil {
		panic(err)
	}
	defer ui.Close()

	ui.Handle("/sys/kbd/q", func(ui.Event) {
		// press q to quit
		ui.StopLoop()
	})

	processInfoPar, streamInfoPar = initStreamInfoBox()
	initVideoInfoBox()
	initAudioInfoBox()
	processLogPar = initStreamLogBox()

	ui.Body.Align()
	ui.Render(ui.Body)

	ui.Loop()
}
