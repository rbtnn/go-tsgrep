package main

import (
	"bufio"
	"flag"
	"path/filepath"
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
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	global struct {
		detector           *chardet.Detector
		stdOut             *bufio.Writer
		bytesOfDetection   *int
		tabstop            *int
		highlight          *bool
		debug              *bool
		ignore_directories *string
		ignore_extensions  *string
		pattern_re         *regexp.Regexp
		modeline_re        *regexp.Regexp
	}
)

type LineInfo struct {
	line string
	lnum int
}

type FileType int

const (
	BinaryFile FileType = iota
	UTF8File
	SJISFile
)

func checkFileType(fp *os.File) FileType {
	// check if is a binary file.
	buf := make([]byte, *global.bytesOfDetection)
	n, err := fp.Read(buf)
	if err == nil {
		for i := 0; i < n; i++ {
			if buf[i] == 0 {
				return BinaryFile
			}
		}
	}
	fp.Seek(0, 0)
	if all, err := global.detector.DetectAll(buf); err == nil {
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
			return UTF8File
		} else if isSJIS {
			return SJISFile
		}
	}
	return UTF8File
}

func grep(wg *sync.WaitGroup, mu *sync.Mutex, path string) {
	defer wg.Done()

	xs := strings.Split(path, "/")
	for _, x := range xs {
		if strings.Contains(*global.ignore_directories, x) {
			return
		}
	}
	if strings.Contains(*global.ignore_extensions, filepath.Ext(path)) {
		return
	}

	fp, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer fp.Close()

	ft := checkFileType(fp)
	if *global.debug {
		if ft == BinaryFile {
			fmt.Println(path + "(0,0): a binary file.")
		} else if ft == UTF8File {
			fmt.Println(path + "(0,0): a UTF-8 file.")
		} else if ft == SJISFile {
			fmt.Println(path + "(0,0): a Shift-JIS file.")
		}
		return
	}
	if ft == BinaryFile {
		return
	}

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
			if ft == SJISFile {
				result, _, _ := transform.Bytes(japanese.ShiftJIS.NewDecoder(), line)
				line = result
			}
			if lnum == 1 {
				firstLine = line
			} else {
				lastLine = line
			}
			matches = append(matches, LineInfo{string(line), lnum})
			line = nil
		}
	}

	ts := parseModeLine(string(firstLine), string(lastLine))
	for i := 0; i < len(matches); i++ {
		matches[i].line = expandTabs(matches[i].line, ts)
	}

	mu.Lock()
	defer mu.Unlock()
	for _, x := range matches {
		xs := global.pattern_re.FindAllIndex([]byte(x.line), -1)
		for i := 0; i < len(xs); i++ {
			head := string(x.line[:xs[i][0]])
			middle := string(x.line[xs[i][0]:xs[i][1]])
			tail := string(x.line[xs[i][1]:])
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
	matches := global.modeline_re.FindAllStringSubmatch(firstLine, 1)
	if 0 < len(matches) {
		if sv, err := strconv.Atoi(matches[0][2]); err == nil {
			ts = sv
		}
	}
	matches = global.modeline_re.FindAllStringSubmatch(lastLine, 1)
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
	global.detector = chardet.NewTextDetector()
	global.stdOut = bufio.NewWriter(colorable.NewColorableStdout())
	global.ignore_directories = flag.String("ignore-dir", ".git,.gh,.hg,.svn,_svn,node_modules", "ignore directories")
	global.ignore_extensions = flag.String("ignore-ext", ".exe,.dll,.obj,.mp3,mp4", "ignore extensions")
	global.tabstop = flag.Int("tabstop", 8, "default tabstop")
	global.debug = flag.Bool("debug", false, "debug mode")
	global.bytesOfDetection = flag.Int("detect", 100, "bytes of filetype detection")
	global.highlight = flag.Bool("color", false, "color matched text")
	flag.Parse()
	args := flag.Args()
	if 2 == len(args) {
		start := time.Now()
		global.pattern_re = regexp.MustCompile(args[0])
		global.modeline_re = regexp.MustCompile("^(/|\\*|\\s)*vim?:\\s*set\\s+.*\\bts=(\\d+).*$")
		var wg sync.WaitGroup
		var mu sync.Mutex
		if matches, err := zglob.Glob(args[1]); err == nil {
			for i := 0; i < len(matches); i++ {
				if f, err := os.Stat(matches[i]); !os.IsNotExist(err) && !f.IsDir() {
					wg.Add(1)
					go grep(&wg, &mu, matches[i])
				}
			}
		}
		wg.Wait()
		elapsed := time.Since(start)
		if !*global.debug {
			fmt.Printf("elapsed time: %v\n", elapsed)
		}
	}
}
