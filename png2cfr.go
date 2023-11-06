package main

import (
	"fmt"
	"image"
	"image/color"
	_ "image/png"
	"os"
	"strings"
)

var COLORS = []color.RGBA{
	{0, 0, 0, 255},
	{0, 0, 255, 255},
	{0, 255, 0, 255},
	{0, 255, 255, 255},
	{255, 0, 0, 255},
	{255, 0, 255, 255},
	{255, 255, 0, 255},
	{255, 255, 255, 255},
}

type ColorMachine struct {
	state int
}

func NewColorMachine() *ColorMachine {
	return &ColorMachine{state: 7}
}

func distance(r1, g1, b1, r2, g2, b2 uint8) int {
	rDiff := int(r1) - int(r2)
	gDiff := int(g1) - int(g2)
	bDiff := int(b1) - int(b2)
	return rDiff*rDiff + gDiff*gDiff + bDiff*bDiff
}

func nearest(rgb color.RGBA) int {
	r, g, b := rgb.R, rgb.G, rgb.B

	minDist, nearestColor := -1, 0

	for i, c := range COLORS {
		dist := distance(r, g, b, c.R, c.G, c.B)
		if minDist == -1 || dist < minDist {
			minDist = dist
			nearestColor = i
		}
	}

	return nearestColor
}

func translate(len int) string {
	return strings.Repeat("C", len)
}

func getCommand(cm *ColorMachine, rgb color.RGBA) string {
	newState := nearest(rgb)

	var result string

	switch {
	case newState > cm.state:
		result = translate(newState - cm.state)
	case newState < cm.state:
		result = translate(len(COLORS) - cm.state + newState)
	}

	cm.state = newState
	return result
}

func checkBalanced(cmds string) bool {
	brackets := 0

	for _, cmd := range cmds {
		switch cmd {
		case '[':
			brackets++
		case ']':
			if brackets > 0 {
				brackets--
			} else {
				return false
			}
		}
	}
	return brackets == 0
}

func compress(cmds string) string {
begin:
	length := len(cmds)
	begin, half := 0, (length+1)/2

	for length > begin {
		start, end := begin, half-1

		for end > start {
			left := cmds[start : end+1]

			if checkBalanced(left) {
				if strings.Index(cmds[end+1:], left) == 0 {
					cmds = strings.ReplaceAll(cmds, cmds[start:2*(end+1)-start], "["+left+"]")
					goto begin
				}
			}

			end--
		}
		begin++
		half = begin + (length-begin+1)/2
	}

	return cmds
}

func scan(i_col, j_col []int, getcolor func(int, int) color.RGBA) string {
	cm := NewColorMachine()

	cmds := make([]string, 0)
	firstLetter := ""

	for _, i := range i_col {
		colors := make([]color.RGBA, 0)

		for _, j := range j_col {
			rgba := getcolor(i, j)

			if i&1 == 1 {
				colors = append([]color.RGBA{rgba}, colors...)
			} else {
				colors = append(colors, rgba)
			}
		}

		for _, c := range colors {
			cmds = append(cmds, getCommand(cm, c), "F")
			if firstLetter != "" {
				cmds = append(cmds, firstLetter)
				firstLetter = ""
			}
		}

		if i&1 == 1 {
			firstLetter = "[RRR]"
		} else {
			firstLetter = "RR"
		}

		cmds = append(cmds, firstLetter)
	}

	return strings.Join(cmds[:len(cmds)-1], "")
}

func fill_map(start, end int) []int {
	arr := make([]int, 0)

	if start > end {
		for i := start; i >= end; i -= 1 {
			arr = append(arr, i)
		}
	} else {
		for i := start; i < end; i += 1 {
			arr = append(arr, i)
		}
	}

	return arr
}

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage: png2cfr <pngfile>")
		return
	}

	imageFilePath := os.Args[1]

	file, err := os.Open(imageFilePath)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer file.Close()

	im, _, err := image.Decode(file)
	if err != nil {
		fmt.Println("Error decoding image:", err)
		return
	}

	bounds := im.Bounds()

	y_col := fill_map(bounds.Min.Y, bounds.Max.Y)
	x_col := fill_map(bounds.Min.X, bounds.Max.X)

	variant1 := compress("RR" + scan(y_col, x_col, func(y, x int) color.RGBA {
		return color.RGBAModel.Convert(im.At(x, y)).(color.RGBA)
	}))

	y_col = fill_map(bounds.Max.Y-1, bounds.Min.Y)

	variant2 := compress(scan(x_col, y_col, func(x, y int) color.RGBA {
		return color.RGBAModel.Convert(im.At(x, y)).(color.RGBA)
	}))

    if len(variant2) > len(variant1) {
        fmt.Println(variant1)
    } else {
        fmt.Println(variant2)
    }
}
