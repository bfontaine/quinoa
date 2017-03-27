package parser

import (
	"fmt"
	"math"
	"sort"
	"strconv"
)

const endSymbol rune = 1114112

/* The rule types inferred from the grammar are below. */
type pegRule uint8

const (
	ruleUnknown pegRule = iota
	ruleProgram
	ruleStatements
	ruleStatementSep
	ruleStatement
	ruleAssign
	ruleFuncCall
	ruleFuncArgs
	ruleExpression
	ruleNoBinopExpression
	ruleLitteral
	ruleVariable
	ruleBinOp
	ruleUnOp
	ruleOp
	ruleNumber
	ruleName
	ruleAlphaChar
	ruleDigit
	ruleAlphaNumericalChar
	ruleComment
	ruleSpaces
	ruleSpace
	ruleNewline
	ruleAction0
	ruleAction1
	ruleAction2
	ruleAction3
	ruleAction4
	ruleAction5
	ruleAction6
	rulePegText
)

var rul3s = [...]string{
	"Unknown",
	"Program",
	"Statements",
	"StatementSep",
	"Statement",
	"Assign",
	"FuncCall",
	"FuncArgs",
	"Expression",
	"NoBinopExpression",
	"Litteral",
	"Variable",
	"BinOp",
	"UnOp",
	"Op",
	"Number",
	"Name",
	"AlphaChar",
	"Digit",
	"AlphaNumericalChar",
	"Comment",
	"Spaces",
	"Space",
	"Newline",
	"Action0",
	"Action1",
	"Action2",
	"Action3",
	"Action4",
	"Action5",
	"Action6",
	"PegText",
}

type token32 struct {
	pegRule
	begin, end uint32
}

func (t *token32) String() string {
	return fmt.Sprintf("\x1B[34m%v\x1B[m %v %v", rul3s[t.pegRule], t.begin, t.end)
}

type node32 struct {
	token32
	up, next *node32
}

func (node *node32) print(pretty bool, buffer string) {
	var print func(node *node32, depth int)
	print = func(node *node32, depth int) {
		for node != nil {
			for c := 0; c < depth; c++ {
				fmt.Printf(" ")
			}
			rule := rul3s[node.pegRule]
			quote := strconv.Quote(string(([]rune(buffer)[node.begin:node.end])))
			if !pretty {
				fmt.Printf("%v %v\n", rule, quote)
			} else {
				fmt.Printf("\x1B[34m%v\x1B[m %v\n", rule, quote)
			}
			if node.up != nil {
				print(node.up, depth+1)
			}
			node = node.next
		}
	}
	print(node, 0)
}

func (node *node32) Print(buffer string) {
	node.print(false, buffer)
}

func (node *node32) PrettyPrint(buffer string) {
	node.print(true, buffer)
}

type tokens32 struct {
	tree []token32
}

func (t *tokens32) Trim(length uint32) {
	t.tree = t.tree[:length]
}

func (t *tokens32) Print() {
	for _, token := range t.tree {
		fmt.Println(token.String())
	}
}

func (t *tokens32) AST() *node32 {
	type element struct {
		node *node32
		down *element
	}
	tokens := t.Tokens()
	var stack *element
	for _, token := range tokens {
		if token.begin == token.end {
			continue
		}
		node := &node32{token32: token}
		for stack != nil && stack.node.begin >= token.begin && stack.node.end <= token.end {
			stack.node.next = node.up
			node.up = stack.node
			stack = stack.down
		}
		stack = &element{node: node, down: stack}
	}
	if stack != nil {
		return stack.node
	}
	return nil
}

func (t *tokens32) PrintSyntaxTree(buffer string) {
	t.AST().Print(buffer)
}

func (t *tokens32) PrettyPrintSyntaxTree(buffer string) {
	t.AST().PrettyPrint(buffer)
}

func (t *tokens32) Add(rule pegRule, begin, end, index uint32) {
	if tree := t.tree; int(index) >= len(tree) {
		expanded := make([]token32, 2*len(tree))
		copy(expanded, tree)
		t.tree = expanded
	}
	t.tree[index] = token32{
		pegRule: rule,
		begin:   begin,
		end:     end,
	}
}

func (t *tokens32) Tokens() []token32 {
	return t.tree
}

