package samplewav

import (
	"fmt"
	"image/color"
	"io"
)

type svgLine struct {
	x1, y1, x2, y2 int64
	rgba           *color.RGBA
	width          uint
}

type svgLinePrinter struct {
	buff                   string
	minx, maxx, miny, maxy int64
	lastwidth              int64
}

func (s *svgLinePrinter) add(line *svgLine) {
	s.buff += fmt.Sprintf(`<line x1="%d" y1="%d" x2="%d" y2="%d" style="stroke:rgba(%d,%d,%d,%d);stroke-width:%d"/>`,
		line.x1, line.y1, line.x2, line.y2, line.rgba.R, line.rgba.G, line.rgba.B, line.rgba.A, line.width)
	s.buff += "\n"
	s.maxx, s.minx = maxAndMin(line.x1, line.x2, s.maxx, s.minx)
	s.maxy, s.miny = maxAndMin(line.y1, line.y2, s.maxy, s.miny)
	s.lastwidth = int64(line.width)
}

func (s *svgLinePrinter) save(w io.Writer) (err error) {
	width := s.maxx - s.minx + s.lastwidth
	height := s.maxy - s.miny

	doc := fmt.Sprintf(`<?xml version="1.0"?>
	<svg width="%d" height="%d" 
		xmlns="http://www.w3.org/2000/svg"
		xmlns:xlink="http://www.w3.org/1999/xlink">
	<g transform="scale(1, -1) translate(0, -%d)">
		`, width, height, height)
	doc += "\n"
	doc += s.buff
	doc += "\n"
	doc += "</g>\n"
	doc += "</svg>"
	_, err = w.Write([]byte(doc))
	return
}

func maxAndMin(n ...int64) (max, min int64) {
	max = n[0]
	min = n[0]
	for _, i := range n {
		if i > max {
			max = i
		}
		if i < min {
			min = i
		}
	}
	return
}
