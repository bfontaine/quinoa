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
	ruleNoOpExpression
	ruleLitteral
	ruleVariable
	ruleBinop
	ruleUnop
	ruleOp
	ruleNumber
	ruleName
	ruleAlphaChar
	ruleDigit
	ruleAlphaNumericalChar
	ruleComment
	ruleSpaces
	ruleSpace
	ruleSimpleSpaces
	ruleSimpleSpace
	ruleNewline
	ruleAction0
	ruleAction1
	ruleAction2
	ruleAction3
	ruleAction4
	ruleAction5
	ruleAction6
	ruleAction7
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
	"NoOpExpression",
	"Litteral",
	"Variable",
	"Binop",
	"Unop",
	"Op",
	"Number",
	"Name",
	"AlphaChar",
	"Digit",
	"AlphaNumericalChar",
	"Comment",
	"Spaces",
	"Space",
	"SimpleSpaces",
	"SimpleSpace",
	"Newline",
	"Action0",
	"Action1",
	"Action2",
	"Action3",
	"Action4",
	"Action5",
	"Action6",
	"Action7",
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
	rules  [36]func() bool
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
			p.AddFuncName(text)
		case ruleAction3:
			p.AddFuncCall()
		case ruleAction4:
			p.AddLitteral(text)
		case ruleAction5:
			p.AddVariable(text)
		case ruleAction6:
			p.AddBinop(text)
		case ruleAction7:
			p.AddUnop(text)

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
		/* 1 Statements <- <((Statement SimpleSpaces StatementSep SimpleSpaces)* Statement)> */
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
					if !_rules[ruleSimpleSpaces]() {
						goto l6
					}
					if !_rules[ruleStatementSep]() {
						goto l6
					}
					if !_rules[ruleSimpleSpaces]() {
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
		/* 2 StatementSep <- <(Newline / Comment / ';')+> */
		func() bool {
			position7, tokenIndex7 := position, tokenIndex
			{
				position8 := position
				{
					position11, tokenIndex11 := position, tokenIndex
					if !_rules[ruleNewline]() {
						goto l12
					}
					goto l11
				l12:
					position, tokenIndex = position11, tokenIndex11
					if !_rules[ruleComment]() {
						goto l13
					}
					goto l11
				l13:
					position, tokenIndex = position11, tokenIndex11
					if buffer[position] != rune(';') {
						goto l7
					}
					position++
				}
			l11:
			l9:
				{
					position10, tokenIndex10 := position, tokenIndex
					{
						position14, tokenIndex14 := position, tokenIndex
						if !_rules[ruleNewline]() {
							goto l15
						}
						goto l14
					l15:
						position, tokenIndex = position14, tokenIndex14
						if !_rules[ruleComment]() {
							goto l16
						}
						goto l14
					l16:
						position, tokenIndex = position14, tokenIndex14
						if buffer[position] != rune(';') {
							goto l10
						}
						position++
					}
				l14:
					goto l9
				l10:
					position, tokenIndex = position10, tokenIndex10
				}
				add(ruleStatementSep, position8)
			}
			return true
		l7:
			position, tokenIndex = position7, tokenIndex7
			return false
		},
		/* 3 Statement <- <(Assign / (FuncCall Action0))> */
		func() bool {
			position17, tokenIndex17 := position, tokenIndex
			{
				position18 := position
				{
					position19, tokenIndex19 := position, tokenIndex
					if !_rules[ruleAssign]() {
						goto l20
					}
					goto l19
				l20:
					position, tokenIndex = position19, tokenIndex19
					if !_rules[ruleFuncCall]() {
						goto l17
					}
					if !_rules[ruleAction0]() {
						goto l17
					}
				}
			l19:
				add(ruleStatement, position18)
			}
			return true
		l17:
			position, tokenIndex = position17, tokenIndex17
			return false
		},
		/* 4 Assign <- <(Variable SimpleSpaces '=' Spaces Expression Action1)> */
		func() bool {
			position21, tokenIndex21 := position, tokenIndex
			{
				position22 := position
				if !_rules[ruleVariable]() {
					goto l21
				}
				if !_rules[ruleSimpleSpaces]() {
					goto l21
				}
				if buffer[position] != rune('=') {
					goto l21
				}
				position++
				if !_rules[ruleSpaces]() {
					goto l21
				}
				if !_rules[ruleExpression]() {
					goto l21
				}
				if !_rules[ruleAction1]() {
					goto l21
				}
				add(ruleAssign, position22)
			}
			return true
		l21:
			position, tokenIndex = position21, tokenIndex21
			return false
		},
		/* 5 FuncCall <- <(Name Action2 SimpleSpaces '(' Spaces FuncArgs Spaces ')' Action3)> */
		func() bool {
			position23, tokenIndex23 := position, tokenIndex
			{
				position24 := position
				if !_rules[ruleName]() {
					goto l23
				}
				if !_rules[ruleAction2]() {
					goto l23
				}
				if !_rules[ruleSimpleSpaces]() {
					goto l23
				}
				if buffer[position] != rune('(') {
					goto l23
				}
				position++
				if !_rules[ruleSpaces]() {
					goto l23
				}
				if !_rules[ruleFuncArgs]() {
					goto l23
				}
				if !_rules[ruleSpaces]() {
					goto l23
				}
				if buffer[position] != rune(')') {
					goto l23
				}
				position++
				if !_rules[ruleAction3]() {
					goto l23
				}
				add(ruleFuncCall, position24)
			}
			return true
		l23:
			position, tokenIndex = position23, tokenIndex23
			return false
		},
		/* 6 FuncArgs <- <((Expression Spaces ',' Spaces)* Expression?)> */
		func() bool {
			{
				position26 := position
			l27:
				{
					position28, tokenIndex28 := position, tokenIndex
					if !_rules[ruleExpression]() {
						goto l28
					}
					if !_rules[ruleSpaces]() {
						goto l28
					}
					if buffer[position] != rune(',') {
						goto l28
					}
					position++
					if !_rules[ruleSpaces]() {
						goto l28
					}
					goto l27
				l28:
					position, tokenIndex = position28, tokenIndex28
				}
				{
					position29, tokenIndex29 := position, tokenIndex
					if !_rules[ruleExpression]() {
						goto l29
					}
					goto l30
				l29:
					position, tokenIndex = position29, tokenIndex29
				}
			l30:
				add(ruleFuncArgs, position26)
			}
			return true
		},
		/* 7 Expression <- <(Binop / NoBinopExpression)> */
		func() bool {
			position31, tokenIndex31 := position, tokenIndex
			{
				position32 := position
				{
					position33, tokenIndex33 := position, tokenIndex
					if !_rules[ruleBinop]() {
						goto l34
					}
					goto l33
				l34:
					position, tokenIndex = position33, tokenIndex33
					if !_rules[ruleNoBinopExpression]() {
						goto l31
					}
				}
			l33:
				add(ruleExpression, position32)
			}
			return true
		l31:
			position, tokenIndex = position31, tokenIndex31
			return false
		},
		/* 8 NoBinopExpression <- <(Unop / NoOpExpression)> */
		func() bool {
			position35, tokenIndex35 := position, tokenIndex
			{
				position36 := position
				{
					position37, tokenIndex37 := position, tokenIndex
					if !_rules[ruleUnop]() {
						goto l38
					}
					goto l37
				l38:
					position, tokenIndex = position37, tokenIndex37
					if !_rules[ruleNoOpExpression]() {
						goto l35
					}
				}
			l37:
				add(ruleNoBinopExpression, position36)
			}
			return true
		l35:
			position, tokenIndex = position35, tokenIndex35
			return false
		},
		/* 9 NoOpExpression <- <(Litteral / Variable / ('(' Spaces Expression Spaces ')'))> */
		func() bool {
			position39, tokenIndex39 := position, tokenIndex
			{
				position40 := position
				{
					position41, tokenIndex41 := position, tokenIndex
					if !_rules[ruleLitteral]() {
						goto l42
					}
					goto l41
				l42:
					position, tokenIndex = position41, tokenIndex41
					if !_rules[ruleVariable]() {
						goto l43
					}
					goto l41
				l43:
					position, tokenIndex = position41, tokenIndex41
					if buffer[position] != rune('(') {
						goto l39
					}
					position++
					if !_rules[ruleSpaces]() {
						goto l39
					}
					if !_rules[ruleExpression]() {
						goto l39
					}
					if !_rules[ruleSpaces]() {
						goto l39
					}
					if buffer[position] != rune(')') {
						goto l39
					}
					position++
				}
			l41:
				add(ruleNoOpExpression, position40)
			}
			return true
		l39:
			position, tokenIndex = position39, tokenIndex39
			return false
		},
		/* 10 Litteral <- <(Number Action4)> */
		func() bool {
			position44, tokenIndex44 := position, tokenIndex
			{
				position45 := position
				if !_rules[ruleNumber]() {
					goto l44
				}
				if !_rules[ruleAction4]() {
					goto l44
				}
				add(ruleLitteral, position45)
			}
			return true
		l44:
			position, tokenIndex = position44, tokenIndex44
			return false
		},
		/* 11 Variable <- <(Name Action5)> */
		func() bool {
			position46, tokenIndex46 := position, tokenIndex
			{
				position47 := position
				if !_rules[ruleName]() {
					goto l46
				}
				if !_rules[ruleAction5]() {
					goto l46
				}
				add(ruleVariable, position47)
			}
			return true
		l46:
			position, tokenIndex = position46, tokenIndex46
			return false
		},
		/* 12 Binop <- <(NoBinopExpression SimpleSpaces Op Action6 Spaces Expression)> */
		func() bool {
			position48, tokenIndex48 := position, tokenIndex
			{
				position49 := position
				if !_rules[ruleNoBinopExpression]() {
					goto l48
				}
				if !_rules[ruleSimpleSpaces]() {
					goto l48
				}
				if !_rules[ruleOp]() {
					goto l48
				}
				if !_rules[ruleAction6]() {
					goto l48
				}
				if !_rules[ruleSpaces]() {
					goto l48
				}
				if !_rules[ruleExpression]() {
					goto l48
				}
				add(ruleBinop, position49)
			}
			return true
		l48:
			position, tokenIndex = position48, tokenIndex48
			return false
		},
		/* 13 Unop <- <(Op Spaces NoOpExpression Action7)> */
		func() bool {
			position50, tokenIndex50 := position, tokenIndex
			{
				position51 := position
				if !_rules[ruleOp]() {
					goto l50
				}
				if !_rules[ruleSpaces]() {
					goto l50
				}
				if !_rules[ruleNoOpExpression]() {
					goto l50
				}
				if !_rules[ruleAction7]() {
					goto l50
				}
				add(ruleUnop, position51)
			}
			return true
		l50:
			position, tokenIndex = position50, tokenIndex50
			return false
		},
		/* 14 Op <- <'+'> */
		func() bool {
			position52, tokenIndex52 := position, tokenIndex
			{
				position53 := position
				if buffer[position] != rune('+') {
					goto l52
				}
				position++
				add(ruleOp, position53)
			}
			return true
		l52:
			position, tokenIndex = position52, tokenIndex52
			return false
		},
		/* 15 Number <- <<Digit+>> */
		func() bool {
			position54, tokenIndex54 := position, tokenIndex
			{
				position55 := position
				{
					position56 := position
					if !_rules[ruleDigit]() {
						goto l54
					}
				l57:
					{
						position58, tokenIndex58 := position, tokenIndex
						if !_rules[ruleDigit]() {
							goto l58
						}
						goto l57
					l58:
						position, tokenIndex = position58, tokenIndex58
					}
					add(rulePegText, position56)
				}
				add(ruleNumber, position55)
			}
			return true
		l54:
			position, tokenIndex = position54, tokenIndex54
			return false
		},
		/* 16 Name <- <<(AlphaChar AlphaNumericalChar*)>> */
		func() bool {
			position59, tokenIndex59 := position, tokenIndex
			{
				position60 := position
				{
					position61 := position
					if !_rules[ruleAlphaChar]() {
						goto l59
					}
				l62:
					{
						position63, tokenIndex63 := position, tokenIndex
						if !_rules[ruleAlphaNumericalChar]() {
							goto l63
						}
						goto l62
					l63:
						position, tokenIndex = position63, tokenIndex63
					}
					add(rulePegText, position61)
				}
				add(ruleName, position60)
			}
			return true
		l59:
			position, tokenIndex = position59, tokenIndex59
			return false
		},
		/* 17 AlphaChar <- <([a-z] / [A-Z] / '_')> */
		func() bool {
			position64, tokenIndex64 := position, tokenIndex
			{
				position65 := position
				{
					position66, tokenIndex66 := position, tokenIndex
					if c := buffer[position]; c < rune('a') || c > rune('z') {
						goto l67
					}
					position++
					goto l66
				l67:
					position, tokenIndex = position66, tokenIndex66
					if c := buffer[position]; c < rune('A') || c > rune('Z') {
						goto l68
					}
					position++
					goto l66
				l68:
					position, tokenIndex = position66, tokenIndex66
					if buffer[position] != rune('_') {
						goto l64
					}
					position++
				}
			l66:
				add(ruleAlphaChar, position65)
			}
			return true
		l64:
			position, tokenIndex = position64, tokenIndex64
			return false
		},
		/* 18 Digit <- <[0-9]> */
		func() bool {
			position69, tokenIndex69 := position, tokenIndex
			{
				position70 := position
				if c := buffer[position]; c < rune('0') || c > rune('9') {
					goto l69
				}
				position++
				add(ruleDigit, position70)
			}
			return true
		l69:
			position, tokenIndex = position69, tokenIndex69
			return false
		},
		/* 19 AlphaNumericalChar <- <(AlphaChar / Digit)> */
		func() bool {
			position71, tokenIndex71 := position, tokenIndex
			{
				position72 := position
				{
					position73, tokenIndex73 := position, tokenIndex
					if !_rules[ruleAlphaChar]() {
						goto l74
					}
					goto l73
				l74:
					position, tokenIndex = position73, tokenIndex73
					if !_rules[ruleDigit]() {
						goto l71
					}
				}
			l73:
				add(ruleAlphaNumericalChar, position72)
			}
			return true
		l71:
			position, tokenIndex = position71, tokenIndex71
			return false
		},
		/* 20 Comment <- <('#' (!Newline .)* Newline)> */
		func() bool {
			position75, tokenIndex75 := position, tokenIndex
			{
				position76 := position
				if buffer[position] != rune('#') {
					goto l75
				}
				position++
			l77:
				{
					position78, tokenIndex78 := position, tokenIndex
					{
						position79, tokenIndex79 := position, tokenIndex
						if !_rules[ruleNewline]() {
							goto l79
						}
						goto l78
					l79:
						position, tokenIndex = position79, tokenIndex79
					}
					if !matchDot() {
						goto l78
					}
					goto l77
				l78:
					position, tokenIndex = position78, tokenIndex78
				}
				if !_rules[ruleNewline]() {
					goto l75
				}
				add(ruleComment, position76)
			}
			return true
		l75:
			position, tokenIndex = position75, tokenIndex75
			return false
		},
		/* 21 Spaces <- <Space*> */
		func() bool {
			{
				position81 := position
			l82:
				{
					position83, tokenIndex83 := position, tokenIndex
					if !_rules[ruleSpace]() {
						goto l83
					}
					goto l82
				l83:
					position, tokenIndex = position83, tokenIndex83
				}
				add(ruleSpaces, position81)
			}
			return true
		},
		/* 22 Space <- <(SimpleSpace / Newline / Comment)> */
		func() bool {
			position84, tokenIndex84 := position, tokenIndex
			{
				position85 := position
				{
					position86, tokenIndex86 := position, tokenIndex
					if !_rules[ruleSimpleSpace]() {
						goto l87
					}
					goto l86
				l87:
					position, tokenIndex = position86, tokenIndex86
					if !_rules[ruleNewline]() {
						goto l88
					}
					goto l86
				l88:
					position, tokenIndex = position86, tokenIndex86
					if !_rules[ruleComment]() {
						goto l84
					}
				}
			l86:
				add(ruleSpace, position85)
			}
			return true
		l84:
			position, tokenIndex = position84, tokenIndex84
			return false
		},
		/* 23 SimpleSpaces <- <SimpleSpace*> */
		func() bool {
			{
				position90 := position
			l91:
				{
					position92, tokenIndex92 := position, tokenIndex
					if !_rules[ruleSimpleSpace]() {
						goto l92
					}
					goto l91
				l92:
					position, tokenIndex = position92, tokenIndex92
				}
				add(ruleSimpleSpaces, position90)
			}
			return true
		},
		/* 24 SimpleSpace <- <(' ' / '\t')> */
		func() bool {
			position93, tokenIndex93 := position, tokenIndex
			{
				position94 := position
				{
					position95, tokenIndex95 := position, tokenIndex
					if buffer[position] != rune(' ') {
						goto l96
					}
					position++
					goto l95
				l96:
					position, tokenIndex = position95, tokenIndex95
					if buffer[position] != rune('\t') {
						goto l93
					}
					position++
				}
			l95:
				add(ruleSimpleSpace, position94)
			}
			return true
		l93:
			position, tokenIndex = position93, tokenIndex93
			return false
		},
		/* 25 Newline <- <(('\r' '\n') / '\n' / '\r')> */
		func() bool {
			position97, tokenIndex97 := position, tokenIndex
			{
				position98 := position
				{
					position99, tokenIndex99 := position, tokenIndex
					if buffer[position] != rune('\r') {
						goto l100
					}
					position++
					if buffer[position] != rune('\n') {
						goto l100
					}
					position++
					goto l99
				l100:
					position, tokenIndex = position99, tokenIndex99
					if buffer[position] != rune('\n') {
						goto l101
					}
					position++
					goto l99
				l101:
					position, tokenIndex = position99, tokenIndex99
					if buffer[position] != rune('\r') {
						goto l97
					}
					position++
				}
			l99:
				add(ruleNewline, position98)
			}
			return true
		l97:
			position, tokenIndex = position97, tokenIndex97
			return false
		},
		/* 27 Action0 <- <{ p.AddStatement() }> */
		func() bool {
			{
				add(ruleAction0, position)
			}
			return true
		},
		/* 28 Action1 <- <{ p.AddAssign() }> */
		func() bool {
			{
				add(ruleAction1, position)
			}
			return true
		},
		/* 29 Action2 <- <{ p.AddFuncName(text) }> */
		func() bool {
			{
				add(ruleAction2, position)
			}
			return true
		},
		/* 30 Action3 <- <{ p.AddFuncCall() }> */
		func() bool {
			{
				add(ruleAction3, position)
			}
			return true
		},
		/* 31 Action4 <- <{ p.AddLitteral(text) }> */
		func() bool {
			{
				add(ruleAction4, position)
			}
			return true
		},
		/* 32 Action5 <- <{ p.AddVariable(text) }> */
		func() bool {
			{
				add(ruleAction5, position)
			}
			return true
		},
		/* 33 Action6 <- <{ p.AddBinop(text) }> */
		func() bool {
			{
				add(ruleAction6, position)
			}
			return true
		},
		/* 34 Action7 <- <{ p.AddUnop(text) }> */
		func() bool {
			{
				add(ruleAction7, position)
			}
			return true
		},
		nil,
	}
	p.rules = _rules
}
