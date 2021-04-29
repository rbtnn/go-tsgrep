
# go-vgr

`go-vgr` is grep for Vim.

## Install

```
# go install github.com/rbtnn/go-vgr
```

## Usage

```
#go-vgr {pattern} {path}
```
`{pattern}` is a Regexp pattern of golang.  
`{path}` is a glob pattern of [zglob](https://github.com/mattn/go-zglob).  

### go-vgr's options
```
-t     : tabstop (default:8)
-c     : highlight matched text (default:false)
```

### vim settings
```
if executable('go-vgr')
	set grepprg=go-vgr
	set grepformat=%f(%l\\,%v):%m
endif
```

We recommend to use `-t` option if `&tabstop` in your Vim is not 8.
```
set grepprg=go-vgr\ -t\ 4
```

