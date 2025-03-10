package main

import (
	"github.com/bhavyagada/xeneinterpreter/token"
	"github.com/nsf/termbox-go"
)

const hlKeyword = termbox.AttrBold | termbox.ColorMagenta
const hlVariable = termbox.ColorCyan
const hlConst = termbox.ColorGreen
const hlString = termbox.ColorYellow

var highlightMap map[string]termbox.Attribute = map[string]termbox.Attribute{
	"function": hlKeyword,
	"return":   hlKeyword,
	"if":       hlKeyword,
	"else":     hlKeyword,
	"while":    hlKeyword,
	"foreach":  hlKeyword,
	"in":       hlKeyword,
	"->":       hlKeyword,
	":":        hlKeyword,
	//"var":          hlVariable,
	//"input":        hlVariable,
	"cust_fn_name": hlVariable,
	"fn_name":      hlVariable,
	"true":         hlConst,
	"false":        hlConst,
	"int":          hlConst,
	"string":       hlString,
}

func getForeground(t *token.Token) termbox.Attribute {
	tokId := token.TokMap.Id(t.Type)
	att, ok := highlightMap[tokId]
	if ok {
		return att
	}
	return termbox.ColorWhite
}
