package main

import (
	"fmt"
	"image/color"
	"image/png"
	"log"
	"math"
	"os"
	"os/exec"
	"runtime/debug"
	"strconv"
	"time"
	"math/rand"
	"github.com/nfnt/resize"
)

func getRGB(m color.Model, c color.Color) [3]int {
	if m == color.RGBAModel {
		return [3]int{int(c.(color.RGBA).R), int(c.(color.RGBA).G), int(c.(color.RGBA).B)}
	} else if m == color.RGBA64Model {
		return [3]int{int(c.(color.RGBA64).R), int(c.(color.RGBA64).G), int(c.(color.RGBA64).B)}
	} else if m == color.NRGBAModel {
		return [3]int{int(c.(color.NRGBA).R), int(c.(color.NRGBA).G), int(c.(color.NRGBA).B)}
	} else if m == color.NRGBA64Model {
		return [3]int{int(c.(color.NRGBA64).R), int(c.(color.NRGBA64).G), int(c.(color.NRGBA64).B)}
	}
	return [3]int{0, 0, 0}
}

func colorSimilar(a, b [3]int, distance float64) bool {
	return (math.Abs(float64(a[0]-b[0])) < distance) && (math.Abs(float64(a[1]-b[1])) < distance) && (math.Abs(float64(a[2]-b[2])) < distance)
}

const charset = "abcdefghijklmnopqrstuvwxyz" +
  "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

var seededRand *rand.Rand = rand.New(
  rand.NewSource(time.Now().UnixNano()))

func StringWithCharset(length int, charset string) string {
  b := make([]byte, length)
  for i := range b {
    b[i] = charset[seededRand.Intn(len(charset))]
  }
  return string(b)
}

func String(length int) string {
  return StringWithCharset(length, charset)
}


func main() {
	defer func() {
		if e := recover(); e != nil {
			log.Printf("%s: %s", e, debug.Stack())
			fmt.Print("程序已崩溃，请保存日志后按任意键退出\n")
			var c string
			fmt.Scanln(&c)
		}
	}()

	var ratio float64
	fmt.Print("请输入跳跃系数(推荐值 2.04):")
	_, err := fmt.Scanln(&ratio)
	if err != nil {
		log.Fatal(err)
	}

	for {
		screenshotPath := fmt.Sprintf("/sdcard/beam/jump_%s.png", String(4))
		fmt.Println("screenshotPath", screenshotPath)

		_, err := exec.Command("adb", "shell", "screencap", "-p", screenshotPath).Output()
		if err != nil {
			panic(fmt.Sprintf("ADB 截图失败，请手动执行 \"adb shell screencap -p %s\" 看是否有报错", screenshotPath))
		}
		_, err = exec.Command("adb", "pull", screenshotPath, "./jump.png").Output()
		if err != nil {
			panic(fmt.Sprintf("ADB 拉取截图失败，请手动执行 \"adb pull %s\" 看是否有报错", screenshotPath))
		}

		args := []string{"shell", "rm", screenshotPath}
		_, err = exec.Command("adb", args...).Output()
		if err != nil {
			panic(fmt.Sprintf("ADB 清除截图失败，请手动执行 \"adb shell rm  %s\" 看是否有报错", screenshotPath))
		}

		infile, err := os.Open("jump.png")
		if err != nil {
			panic(err)
		}

		src, err := png.Decode(infile)
		if err != nil {
			panic(err)
		}
		src = resize.Resize(720, 0, src, resize.Lanczos3)
		bounds := src.Bounds()
		w, h := bounds.Max.X, bounds.Max.Y

		jumpCubeColor := [3]int{54, 52, 92}
		points := [][]int{}
		for y := 0; y < h; y++ {
			line := 0
			for x := 0; x < w; x++ {
				c := src.At(x, y)
				if colorSimilar(getRGB(src.ColorModel(), c), jumpCubeColor, 20) {
					line++
				} else {
					if y > 200 && x-line > 10 && line > 30 {
						points = append(points, []int{x - line/2, y, line})
					}
					line = 0
				}
			}
		}
		jumpCube := []int{0, 0, 0}
		for _, point := range points {
			if point[2] > jumpCube[2] {
				jumpCube = point
			}
		}
		jumpCube = []int{jumpCube[0], jumpCube[1]}

		possible := [][]int{}
		for y := 0; y < h; y++ {
			line := 0
			bgColor := getRGB(src.ColorModel(), src.At(w-10, y))
			for x := 0; x < w; x++ {
				c := src.At(x, y)
				if !colorSimilar(getRGB(src.ColorModel(), c), bgColor, 10) {
					line++
				} else {
					if y > 200 && x-line > 10 && line > 35 && ((x-line/2) < (jumpCube[0]-20) || (x-line/2) > (jumpCube[0]+20)) {
						possible = append(possible, []int{x - line/2, y, line, x})
					}
					line = 0
				}
			}
		}
		target := possible[0]
		for _, point := range possible {
			if point[3] > target[3] && point[1]-target[1] <= 1 {
				target = point
			}
		}
		target = []int{target[0], target[1]}

		ms := int(math.Pow(math.Pow(float64(jumpCube[0]-target[0]), 2)+math.Pow(float64(jumpCube[1]-target[1]), 2), 0.5) * ratio)
		log.Printf("from:%v to:%v press:%vms", jumpCube, target, ms)

		_, err = exec.Command("adb", "shell", "input", "swipe", "320", "410", "320", "410", strconv.Itoa(ms)).Output()
		if err != nil {
			panic("ADB 执行失败，请手动执行 \"adb shell input swipe 320 410 320 410 300\" 看是否有报错")
		}

		infile.Close()
		time.Sleep(time.Millisecond * time.Duration(seededRand.Intn(1000) + 1500))
	}
}
