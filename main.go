package main

import (
	"bufio"
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
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	global struct {
		stdOut            *bufio.Writer
		bytesOfDetection  *int
		tabstop           *int
		highlight         *bool
		ignoreDirectories *string
		ignoreExtensions  *string
		regexMode         *bool
		inputPattern      *string
		patternRegex      *regexp.Regexp
		modelineRegex     *regexp.Regexp
	}
)

type LineInfo struct {
	line string
	lnum int
}

type GrepResult struct {
	path   string
	lnum   int
	col    int
	head   string
	middle string
	tail   string
}

type FileType int
const (
	BinaryFile FileType = iota
	UTF8File
	SJISFile
	UnknownFile
)

func CheckFileType(path string, bytesOfDetection int) (FileType, error) {
	fp, err := os.Open(path)
	if err != nil {
		return UnknownFile, err
	}
	defer fp.Close()

	// check if is a binary file.
	buf := make([]byte, bytesOfDetection)
	n, err := fp.Read(buf)
	if err == nil {
		for i := 0; i < n; i++ {
			if buf[i] == 0 {
				return BinaryFile, nil
			}
		}
	}

	detector := chardet.NewTextDetector()
	if all, err := detector.DetectAll(buf); err == nil {
		isUTF8 := false
		isSJIS := false
		for i := 0; i < len(all); i++ {
			switch all[i].Charset {
			case "UTF-8":
				isUTF8 = true
			case "Shift_JIS":
				isSJIS = true
			}
		}
		if isUTF8 {
			return UTF8File, nil
		} else if isSJIS {
			return SJISFile, nil
		}
	}
	return UTF8File, nil
}

func GrepWrapper(wg *sync.WaitGroup, mu *sync.Mutex, path string) {
	defer wg.Done()

	xs := strings.Split(path, "/")
	for _, x := range xs {
		if strings.Contains(*global.ignoreDirectories, x) {
			return
		}
	}
	if strings.Contains(*global.ignoreExtensions, filepath.Ext(path)) {
		return
	}

	ft, err := CheckFileType(path, *global.bytesOfDetection)
	if err != nil {
		return
	}

	result := GrepFile(path, ft, *global.regexMode, *global.tabstop, global.modelineRegex, global.patternRegex, *global.inputPattern)

	mu.Lock()
	defer mu.Unlock()

	for _, x := range result {
		fmt.Fprint(global.stdOut, fmt.Sprintf("%s(%d,%d):", x.path, x.lnum, x.col))
		fmt.Fprint(global.stdOut, x.head)
		if *global.highlight {
			fmt.Fprint(global.stdOut, string("\x1b[32m"))
		}
		fmt.Fprint(global.stdOut, x.middle)
		if *global.highlight {
			fmt.Fprint(global.stdOut, string("\x1b[39m"))
		}
		fmt.Fprintln(global.stdOut, x.tail)
		global.stdOut.Flush()
	}
}

func GrepFile(path string, ft FileType, regexMode bool, tabstop int,
	modelineRegex *regexp.Regexp, patternRegex *regexp.Regexp, inputPattern string) []GrepResult {
	result := []GrepResult{}
	if ft != BinaryFile {
		fp, err := os.Open(path)
		if err != nil {
			panic(err)
		}
		defer fp.Close()

		reader := bufio.NewReaderSize(fp, 1024)
		lnum := 0
		lines := []LineInfo{}
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
				if ft == SJISFile {
					result, _, _ := transform.Bytes(japanese.ShiftJIS.NewDecoder(), line)
					line = result
				}
				if lnum == 1 {
					firstLine = line
				} else {
					lastLine = line
				}
				lines = append(lines, LineInfo{string(line), lnum})
				line = nil
			}
		}

		ts := ParseModeLine(tabstop, modelineRegex, string(firstLine), string(lastLine))
		for i := 0; i < len(lines); i++ {
			lines[i].line = ExpandTabs(lines[i].line, ts)
		}

		for _, x := range lines {
			var xs [][]int
			if regexMode {
				xs = patternRegex.FindAllIndex([]byte(x.line), -1)
			} else {
				xs = StringIndecies(x.line, inputPattern)
			}
			for i := 0; i < len(xs); i++ {
				head := string(x.line[:xs[i][0]])
				middle := string(x.line[xs[i][0]:xs[i][1]])
				tail := string(x.line[xs[i][1]:])
				col := runewidth.StringWidth(head) + 1
				result = append(result, GrepResult{ path, x.lnum, col, head, middle, tail })
			}
		}
	}
	return result
}

func GetModeLineRegex() *regexp.Regexp {
	return regexp.MustCompile("^(/|\\*|\\s)*vim?:\\s*set?\\s+.*\\bts=(\\d+).*:.*$")
}

func ParseModeLine(defaultTabStop int, re *regexp.Regexp, firstLine string, lastLine string, ) int {
	ts := defaultTabStop
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

func ExpandTabs(s string, ts int) string {
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

func StringIndecies(s string, substr string) [][]int {
	var xs [][]int
	offset := 0
	for {
		i := strings.Index(s[offset:], substr)
		if -1 == i {
			break
		}
		xs = append(xs, []int{offset + i, offset + i + len(substr)})
		offset += i + len(substr)
	}
	return xs
}

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "%s [OPTIONS] {pattern} {path}\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\n")
		fmt.Fprintf(os.Stderr, "Example:\n")
		fmt.Fprintf(os.Stderr, "  >go-tsgrep -color -tabstop 4 set **/*\n")
	}
	global.stdOut = bufio.NewWriter(colorable.NewColorableStdout())
	global.ignoreDirectories = flag.String("ignore-dir", ".git,.gh,.hg,.svn,_svn,node_modules", "ignore directories")
	global.ignoreExtensions = flag.String("ignore-ext", ".exe,.dll,.obj,.mp3,mp4", "ignore extensions")
	global.tabstop = flag.Int("tabstop", 8, "default tabstop")
	global.bytesOfDetection = flag.Int("detect", 100, "bytes of filetype detection")
	global.highlight = flag.Bool("color", false, "color matched text")
	global.regexMode = flag.Bool("regex", false, "deal with {pattern} as regex")

	flag.Parse()
	args := flag.Args()
	if 2 == len(args) {
		start := time.Now()
		global.inputPattern = &args[0]
		if *global.regexMode {
			global.patternRegex = regexp.MustCompile(*global.inputPattern)
		}
		global.modelineRegex = GetModeLineRegex()
		var wg sync.WaitGroup
		var mu sync.Mutex
		if matches, err := zglob.Glob(args[1]); err == nil {
			for i := 0; i < len(matches); i++ {
				if f, err := os.Stat(matches[i]); err == nil {
					if !os.IsNotExist(err) && !f.IsDir() {
						wg.Add(1)
						go GrepWrapper(&wg, &mu, matches[i])
					}
				}
			}
		}
		wg.Wait()
		elapsed := time.Since(start)
		fmt.Printf("elapsed time: %v\n", elapsed)
	}
}
