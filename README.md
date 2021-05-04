
# go-tsgrep

`go-tsgrep` is grep for Vim. This considers tabstop and modeline each file.  

There are following lines as `abc.txt`:  
```
// vi: set ts=4 :
This tabstop is 4.
--->This tabstop is 4.
--->--->This tabstop is 4.
```
And execute `go-tsgrep is abc.txt`:  
![](https://raw.githubusercontent.com/rbtnn/go-tsgrep/main/abc.jpg)


## Install

```
# go install github.com/rbtnn/go-tsgrep
```

## Usage

```
>go-tsgrep.exe -h
go-tsgrep.exe [OPTIONS] {pattern} {path}

Options:
  -color
        color matched text
  -debug
        debug mode
  -detect int
        bytes of filetype detection (default 100)
  -ignore-dir string
        ignore directories (default ".git,.gh,.hg,.svn,_svn,node_modules")
  -ignore-ext string
        ignore extensions (default ".exe,.dll,.obj,.mp3,mp4")
  -tabstop int
        default tabstop (default 8)

Example:
  >go-tsgrep -color -tabstop 4 set **/*

```

### Using Vim

If you use this in Vim, you put following code in your .vimrc:

```
if executable('go-tsgrep')
	set grepprg=go-tsgrep
	set grepformat=%f(%l\\,%v):%m
endif
```

We recommend to use `-tabstop` option if `&tabstop` in your Vim is not 8.

```
set grepprg=go-tsgrep\ -tabstop\ 4
```

## License
Distributed under MIT License. See LICENSE.

