
# go-vgr

`go-vgr` is grep for Vim. This considers tabstop and modeline each file.  

There are following lines as `abc.txt`:  
```
This tabstop is 4.
--->This tabstop is 4.
--->--->This tabstop is 4.
// vim: set ts=4 :
```
And execute `go-vgr is abc.txt`:  
![](https://raw.githubusercontent.com/rbtnn/go-vgr/main/abc.jpg)


## Install

```
# go install github.com/rbtnn/go-vgr
```

## Usage

```
>go-vgr.exe -h
go-vgr.exe [OPTIONS] {pattern} {path}

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
  >go-vgr -color -tabstop 4 set **/*

```

### Using Vim

If you use this in Vim, you put following code in your .vimrc:

```
if executable('go-vgr')
	set grepprg=go-vgr
	set grepformat=%f(%l\\,%v):%m
endif
```

We recommend to use `-tabstop` option if `&tabstop` in your Vim is not 8.

```
set grepprg=go-vgr\ -tabstop\ 4
```

## License
Distributed under MIT License. See LICENSE.

