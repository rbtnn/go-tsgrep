
# go-vgr

`go-vgr` is grep for Vim.

## Install

```
# go install github.com/rbtnn/go-vgr
```

## Usage

__go-vgr's options__
```
-t     : tabstop (default:8)
-c     : highlight matched text (default:false)
```

__vim settings__
```
if executable('go-vgr')
	set grepprg=go-vgr
	set grepformat=%f(%l\\,%v):%m
endif
```

We recommend to use `-t` option if `&tabstop` in your Vim is not 8.

__example__
```
set grepprg=go-vgr\ -t\ 4
```