type Parser struct {
	stack *nodeStack

	Buffer string
	buffer []rune
	rules  [32]func() bool
	parse  func(rule ...int) error
	reset  func()
	Pretty bool
	tokens32
}

func (p *Parser) Parse(rule ...int) error {
	return p.parse(rule...)
}

func (p *Parser) Reset() {
	p.reset()
}

type textPosition struct {
	line, symbol int
}

type textPositionMap map[int]textPosition

func translatePositions(buffer []rune, positions []int) textPositionMap {
	length, translations, j, line, symbol := len(positions), make(textPositionMap, len(positions)), 0, 1, 0
	sort.Ints(positions)

search:
	for i, c := range buffer {
		if c == '\n' {
			line, symbol = line+1, 0
		} else {
			symbol++
		}
		if i == positions[j] {
			translations[positions[j]] = textPosition{line, symbol}
			for j++; j < length; j++ {
				if i != positions[j] {
					continue search
				}
			}
			break search
		}
	}

	return translations
}

type parseError struct {
	p   *Parser
	max token32
}

func (e *parseError) Error() string {
	tokens, error := []token32{e.max}, "\n"
	positions, p := make([]int, 2*len(tokens)), 0
	for _, token := range tokens {
		positions[p], p = int(token.begin), p+1
		positions[p], p = int(token.end), p+1
	}
	translations := translatePositions(e.p.buffer, positions)
	format := "parse error near %v (line %v symbol %v - line %v symbol %v):\n%v\n"
	if e.p.Pretty {
		format = "parse error near \x1B[34m%v\x1B[m (line %v symbol %v - line %v symbol %v):\n%v\n"
	}
	for _, token := range tokens {
		begin, end := int(token.begin), int(token.end)
		error += fmt.Sprintf(format,
			rul3s[token.pegRule],
			translations[begin].line, translations[begin].symbol,
			translations[end].line, translations[end].symbol,
			strconv.Quote(string(e.p.buffer[begin:end])))
	}

	return error
}

func (p *Parser) PrintSyntaxTree() {
	if p.Pretty {
		p.tokens32.PrettyPrintSyntaxTree(p.Buffer)
	} else {
		p.tokens32.PrintSyntaxTree(p.Buffer)
	}
}

func (p *Parser) Execute() {
	buffer, _buffer, text, begin, end := p.Buffer, p.buffer, "", 0, 0
	for _, token := range p.Tokens() {
		switch token.pegRule {

		case rulePegText:
			begin, end = int(token.begin), int(token.end)
			text = string(_buffer[begin:end])

		case ruleAction0:
			p.AddStatement()
		case ruleAction1:
			p.AddAssign()
		case ruleAction2:
			p.AddFnCall(text)
		case ruleAction3:
			p.AddLitteral(text)
		case ruleAction4:
			p.AddVariable(text)
		case ruleAction5:
			p.AddBinOp(text)
		case ruleAction6:
			p.AddUnOp(text)

		}
	}
	_, _, _, _, _ = buffer, _buffer, text, begin, end
}

