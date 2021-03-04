package samplewav

import (
	"image/color"
	"io"
	"math"

	"github.com/go-audio/audio"
	"github.com/go-audio/wav"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
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
	ret["Duration"] = dur
	ret["BitsPerSample"] = uint(w.decoder.SampleBitDepth())
	return ret
}

//GenSVGWareform ...
func (w *Wareform) GenSVGWareform(path string) (err error) {

	plo, err := plot.New()
	if err != nil {
		return
	}

	linecount := 0
	//不使用gorountine 和 chan，尽量提高效率
	uper, lower := w.genSampleLine(10, func(line *plotter.XYs) {
		linecount++
		l, err := plotter.NewLine(line)
		if err != nil {
			return
		}
		l.LineStyle.Width = vg.Points(2)
		l.Color = &color.RGBA{R: 50, G: uint8(155), B: 240, A: 255}
		plo.Add(l)
	})

	plo.HideX()
	plo.HideY()
	plo.X.Min = 0
	plo.X.Max = float64(linecount)
	plo.Y.Min = lower * 1.1
	plo.Y.Max = uper * 1.1
	plo.BackgroundColor = color.White

	return plo.Save(vg.Points(float64(linecount)*4), 540, path)
}

func (w *Wareform) genSampleLine(lineNumPerSec int, drawFun func(line *plotter.XYs)) (uper, lower float64) {

	//缩放倍数
	scal := float64(w.decoder.SampleRate) / float64(lineNumPerSec)
	downtoss := int(math.Sqrt(scal / 4))
	if downtoss < 1 {
		downtoss = 1
	}
	merge := int(math.Sqrt(scal/4) * 4)
	if merge < 1 {
		merge = 1
	}
	space := float64(1)

	//mergef func
	mergeTick := 0
	isValid := false
	max := float64(0)
	min := float64(0)
	x := 0 - space
	_mergeFun := func(fd float64) {
		mergeTick++
		if !isValid {
			if merge == 1 {
				if fd > 0 {
					max = fd
					min = 0
				} else if fd < 0 {
					max = 0
					min = fd
				} else {
					max = 0
					min = 0
				}
			} else {
				max = fd
				min = fd
			}
			isValid = true
		}

		if max < fd {
			max = fd
		}
		if min > fd {
			min = fd
		}
		if mergeTick == merge {
			if max == min {
				max = min + 1
			}
			x += space
			drawFun(&plotter.XYs{{X: x, Y: min}, {X: x, Y: max}})

			if max > uper {
				uper = max
			}
			if min < lower {
				lower = min
			}
			mergeTick = 0
			isValid = false
		}
	}

	//readdata
	bufSampleCount := downtoss
	if bufSampleCount < 4096 {
		bufSampleCount = 4096
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
			mergeVal := abuf.Data[sampleStart]

			for j := 1; j < int(w.decoder.NumChans); j++ {
				//多个声道合一，获取绝对值最大的
				if math.Abs(float64(abuf.Data[sampleStart+j])) > math.Abs(float64(mergeVal)) {
					mergeVal = abuf.Data[sampleStart+j]
				}
			}
			_mergeFun(float64(mergeVal))
		}
	}
	return
}
