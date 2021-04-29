package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"github.com/mattn/go-colorable"
	"github.com/mattn/go-runewidth"
	"github.com/mattn/go-zglob"
	"github.com/saintfish/chardet"
	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
	"io"
	"os"
	"regexp"
	"strings"
)

var (
	global struct {
		detector *chardet.Detector
		stdOut   *bufio.Writer
		tabstop  *int
		color    *bool
		regex    *regexp.Regexp
	}
)

func convert(line []byte) ([]byte, error) {
	if all, err := global.detector.DetectAll(line); err == nil {
		for i := 0; i < len(all); i++ {
			switch all[i].Charset {
			case "UTF-8":
				return line, nil
			case "EUC-JP":
				result, _, _ := transform.Bytes(japanese.EUCJP.NewDecoder(), line)
				return result, nil
			case "Shift_JIS":
				result, _, _ := transform.Bytes(japanese.ShiftJIS.NewDecoder(), line)
				return result, nil
			}
		}
	}
	return nil, errors.New("Charset not detected.")
}

func grep(path string) {
	fp, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer fp.Close()

	reader := bufio.NewReaderSize(fp, 1024)
	lnum := 0
	for {
		line, _, err := reader.ReadLine()
		if err == io.EOF {
			break
		}
		if err != nil {
			panic(err)
		}
		lnum++
		if text, err := convert(line); err == nil {
			text = []byte(expandTabs(string(text)))
			xs := global.regex.FindAllIndex(text, -1)
			for i := 0; i < len(xs); i++ {
				head := string(text[:xs[i][0]])
				middle := string(text[xs[i][0]:xs[i][1]])
				tail := string(text[xs[i][1]:])
				col := runewidth.StringWidth(head) + 1
				fmt.Fprint(global.stdOut, fmt.Sprintf("%s(%d,%d):", path, lnum, col))
				fmt.Fprint(global.stdOut, head)
				if *global.color {
					fmt.Fprint(global.stdOut, string("\x1b[32m"))
				}
				fmt.Fprint(global.stdOut, middle)
				if *global.color {
					fmt.Fprint(global.stdOut, string("\x1b[39m"))
				}
				fmt.Fprintln(global.stdOut, tail)
				global.stdOut.Flush()
			}
		}
	}
}

func contains(xs []string, x string) bool {
	for i := 0; i < len(xs); i++ {
		if xs[i] == x {
			return true
		}
	}
	return false
}

func expandTabs(s string) string {
	text := ""
	col := 0
	for _, ch := range s {
		if ch == 9 {
			text += strings.Repeat(" ", *global.tabstop-col%*global.tabstop)
			col += *global.tabstop - col%*global.tabstop
		} else {
			text += string(ch)
			col += runewidth.StringWidth(string(ch))
		}
	}
	return text
}

func isBinary(path string) bool {
	fp, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer fp.Close()

	buf := make([]byte, 100)
	n, err := fp.Read(buf)
	if err != nil {
		panic(err)
	}
	for i := 0; i < n; i++ {
		if buf[i] == 0 {
			return true
		}
	}
	return false
}

func main() {
	global.detector = chardet.NewTextDetector()
	global.stdOut = bufio.NewWriter(colorable.NewColorableStdout())
	global.tabstop = flag.Int("t", 8, "tabstop")
	global.color = flag.Bool("c", false, "color")
	flag.Parse()
	args := flag.Args()
	if 2 == len(args) {
		global.regex = regexp.MustCompile(args[0])
		if matches, err := zglob.Glob(args[1]); err == nil {
			for i := 0; i < len(matches); i++ {
				if f, err := os.Stat(matches[i]); !os.IsNotExist(err) && !f.IsDir() {
					if isBinary(matches[i]) {
						// fmt.Fprintln(global.stdOut, fmt.Sprintf("'%s' is a binary file.", matches[i]))
						// global.stdOut.Flush()
					} else {
						grep(matches[i])
					}
				}
			}
		}
	}
}