func (p *Parser) Init() {
	var (
		max                  token32
		position, tokenIndex uint32
		buffer               []rune
	)
	p.reset = func() {
		max = token32{}
		position, tokenIndex = 0, 0

		p.buffer = []rune(p.Buffer)
		if len(p.buffer) == 0 || p.buffer[len(p.buffer)-1] != endSymbol {
			p.buffer = append(p.buffer, endSymbol)
		}
		buffer = p.buffer
	}
	p.reset()

	_rules := p.rules
	tree := tokens32{tree: make([]token32, math.MaxInt16)}
	p.parse = func(rule ...int) error {
		r := 1
		if len(rule) > 0 {
			r = rule[0]
		}
		matches := p.rules[r]()
		p.tokens32 = tree
		if matches {
			p.Trim(tokenIndex)
			return nil
		}
		return &parseError{p, max}
	}

	add := func(rule pegRule, begin uint32) {
		tree.Add(rule, begin, position, tokenIndex)
		tokenIndex++
		if begin != position && position > max.end {
			max = token32{rule, begin, position}
		}
	}

	matchDot := func() bool {
		if buffer[position] != endSymbol {
			position++
			return true
		}
		return false
	}

	/*matchChar := func(c byte) bool {
		if buffer[position] == c {
			position++
			return true
		}
		return false
	}*/

	/*matchRange := func(lower byte, upper byte) bool {
		if c := buffer[position]; c >= lower && c <= upper {
			position++
			return true
		}
		return false
	}*/

	_rules = [...]func() bool{
		nil,
		/* 0 Program <- <(Spaces Statements Spaces !.)> */
		func() bool {
			position0, tokenIndex0 := position, tokenIndex
			{
				position1 := position
				if !_rules[ruleSpaces]() {
					goto l0
				}
				if !_rules[ruleStatements]() {
					goto l0
				}
				if !_rules[ruleSpaces]() {
					goto l0
				}
				{
					position2, tokenIndex2 := position, tokenIndex
					if !matchDot() {
						goto l2
					}
					goto l0
				l2:
					position, tokenIndex = position2, tokenIndex2
				}
				add(ruleProgram, position1)
			}
			return true
		l0:
			position, tokenIndex = position0, tokenIndex0
			return false
		},
		/* 1 Statements <- <((Statement Spaces StatementSep Spaces)* Statement)> */
		func() bool {
			position3, tokenIndex3 := position, tokenIndex
			{
				position4 := position
			l5:
				{
					position6, tokenIndex6 := position, tokenIndex
					if !_rules[ruleStatement]() {
						goto l6
					}
					if !_rules[ruleSpaces]() {
						goto l6
					}
					if !_rules[ruleStatementSep]() {
						goto l6
					}
					if !_rules[ruleSpaces]() {
						goto l6
					}
					goto l5
				l6:
					position, tokenIndex = position6, tokenIndex6
				}
				if !_rules[ruleStatement]() {
					goto l3
				}
				add(ruleStatements, position4)
			}
			return true
		l3:
			position, tokenIndex = position3, tokenIndex3
			return false
		},
		/* 2 StatementSep <- <(';' / Newline / Comment)> */
		func() bool {
			position7, tokenIndex7 := position, tokenIndex
			{
				position8 := position
				{
					position9, tokenIndex9 := position, tokenIndex
					if buffer[position] != rune(';') {
						goto l10
					}
					position++
					goto l9
				l10:
					position, tokenIndex = position9, tokenIndex9
					if !_rules[ruleNewline]() {
						goto l11
					}
					goto l9
				l11:
					position, tokenIndex = position9, tokenIndex9
					if !_rules[ruleComment]() {
						goto l7
					}
				}
			l9:
				add(ruleStatementSep, position8)
			}
			return true
		l7:
			position, tokenIndex = position7, tokenIndex7
			return false
		},
		/* 3 Statement <- <(Assign / (FuncCall Action0))> */
		func() bool {
			position12, tokenIndex12 := position, tokenIndex
			{
				position13 := position
				{
					position14, tokenIndex14 := position, tokenIndex
					if !_rules[ruleAssign]() {
						goto l15
					}
					goto l14
				l15:
					position, tokenIndex = position14, tokenIndex14
					if !_rules[ruleFuncCall]() {
						goto l12
					}
					if !_rules[ruleAction0]() {
						goto l12
					}
				}
			l14:
				add(ruleStatement, position13)
			}
			return true
		l12:
			position, tokenIndex = position12, tokenIndex12
			return false
		},
		/* 4 Assign <- <(Variable Spaces '=' Spaces Expression Action1)> */
		func() bool {
			position16, tokenIndex16 := position, tokenIndex
			{
				position17 := position
				if !_rules[ruleVariable]() {
					goto l16
				}
				if !_rules[ruleSpaces]() {
					goto l16
				}
				if buffer[position] != rune('=') {
					goto l16
				}
				position++
				if !_rules[ruleSpaces]() {
					goto l16
				}
				if !_rules[ruleExpression]() {
					goto l16
				}
				if !_rules[ruleAction1]() {
					goto l16
				}
				add(ruleAssign, position17)
			}
			return true
		l16:
			position, tokenIndex = position16, tokenIndex16
			return false
		},
		/* 5 FuncCall <- <(Name Action2 Spaces '(' Spaces FuncArgs Spaces ')')> */
		func() bool {
			position18, tokenIndex18 := position, tokenIndex
			{
				position19 := position
				if !_rules[ruleName]() {
					goto l18
				}
				if !_rules[ruleAction2]() {
					goto l18
				}
				if !_rules[ruleSpaces]() {
					goto l18
				}
				if buffer[position] != rune('(') {
					goto l18
				}
				position++
				if !_rules[ruleSpaces]() {
					goto l18
				}
				if !_rules[ruleFuncArgs]() {
					goto l18
				}
				if !_rules[ruleSpaces]() {
					goto l18
				}
				if buffer[position] != rune(')') {
					goto l18
				}
				position++
				add(ruleFuncCall, position19)
			}
			return true
		l18:
			position, tokenIndex = position18, tokenIndex18
			return false
		},
		/* 6 FuncArgs <- <((Expression Spaces ',' Spaces)* Expression?)> */
		func() bool {
			{
				position21 := position
			l22:
				{
					position23, tokenIndex23 := position, tokenIndex
					if !_rules[ruleExpression]() {
						goto l23
					}
					if !_rules[ruleSpaces]() {
						goto l23
					}
					if buffer[position] != rune(',') {
						goto l23
					}
					position++
					if !_rules[ruleSpaces]() {
						goto l23
					}
					goto l22
				l23:
					position, tokenIndex = position23, tokenIndex23
				}
				{
					position24, tokenIndex24 := position, tokenIndex
					if !_rules[ruleExpression]() {
						goto l24
					}
					goto l25
				l24:
					position, tokenIndex = position24, tokenIndex24
				}
			l25:
				add(ruleFuncArgs, position21)
			}
			return true
		},
		/* 7 Expression <- <(NoBinopExpression / BinOp)> */
		func() bool {
			position26, tokenIndex26 := position, tokenIndex
			{
				position27 := position
				{
					position28, tokenIndex28 := position, tokenIndex
					if !_rules[ruleNoBinopExpression]() {
						goto l29
					}
					goto l28
				l29:
					position, tokenIndex = position28, tokenIndex28
					if !_rules[ruleBinOp]() {
						goto l26
					}
				}
			l28:
				add(ruleExpression, position27)
			}
			return true
		l26:
			position, tokenIndex = position26, tokenIndex26
			return false
		},
		/* 8 NoBinopExpression <- <(Litteral / Variable / UnOp / ('(' Spaces Expression Spaces ')'))> */
		func() bool {
			position30, tokenIndex30 := position, tokenIndex
			{
				position31 := position
				{
					position32, tokenIndex32 := position, tokenIndex
					if !_rules[ruleLitteral]() {
						goto l33
					}
					goto l32
				l33:
					position, tokenIndex = position32, tokenIndex32
					if !_rules[ruleVariable]() {
						goto l34
					}
					goto l32
				l34:
					position, tokenIndex = position32, tokenIndex32
					if !_rules[ruleUnOp]() {
						goto l35
					}
					goto l32
				l35:
					position, tokenIndex = position32, tokenIndex32
					if buffer[position] != rune('(') {
						goto l30
					}
					position++
					if !_rules[ruleSpaces]() {
						goto l30
					}
					if !_rules[ruleExpression]() {
						goto l30
					}
					if !_rules[ruleSpaces]() {
						goto l30
					}
					if buffer[position] != rune(')') {
						goto l30
					}
					position++
				}
			l32:
				add(ruleNoBinopExpression, position31)
			}
			return true
		l30:
			position, tokenIndex = position30, tokenIndex30
			return false
		},
		/* 9 Litteral <- <(Number Action3)> */
		func() bool {
			position36, tokenIndex36 := position, tokenIndex
			{
				position37 := position
				if !_rules[ruleNumber]() {
					goto l36
				}
				if !_rules[ruleAction3]() {
					goto l36
				}
				add(ruleLitteral, position37)
			}
			return true
		l36:
			position, tokenIndex = position36, tokenIndex36
			return false
		},
		/* 10 Variable <- <(Name Action4)> */
		func() bool {
			position38, tokenIndex38 := position, tokenIndex
			{
				position39 := position
				if !_rules[ruleName]() {
					goto l38
				}
				if !_rules[ruleAction4]() {
					goto l38
				}
				add(ruleVariable, position39)
			}
			return true
		l38:
			position, tokenIndex = position38, tokenIndex38
			return false
		},
		/* 11 BinOp <- <(NoBinopExpression Spaces Op Action5 Spaces Expression)> */
		func() bool {
			position40, tokenIndex40 := position, tokenIndex
			{
				position41 := position
				if !_rules[ruleNoBinopExpression]() {
					goto l40
				}
				if !_rules[ruleSpaces]() {
					goto l40
				}
				if !_rules[ruleOp]() {
					goto l40
				}
				if !_rules[ruleAction5]() {
					goto l40
				}
				if !_rules[ruleSpaces]() {
					goto l40
				}
				if !_rules[ruleExpression]() {
					goto l40
				}
				add(ruleBinOp, position41)
			}
			return true
		l40:
			position, tokenIndex = position40, tokenIndex40
			return false
		},
		/* 12 UnOp <- <(Op Spaces Expression Action6)> */
		func() bool {
			position42, tokenIndex42 := position, tokenIndex
			{
				position43 := position
				if !_rules[ruleOp]() {
					goto l42
				}
				if !_rules[ruleSpaces]() {
					goto l42
				}
				if !_rules[ruleExpression]() {
					goto l42
				}
				if !_rules[ruleAction6]() {
					goto l42
				}
				add(ruleUnOp, position43)
			}
			return true
		l42:
			position, tokenIndex = position42, tokenIndex42
			return false
		},
		/* 13 Op <- <'+'> */
		func() bool {
			position44, tokenIndex44 := position, tokenIndex
			{
				position45 := position
				if buffer[position] != rune('+') {
					goto l44
				}
				position++
				add(ruleOp, position45)
			}
			return true
		l44:
			position, tokenIndex = position44, tokenIndex44
			return false
		},
		/* 14 Number <- <<Digit+>> */
		func() bool {
			position46, tokenIndex46 := position, tokenIndex
			{
				position47 := position
				{
					position48 := position
					if !_rules[ruleDigit]() {
						goto l46
					}
				l49:
					{
						position50, tokenIndex50 := position, tokenIndex
						if !_rules[ruleDigit]() {
							goto l50
						}
						goto l49
					l50:
						position, tokenIndex = position50, tokenIndex50
					}
					add(rulePegText, position48)
				}
				add(ruleNumber, position47)
			}
			return true
		l46:
			position, tokenIndex = position46, tokenIndex46
			return false
		},
		/* 15 Name <- <<(AlphaChar AlphaNumericalChar*)>> */
		func() bool {
			position51, tokenIndex51 := position, tokenIndex
			{
				position52 := position
				{
					position53 := position
					if !_rules[ruleAlphaChar]() {
						goto l51
					}
				l54:
					{
						position55, tokenIndex55 := position, tokenIndex
						if !_rules[ruleAlphaNumericalChar]() {
							goto l55
						}
						goto l54
					l55:
						position, tokenIndex = position55, tokenIndex55
					}
					add(rulePegText, position53)
				}
				add(ruleName, position52)
			}
			return true
		l51:
			position, tokenIndex = position51, tokenIndex51
			return false
		},
		/* 16 AlphaChar <- <([a-z] / [A-Z] / '_')> */
		func() bool {
			position56, tokenIndex56 := position, tokenIndex
			{
				position57 := position
				{
					position58, tokenIndex58 := position, tokenIndex
					if c := buffer[position]; c < rune('a') || c > rune('z') {
						goto l59
					}
					position++
					goto l58
				l59:
					position, tokenIndex = position58, tokenIndex58
					if c := buffer[position]; c < rune('A') || c > rune('Z') {
						goto l60
					}
					position++
					goto l58
				l60:
					position, tokenIndex = position58, tokenIndex58
					if buffer[position] != rune('_') {
						goto l56
					}
					position++
				}
			l58:
				add(ruleAlphaChar, position57)
			}
			return true
		l56:
			position, tokenIndex = position56, tokenIndex56
			return false
		},
		/* 17 Digit <- <[0-9]> */
		func() bool {
			position61, tokenIndex61 := position, tokenIndex
			{
				position62 := position
				if c := buffer[position]; c < rune('0') || c > rune('9') {
					goto l61
				}
				position++
				add(ruleDigit, position62)
			}
			return true
		l61:
			position, tokenIndex = position61, tokenIndex61
			return false
		},
		/* 18 AlphaNumericalChar <- <(AlphaChar / Digit)> */
		func() bool {
			position63, tokenIndex63 := position, tokenIndex
			{
				position64 := position
				{
					position65, tokenIndex65 := position, tokenIndex
					if !_rules[ruleAlphaChar]() {
						goto l66
					}
					goto l65
				l66:
					position, tokenIndex = position65, tokenIndex65
					if !_rules[ruleDigit]() {
						goto l63
					}
				}
			l65:
				add(ruleAlphaNumericalChar, position64)
			}
			return true
		l63:
			position, tokenIndex = position63, tokenIndex63
			return false
		},
		/* 19 Comment <- <('#' (!Newline .)* Newline)> */
		func() bool {
			position67, tokenIndex67 := position, tokenIndex
			{
				position68 := position
				if buffer[position] != rune('#') {
					goto l67
				}
				position++
			l69:
				{
					position70, tokenIndex70 := position, tokenIndex
					{
						position71, tokenIndex71 := position, tokenIndex
						if !_rules[ruleNewline]() {
							goto l71
						}
						goto l70
					l71:
						position, tokenIndex = position71, tokenIndex71
					}
					if !matchDot() {
						goto l70
					}
					goto l69
				l70:
					position, tokenIndex = position70, tokenIndex70
				}
				if !_rules[ruleNewline]() {
					goto l67
				}
				add(ruleComment, position68)
			}
			return true
		l67:
			position, tokenIndex = position67, tokenIndex67
			return false
		},
		/* 20 Spaces <- <Space*> */
		func() bool {
			{
				position73 := position
			l74:
				{
					position75, tokenIndex75 := position, tokenIndex
					if !_rules[ruleSpace]() {
						goto l75
					}
					goto l74
				l75:
					position, tokenIndex = position75, tokenIndex75
				}
				add(ruleSpaces, position73)
			}
			return true
		},
		/* 21 Space <- <(' ' / '\t' / Newline / Comment)> */
		func() bool {
			position76, tokenIndex76 := position, tokenIndex
			{
				position77 := position
				{
					position78, tokenIndex78 := position, tokenIndex
					if buffer[position] != rune(' ') {
						goto l79
					}
					position++
					goto l78
				l79:
					position, tokenIndex = position78, tokenIndex78
					if buffer[position] != rune('\t') {
						goto l80
					}
					position++
					goto l78
				l80:
					position, tokenIndex = position78, tokenIndex78
					if !_rules[ruleNewline]() {
						goto l81
					}
					goto l78
				l81:
					position, tokenIndex = position78, tokenIndex78
					if !_rules[ruleComment]() {
						goto l76
					}
				}
			l78:
				add(ruleSpace, position77)
			}
			return true
		l76:
			position, tokenIndex = position76, tokenIndex76
			return false
		},
		/* 22 Newline <- <(('\r' '\n') / '\n' / '\r')> */
		func() bool {
			position82, tokenIndex82 := position, tokenIndex
			{
				position83 := position
				{
					position84, tokenIndex84 := position, tokenIndex
					if buffer[position] != rune('\r') {
						goto l85
					}
					position++
					if buffer[position] != rune('\n') {
						goto l85
					}
					position++
					goto l84
				l85:
					position, tokenIndex = position84, tokenIndex84
					if buffer[position] != rune('\n') {
						goto l86
					}
					position++
					goto l84
				l86:
					position, tokenIndex = position84, tokenIndex84
					if buffer[position] != rune('\r') {
						goto l82
					}
					position++
				}
			l84:
				add(ruleNewline, position83)
			}
			return true
		l82:
			position, tokenIndex = position82, tokenIndex82
			return false
		},
		/* 24 Action0 <- <{ p.AddStatement() }> */
		func() bool {
			{
				add(ruleAction0, position)
			}
			return true
		},
		/* 25 Action1 <- <{ p.AddAssign() }> */
		func() bool {
			{
				add(ruleAction1, position)
			}
			return true
		},
		/* 26 Action2 <- <{ p.AddFnCall(text) }> */
		func() bool {
			{
				add(ruleAction2, position)
			}
			return true
		},
		/* 27 Action3 <- <{ p.AddLitteral(text) }> */
		func() bool {
			{
				add(ruleAction3, position)
			}
			return true
		},
		/* 28 Action4 <- <{ p.AddVariable(text) }> */
		func() bool {
			{
				add(ruleAction4, position)
			}
			return true
		},
		/* 29 Action5 <- <{ p.AddBinOp(text) }> */
		func() bool {
			{
				add(ruleAction5, position)
			}
			return true
		},
		/* 30 Action6 <- <{ p.AddUnOp(text) }> */
		func() bool {
			{
				add(ruleAction6, position)
			}
			return true
		},
		nil,
	}
	p.rules = _rules
}
