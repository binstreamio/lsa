package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/3d0c/gmf"
)

type streamInfo struct {
	ProcessStartTime time.Time

	StreamInfo string

	VideoFrames    uint64
	VideoKeyFrames uint64
	AudioFrames    uint64
	ReceivedSize   uint64

	AnalyzingLog string
}

func dumpStreamInfo(filename string, inputFormat *gmf.FmtCtx, videoStream, audioStream *gmf.Stream) string {
	streamInfo := fmt.Sprintf("Input from: %s, bitrate: %d Kbps", filename, inputFormat.BitRate()/1024)
	videoInfo := ""
	audioInfo := ""

	codecID2StrMap := map[int]string{
		27:    "h264",
		86018: "aac",
	}

	if videoStream != nil {
		codecPar := videoStream.GetCodecPar()

		codecString := codecID2StrMap[codecPar.GetCodecId()]
		if codecString == "" {
			codecString = fmt.Sprintf("%d", codecPar.GetCodecId())
		}

		videoInfo = fmt.Sprintf(" Stream Video: %s, %dX%d, %d Kbps, %f fps,",
			codecString,
			codecPar.GetWidth(), codecPar.GetHeight(),
			codecPar.GetBitRate()/1000,
			videoStream.GetRFrameRate().AVR().Av2qd(),
		)
	}

	if audioStream != nil {
		codecPar := audioStream.GetCodecPar()

		codecString := codecID2StrMap[codecPar.GetCodecId()]
		if codecString == "" {
			codecString = fmt.Sprintf("%d", codecPar.GetCodecId())
		}

		audioInfo = fmt.Sprintf(" Stream Audio: %s, %d Kbps",
			codecString,
			codecPar.GetBitRate()/1000,
		)
	}
	return strings.Join([]string{streamInfo, videoInfo, audioInfo}, "\n")
}

func appendLog(sInfo *streamInfo, log string) {
	sInfo.AnalyzingLog += "\n" + log

	logs := strings.Split(sInfo.AnalyzingLog, "\n")
	if len(logs) > 5 {
		sInfo.AnalyzingLog = strings.Join(logs[1:], "\n")
	}
}

func demuxing(source string, streamInfoChan chan *streamInfo) {
	gmf.LogSetLevel(gmf.AV_LOG_QUIET)

	sInfo := &streamInfo{
		ProcessStartTime: time.Now(),
	}
	defer func() { streamInfoChan <- sInfo }()

	streamInfoChan <- sInfo

	ictx, err := gmf.NewInputCtx(source)
	if err != nil {
		appendLog(sInfo, fmt.Sprintf("Open input error: %s", err))
		return
	}
	appendLog(sInfo, fmt.Sprintf("Open input done ..."))
	streamInfoChan <- sInfo

	srcVideoStream, err := ictx.GetBestStream(gmf.AVMEDIA_TYPE_VIDEO)
	if err != nil {
		appendLog(sInfo, fmt.Sprintf("Find video stream error: %s", err))
		return
	}
	appendLog(sInfo, fmt.Sprintf("Find video stream done ..."))

	srcAudioStream, err := ictx.GetBestStream(gmf.AVMEDIA_TYPE_AUDIO)
	if err != nil {
		appendLog(sInfo, fmt.Sprintf("Find audio stream error: %s", err))
		return
	}
	appendLog(sInfo, fmt.Sprintf("Find audio stream done ..."))
	streamInfoChan <- sInfo

	sInfo.StreamInfo = dumpStreamInfo(source, ictx, srcVideoStream, srcAudioStream)

	videoIndex := srcVideoStream.Index()
	audioIndex := srcAudioStream.Index()

	var (
		pkt       *gmf.Packet
		streamIdx int
	)

	for {
		pkt, err = ictx.GetNextPacket()
		if err != nil {
			appendLog(sInfo, fmt.Sprintf("Get next packet error: %s", err))
			break
		}
		streamIdx = pkt.StreamIndex()
		sInfo.ReceivedSize += uint64(pkt.Size())

		ist, err := ictx.GetStream(streamIdx)
		if err != nil {
			appendLog(sInfo, fmt.Sprintf("Get stream error: %s", err))
			break
		}
		frames, err := ist.CodecCtx().Decode(pkt)
		if err != nil {
			appendLog(sInfo, fmt.Sprintf("Decode frames error: %s", err))
			break
		}

		if streamIdx == videoIndex {
			sInfo.VideoFrames += uint64(len(frames))

			for i := range frames {
				if frames[i].KeyFrame() == 1 {
					sInfo.VideoKeyFrames++

					appendLog(sInfo, fmt.Sprintf("%s -- key frame: %f",
						time.Now().Format(time.Stamp),
						float32(frames[i].PktPts())*float32(ist.TimeBase().AVR().Av2qd())))
				}
			}

		} else if streamIdx == audioIndex {
			sInfo.AudioFrames += uint64(len(frames))
		}

		for i := range frames {
			frames[i].Free()
		}

		if pkt != nil {
			pkt.Free()
		}

		streamInfoChan <- sInfo
	}
}
