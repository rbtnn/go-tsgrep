package main

import (
	"testing"
	"reflect"
)

func TestStringIndecies_1(t *testing.T) {
	expect := [][]int{
		[]int{ 1, 2 },
		[]int{ 5, 6 },
		[]int{ 9, 10 },
	}
	actual := StringIndecies("abc abc abc", "b")
	if !reflect.DeepEqual(actual, expect) {
		t.Errorf("actual: %q, expected: %q", actual, expect)
	}
}

func TestStringIndecies_2(t *testing.T) {
	expect := [][]int{
		[]int{ 1, 3 },
		[]int{ 5, 7 },
		[]int{ 9, 11 },
	}
	actual := StringIndecies("abc abc abc", "bc")
	if !reflect.DeepEqual(actual, expect) {
		t.Errorf("actual: %q, expected: %q", actual, expect)
	}
}

func TestCheckFileType_1(t *testing.T) {
	expect := BinaryFile
	actual := CheckFileType("./test/bin.txt", 100)
	if actual != expect {
		t.Errorf("actual: %q, expected: %q", actual, expect)
	}
}

func TestCheckFileType_2(t *testing.T) {
	expect := UTF8File
	actual := CheckFileType("./test/utf8.txt", 100)
	if actual != expect {
		t.Errorf("actual: %q, expected: %q", actual, expect)
	}
}

func TestCheckFileType_3(t *testing.T) {
	expect := SJISFile
	actual := CheckFileType("./test/sjis.txt", 100)
	if actual != expect {
		t.Errorf("actual: %q, expected: %q", actual, expect)
	}
}

func TestParseModeLine_1(t *testing.T) {
	expect := 0
	actual := ParseModeLine(0, GetModeLineRegex(), "", "")
	if actual != expect {
		t.Errorf("actual: %q, expected: %q", actual, expect)
	}
}

func TestParseModeLine_2(t *testing.T) {
	expect := 0
	actual := ParseModeLine(0, GetModeLineRegex(), "vim: ts=4", "")
	if actual != expect {
		t.Errorf("actual: %q, expected: %q", actual, expect)
	}
}

func TestParseModeLine_3(t *testing.T) {
	expect := 0
	actual := ParseModeLine(0, GetModeLineRegex(), "vim: set ts=4", "")
	if actual != expect {
		t.Errorf("actual: %q, expected: %q", actual, expect)
	}
}

func TestParseModeLine_4(t *testing.T) {
	expect := 4
	actual := ParseModeLine(0, GetModeLineRegex(), "vim: set ts=4:", "")
	if actual != expect {
		t.Errorf("actual: %q, expected: %q", actual, expect)
	}
}

func TestParseModeLine_5(t *testing.T) {
	expect := 8
	actual := ParseModeLine(0, GetModeLineRegex(), "vim: set ts=4:", "vim: set ts=8:")
	if actual != expect {
		t.Errorf("actual: %q, expected: %q", actual, expect)
	}
}

func TestExpandTabs_1(t *testing.T) {
	expect := "    $"
	actual := ExpandTabs("\t$", 4)
	if actual != expect {
		t.Errorf("actual: %q, expected: %q", actual, expect)
	}
}

func TestExpandTabs_2(t *testing.T) {
	expect := "X   $"
	actual := ExpandTabs("X\t$", 4)
	if actual != expect {
		t.Errorf("actual: %q, expected: %q", actual, expect)
	}
}

func TestExpandTabs_3(t *testing.T) {
	expect := "XX  $"
	actual := ExpandTabs("XX\t$", 4)
	if actual != expect {
		t.Errorf("actual: %q, expected: %q", actual, expect)
	}
}

func TestExpandTabs_4(t *testing.T) {
	expect := "XXX $"
	actual := ExpandTabs("XXX\t$", 4)
	if actual != expect {
		t.Errorf("actual: %q, expected: %q", actual, expect)
	}
}

func TestExpandTabs_5(t *testing.T) {
	expect := "XXXX    $"
	actual := ExpandTabs("XXXX\t$", 4)
	if actual != expect {
		t.Errorf("actual: %q, expected: %q", actual, expect)
	}
}

func TestExpandTabs_6(t *testing.T) {
	expect := "    X   $"
	actual := ExpandTabs("\tX\t$", 4)
	if actual != expect {
		t.Errorf("actual: %q, expected: %q", actual, expect)
	}
}

func TestGrepFile_1(t *testing.T) {
	expect := []GrepResult{
		GrepResult{"./test/utf8.txt", 1, 5, "あい", "う", "えお"},
		GrepResult{"./test/utf8.txt", 2, 9, "    あい", "う", "えお"},
		GrepResult{"./test/utf8.txt", 3, 13, "        あい", "う", "えお"},
	}
	actual := GrepFile("./test/utf8.txt", UTF8File, false, 4, GetModeLineRegex(), nil, "う")
	if !reflect.DeepEqual(actual, expect) {
		t.Errorf("actual: %q, expected: %q", actual, expect)
	}
}

func TestGrepFile_2(t *testing.T) {
	expect := []GrepResult{
		GrepResult{"./test/sjis.txt", 1, 5, "あい", "う", "えお"},
		GrepResult{"./test/sjis.txt", 2, 9, "    あい", "う", "えお"},
		GrepResult{"./test/sjis.txt", 3, 13, "        あい", "う", "えお"},
	}
	actual := GrepFile("./test/sjis.txt", SJISFile, false, 4, GetModeLineRegex(), nil, "う")
	if !reflect.DeepEqual(actual, expect) {
		t.Errorf("actual: %q, expected: %q", actual, expect)
	}
}
