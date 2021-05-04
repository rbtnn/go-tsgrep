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
	"strconv"
	"sync"
	"time"
)

var (
	global struct {
		detector           *chardet.Detector
		stdOut             *bufio.Writer
		tabstop            *int
		highlight          *bool
		ignore_directories *string
		regex              *regexp.Regexp
	}
)

type LineInfo struct {
	line string
	lnum int
}

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

func grep(wg *sync.WaitGroup, mu *sync.Mutex, path string) {
	defer wg.Done()

	fp, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer fp.Close()

	reader := bufio.NewReaderSize(fp, 1024)
	lnum := 0
	matches := []LineInfo{}
	var line []byte
	var firstLine []byte
	var lastLine []byte
	for {
		tmp, isPrefix, err := reader.ReadLine()
		if err == io.EOF {
			break
		}
		if err != nil {
			panic(err)
		}
		line = append(line, tmp...)
		if !isPrefix {
			lnum++
			if text, err := convert(line); err == nil {
				if lnum == 1 {
					firstLine = text
				} else {
					lastLine = text
				}
				matches = append(matches, LineInfo{ string(text), lnum})
			}
			line = nil
		}
	}

	ts := parseModeLine(string(firstLine), string(lastLine))

	mu.Lock()
	defer mu.Unlock()
	for _, x := range matches {
		text := []byte(expandTabs(x.line, ts))
		xs := global.regex.FindAllIndex(text, -1)
		for i := 0; i < len(xs); i++ {
			head := string(text[:xs[i][0]])
			middle := string(text[xs[i][0]:xs[i][1]])
			tail := string(text[xs[i][1]:])
			col := runewidth.StringWidth(head) + 1
			fmt.Fprint(global.stdOut, fmt.Sprintf("%s(%d,%d):", path, x.lnum, col))
			fmt.Fprint(global.stdOut, head)
			if *global.highlight {
				fmt.Fprint(global.stdOut, string("\x1b[32m"))
			}
			fmt.Fprint(global.stdOut, middle)
			if *global.highlight {
				fmt.Fprint(global.stdOut, string("\x1b[39m"))
			}
			fmt.Fprintln(global.stdOut, tail)
			global.stdOut.Flush()
		}
	}
}

func parseModeLine(firstLine string, lastLine string) int {
	ts := *global.tabstop
	re := regexp.MustCompile("^(/|\\*|\\s)*vim?:\\s*set\\s+.*\\bts=(\\d+).*$")
	matches := re.FindAllStringSubmatch(firstLine, 1)
	if 0 < len(matches) {
		if sv, err := strconv.Atoi(matches[0][2]); err == nil {
			ts = sv
		}
	}
	matches = re.FindAllStringSubmatch(lastLine, 1)
	if 0 < len(matches) {
		if sv, err := strconv.Atoi(matches[0][2]); err == nil {
			ts = sv
		}
	}
	return ts
}

func expandTabs(s string, ts int) string {
	text := ""
	col := 0
	for _, ch := range s {
		if ch == 9 {
			text += strings.Repeat(" ", ts-col%ts)
			col += ts - col%ts
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
	if err == nil {
		for i := 0; i < n; i++ {
			if buf[i] == 0 {
				return true
			}
		}
	}
	return false
}

func main() {
	start := time.Now()
	global.detector = chardet.NewTextDetector()
	global.stdOut = bufio.NewWriter(colorable.NewColorableStdout())
	global.ignore_directories = flag.String("i", ".git,.gh,.hg,.svn,_svn", "ignore directories")
	global.tabstop = flag.Int("t", 8, "default tabstop")
	global.highlight = flag.Bool("c", false, "highlight matched text")
	flag.Parse()
	args := flag.Args()
	if 2 == len(args) {
		global.regex = regexp.MustCompile(args[0])
		var wg sync.WaitGroup
		var mu sync.Mutex
		if matches, err := zglob.Glob(args[1]); err == nil {
			for i := 0; i < len(matches); i++ {
				if f, err := os.Stat(matches[i]); !os.IsNotExist(err) && !f.IsDir() {
					if !isBinary(matches[i]) {
						xs := strings.Split(matches[i], "/")
						skip := false
						for _, x := range xs {
							if strings.Contains(*global.ignore_directories, x) {
								skip = true
								break
							}
						}
						if !skip {
							wg.Add(1)
							go grep(&wg, &mu, matches[i])
						}
					}
				}
			}
		}
		wg.Wait()
	}
	elapsed := time.Since(start)
	fmt.Printf("elapsed time: %v\n", elapsed)
}
