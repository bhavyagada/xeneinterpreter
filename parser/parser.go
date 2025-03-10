package parser

import (
	"bytes"
	"errors"
	"fmt"
	parseError "github.com/bhavyagada/xeneinterpreter/errors"
	"github.com/bhavyagada/xeneinterpreter/token"
)

const (
	numProductions = 74
	numStates      = 774
	numSymbols     = 74
)

// Stack

type stack struct {
	state  []int
	attrib []Attrib
}

const iNITIAL_STACK_SIZE = 100

func newStack() *stack {
	return &stack{state: make([]int, 0, iNITIAL_STACK_SIZE),
		attrib: make([]Attrib, 0, iNITIAL_STACK_SIZE),
	}
}

func (this *stack) reset() {
	this.state = this.state[0:0]
	this.attrib = this.attrib[0:0]
}

func (this *stack) push(s int, a Attrib) {
	this.state = append(this.state, s)
	this.attrib = append(this.attrib, a)
}

func (this *stack) top() int {
	return this.state[len(this.state)-1]
}

func (this *stack) peek(pos int) int {
	return this.state[pos]
}

func (this *stack) topIndex() int {
	return len(this.state) - 1
}

func (this *stack) popN(items int) []Attrib {
	lo, hi := len(this.state)-items, len(this.state)

	attrib := this.attrib[lo:hi]

	this.state = this.state[:lo]
	this.attrib = this.attrib[:lo]

	return attrib
}

func (S *stack) String() string {
	w := new(bytes.Buffer)
	fmt.Fprintf(w, "stack:\n")
	for i, st := range S.state {
		fmt.Fprintf(w, "\t%d:%d , ", i, st)
		if S.attrib[i] == nil {
			fmt.Fprintf(w, "nil")
		} else {
			fmt.Fprintf(w, "%v", S.attrib[i])
		}
		w.WriteString("\n")
	}
	return w.String()
}

// Parser

type Parser struct {
	stack     *stack
	nextToken *token.Token
	pos       int
}

type Scanner interface {
	Scan() (tok *token.Token)
}

func NewParser() *Parser {
	p := &Parser{stack: newStack()}
	p.Reset()
	return p
}

func (P *Parser) Reset() {
	P.stack.reset()
	P.stack.push(0, nil)
}

func (P *Parser) Error(err error, scanner Scanner) (recovered bool, errorAttrib *parseError.Error) {
	errorAttrib = &parseError.Error{
		Err:            err,
		ErrorToken:     P.nextToken,
		ErrorSymbols:   P.popNonRecoveryStates(),
		ExpectedTokens: make([]string, 0, 8),
	}
	for t, action := range actionTab[P.stack.top()].actions {
		if action != nil {
			errorAttrib.ExpectedTokens = append(errorAttrib.ExpectedTokens, token.TokMap.Id(token.Type(t)))
		}
	}

	if action := actionTab[P.stack.top()].actions[token.TokMap.Type("error")]; action != nil {
		P.stack.push(int(action.(shift)), errorAttrib) // action can only be shift
	} else {
		return
	}

	if action := actionTab[P.stack.top()].actions[P.nextToken.Type]; action != nil {
		recovered = true
	}
	for !recovered && P.nextToken.Type != token.EOF {
		P.nextToken = scanner.Scan()
		if action := actionTab[P.stack.top()].actions[P.nextToken.Type]; action != nil {
			recovered = true
		}
	}

	return
}

func (P *Parser) popNonRecoveryStates() (removedAttribs []parseError.ErrorSymbol) {
	if rs, ok := P.firstRecoveryState(); ok {
		errorSymbols := P.stack.popN(int(P.stack.topIndex() - rs))
		removedAttribs = make([]parseError.ErrorSymbol, len(errorSymbols))
		for i, e := range errorSymbols {
			removedAttribs[i] = e
		}
	} else {
		removedAttribs = []parseError.ErrorSymbol{}
	}
	return
}

// recoveryState points to the highest state on the stack, which can recover
func (P *Parser) firstRecoveryState() (recoveryState int, canRecover bool) {
	recoveryState, canRecover = P.stack.topIndex(), actionTab[P.stack.top()].canRecover
	for recoveryState > 0 && !canRecover {
		recoveryState--
		canRecover = actionTab[P.stack.peek(recoveryState)].canRecover
	}
	return
}

func (P *Parser) newError(err error) error {
	w := new(bytes.Buffer)
	fmt.Fprintf(w, "Error in S%d: %s, %s", P.stack.top(), token.TokMap.TokenString(P.nextToken), P.nextToken.Pos.String())
	if err != nil {
		w.WriteString(err.Error())
	} else {
		w.WriteString(", expected one of: ")
		actRow := actionTab[P.stack.top()]
		for i, t := range actRow.actions {
			if t != nil {
				fmt.Fprintf(w, "%s ", token.TokMap.Id(token.Type(i)))
			}
		}
	}
	return errors.New(w.String())
}

func (this *Parser) Parse(scanner Scanner) (res interface{}, err error) {
	this.Reset()
	this.nextToken = scanner.Scan()
	for acc := false; !acc; {
		action := actionTab[this.stack.top()].actions[this.nextToken.Type]
		if action == nil {
			if recovered, errAttrib := this.Error(nil, scanner); !recovered {
				this.nextToken = errAttrib.ErrorToken
				return nil, this.newError(nil)
			}
			if action = actionTab[this.stack.top()].actions[this.nextToken.Type]; action == nil {
				panic("Error recovery led to invalid action")
			}
		}

		// fmt.Printf("S%d %s %s\n", this.stack.top(), token.TokMap.TokenString(this.nextToken), action.String())

		switch act := action.(type) {
		case accept:
			res = this.stack.popN(1)[0]
			acc = true
		case shift:
			this.stack.push(int(act), this.nextToken)
			this.nextToken = scanner.Scan()
		case reduce:
			prod := productionsTable[int(act)]
			attrib, err := prod.ReduceFunc(this.stack.popN(prod.NumSymbols))
			if err != nil {
				return nil, this.newError(err)
			} else {
				this.stack.push(gotoTab[this.stack.top()][prod.NTType], attrib)
			}
		default:
			panic("unknown action: " + action.String())
		}
	}
	return res, nil
}
