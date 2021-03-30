package samplewav

import (
	"image/color"
	"io"
	"math"
	"time"
	"fmt"
	"github.com/go-audio/audio"
	"github.com/go-audio/wav"
	tools "github.com/yinyajiang/go-ytools/utils"
)

// WaveFormat wave format tag
const (
	WaveFormatPCM        = 0x0001
	WaveFormatIEEEFloat  = 0x0003
	WaveFormatALaw       = 0x0006
	WaveFormatMULaw      = 0x0007
	WaveFormatExtensible = 0xFFFE
)

//Wareform ...
type Wareform struct {
	decoder *wav.Decoder
}

//NewWareform ...
func NewWareform(r io.ReadSeeker) *Wareform {
	decoder := wav.NewDecoder(r)
	decoder.ReadInfo()
	return &Wareform{
		decoder: decoder,
	}
}

//AudioInfo ...
func (w *Wareform) AudioInfo() map[string]interface{} {
	ret := make(map[string]interface{}, 5)
	ret["AudioFormat"] = uint(w.decoder.WavAudioFormat)
	ret["NumChannels"] = uint(w.decoder.NumChans)
	ret["SampleRate"] = uint(w.decoder.SampleRate)
	dur, _ := w.decoder.Duration()
	ret["Duration"] = int(dur / time.Millisecond)
	ret["BitsPerSample"] = uint(w.decoder.SampleBitDepth())
	return ret
}

//GenWareform .png or .svg
func (w *Wareform) GenWareform(path string) (err error) {

	duration, err := w.decoder.Duration()
	if err != nil {
		return
	}
	dur := int(duration / time.Second)
	if dur < 1 {
		dur = 1
	}
	di := int(math.Log2(float64(dur)))
	if di < 1 {
		di = 1
	}
	linePerSec := 40 / di

	space := uint(linePerSec)
	if space < 1 {
		space = 2
	}
	lineWidth := space / 2
	if space < 1 {
		space = 1
	}

	startColor := [4]float64{172, 185, 255, 255}
	endColor := [4]float64{109, 129, 255, 255}
	step := [4]float64{
		(endColor[0] - startColor[0]) / float64(linePerSec) / float64(dur), //R
		(endColor[1] - startColor[1]) / float64(linePerSec) / float64(dur), //G
		(endColor[2] - startColor[2]) / float64(linePerSec) / float64(dur), //B
		(endColor[3] - startColor[3]) / float64(linePerSec) / float64(dur), //A
	}

	linecount := 0

	svg := &svgLinePrinter{}
	//不使用gorountine 和 chan，尽量提高效率
	w.genSampleLine(linePerSec, space, lineWidth, func(l *svgLine) {
		linecount++

		R := startColor[0] + step[0]*float64(linecount)
		G := startColor[1] + step[1]*float64(linecount)
		B := startColor[2] + step[2]*float64(linecount)
		A := startColor[3] + step[3]*float64(linecount)

		l.rgba = &color.RGBA{R: uint8(R), G: uint8(G), B: uint8(B), A: uint8(A)}
		svg.add(l)
	})
	fmt.Println("linecount:", linecount)
	f, err := tools.CreateFile(path)
	if err != nil {
		return
	}
	defer f.Close()
	return svg.save(f)
}

func (w *Wareform) genSampleLine(lineNumPerSec int, space, lineWidth uint, drawFun func(line *svgLine)) {

	//缩放倍数
	downtoss := 1
	if uint32(lineNumPerSec) <= w.decoder.SampleRate {
		downtoss = int(float64(w.decoder.SampleRate) / float64(lineNumPerSec))
	}

	//mergef func
	x := 0 - space - lineWidth
	_draw := func(fd float64) {
		if fd == 0 {
			fd = 1.1
		}
		fd = math.Log2(fd)
		x += space + lineWidth
		drawFun(&svgLine{
			x1:    int64(x),
			y1:    int64(0),
			x2:    int64(x),
			y2:    int64(fd) * 10,
			width: lineWidth,
		})
	}

	//readdata
	bufSampleCount := downtoss
	if bufSampleCount < 4096 {
		bufSampleCount = 4096 / downtoss * downtoss
	}
	abuf := &audio.IntBuffer{Data: make([]int, bufSampleCount*int(w.decoder.NumChans))}
	for {
		n, err := w.decoder.PCMBuffer(abuf)
		n /= int(w.decoder.NumChans)
		if n == 0 || err != nil {
			break
		}

		//过滤信息
		for i := 0; i < n; i += downtoss {
			sampleStart := i * int(w.decoder.NumChans)
			mergeVal := math.Abs(float64(abuf.Data[sampleStart]))

			for j := 1; j < int(w.decoder.NumChans); j++ {
				//多个声道合-
				mergeVal += math.Abs(float64(abuf.Data[sampleStart+j]))
			}

			mergeVal /= float64(w.decoder.NumChans)
			_draw(float64(mergeVal))
		}
	}
	return
}
