//line y.y:2
package jaml

import __yyfmt__ "fmt"

//line y.y:2
import (
	"fmt"
	"reflect"
	"strconv"
)

type fieldInfo struct {
	pth string
	val *Value
	imp string
	lxr *lexer
}

type fieldPath struct {
	pth string
	arr bool
	key *Value
	lxr *lexer
}

//line y.y:26
type yySymType struct {
	yys int
	val *Value
	fld fieldInfo
	str string
	pth fieldPath
	tkn token
	tag map[string]string
}

const DEC = 57346
const OCT = 57347
const HEX = 57348
const FLOAT = 57349
const STRING = 57350
const RUNE = 57351
const BOOL = 57352
const IDENTIFIER = 57353
const NEWLINE = 57354
const SQUARE_OPEN = 57355
const SQUARE_CLOSE = 57356
const INDENT = 57357
const DEDENT = 57358
const IMPORT = 57359
const TIME = 57360
const DURATION = 57361

var yyToknames = [...]string{
	"$end",
	"error",
	"$unk",
	"DEC",
	"OCT",
	"HEX",
	"FLOAT",
	"STRING",
	"RUNE",
	"BOOL",
	"IDENTIFIER",
	"NEWLINE",
	"SQUARE_OPEN",
	"SQUARE_CLOSE",
	"INDENT",
	"DEDENT",
	"IMPORT",
	"TIME",
	"DURATION",
	"'.'",
	"'['",
	"']'",
	"':'",
	"'@'",
	"'`'",
	"','",
	"'='",
}
var yyStatenames = [...]string{}

const yyEofCode = 1
const yyErrCode = 2
const yyMaxDepth = 200

//line y.y:341

var (
	tokenTypeMap = map[tokenType]int{
		tokenEOF:         0,
		tokenError:       0,
		tokenIdentifier:  IDENTIFIER,
		tokenDec:         DEC,
		tokenOct:         OCT,
		tokenHex:         HEX,
		tokenFloat:       FLOAT,
		tokenString:      STRING,
		tokenRune:        RUNE,
		tokenBool:        BOOL,
		tokenIndent:      INDENT,
		tokenDedent:      DEDENT,
		tokenChar:        0,
		tokenNewline:     NEWLINE,
		tokenSquareOpen:  SQUARE_OPEN,
		tokenSquareClose: SQUARE_CLOSE,
		tokenImport:      IMPORT,
	}
)

func (l *lexer) Lex(lval *yySymType) (res int) {
	if l.err != nil {
		return 0
	}
	token := <-l.tokens
	switch token.typ {
	case tokenError:
		l.err = token.err
		res = 0
	case tokenChar:
		res = int(token.run)
	default:
		lval.tkn = token
		res = tokenTypeMap[token.typ]
	}
	l.debug("lex: generate token %d: %s", res, token)
	return
}

func (l *lexer) Error(s string) {
	if l.err == nil {
		l.err = fmt.Errorf("yacc error: %s", s)
	} else {
		l.err = fmt.Errorf("yacc error: %s, caused by lexer error: %s", s, l.err)
	}
}

//line yacctab:1
var yyExca = [...]int{
	-1, 1,
	1, -1,
	-2, 0,
	-1, 8,
	23, 11,
	-2, 9,
	-1, 32,
	23, 11,
	-2, 10,
}

const yyNprod = 39
const yyPrivate = 57344

var yyTokenNames []string
var yyStates []string

const yyLast = 79

var yyAct = [...]int{

	16, 2, 23, 24, 25, 26, 27, 29, 28, 60,
	54, 30, 52, 53, 36, 11, 21, 22, 49, 43,
	14, 13, 18, 23, 24, 25, 26, 27, 29, 28,
	50, 42, 51, 23, 24, 25, 26, 27, 29, 28,
	9, 34, 30, 8, 59, 19, 46, 21, 22, 5,
	6, 55, 38, 56, 9, 37, 57, 48, 32, 3,
	33, 61, 58, 40, 31, 39, 12, 47, 35, 15,
	20, 41, 10, 4, 17, 7, 45, 44, 1,
}
var yyPact = [...]int{

	32, -1000, 42, -1000, -8, 58, -1000, 1, -1, 32,
	-1000, -2, -1000, 47, 19, -1000, -1000, -11, 44, -1000,
	-1000, 57, 55, -1000, -1000, -1000, -1000, -1000, -1000, -1000,
	29, -1000, -1, -3, -1000, 34, 46, -1000, -1000, -1000,
	-1000, 4, -1000, -1000, -1000, -1000, 17, -13, -17, -1000,
	29, 32, -1000, 45, 54, -1000, 28, -18, -1000, -1000,
	53, -1000,
}
var yyPgo = [...]int{

	0, 78, 1, 77, 76, 75, 74, 50, 73, 72,
	0, 71, 70, 45, 68, 67, 59,
}
var yyR1 = [...]int{

	0, 1, 2, 2, 2, 16, 16, 8, 8, 5,
	5, 7, 7, 7, 9, 9, 9, 6, 6, 14,
	14, 15, 15, 13, 13, 13, 13, 13, 13, 13,
	10, 10, 10, 10, 12, 11, 11, 3, 4,
}
var yyR2 = [...]int{

	0, 1, 1, 3, 2, 2, 2, 1, 3, 1,
	3, 1, 4, 3, 2, 4, 3, 0, 2, 0,
	3, 5, 3, 1, 1, 1, 1, 1, 1, 1,
	1, 1, 2, 2, 3, 1, 3, 1, 4,
}
var yyChk = [...]int{

	-1000, -1, -2, -16, -8, 17, -7, -5, 11, 12,
	-9, 23, 8, 20, 21, -16, -10, -6, 24, -13,
	-12, 18, 19, 4, 5, 6, 7, 8, 10, 9,
	13, -7, 11, -13, 22, -14, 25, 11, 8, 8,
	8, -11, -10, 22, -3, -4, 12, -15, 11, 14,
	26, 15, 25, 26, 27, -10, -2, 11, 8, 16,
	27, 8,
}
var yyDef = [...]int{

	0, -2, 1, 2, 0, 0, 7, 0, -2, 4,
	5, 17, 6, 0, 0, 3, 14, 19, 0, 30,
	31, 0, 0, 23, 24, 25, 26, 27, 28, 29,
	0, 8, -2, 0, 13, 0, 0, 16, 18, 32,
	33, 0, 35, 12, 15, 37, 0, 0, 0, 34,
	0, 0, 20, 0, 0, 36, 0, 0, 22, 38,
	0, 21,
}
var yyTok1 = [...]int{

	1, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 26, 3, 20, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 23, 3,
	3, 27, 3, 3, 24, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 21, 3, 22, 3, 3, 25,
}
var yyTok2 = [...]int{

	2, 3, 4, 5, 6, 7, 8, 9, 10, 11,
	12, 13, 14, 15, 16, 17, 18, 19,
}
var yyTok3 = [...]int{
	0,
}

var yyErrorMessages = [...]struct {
	state int
	token int
	msg   string
}{}

//line yaccpar:1

/*	parser for yacc output	*/

var (
	yyDebug        = 0
	yyErrorVerbose = false
)

type yyLexer interface {
	Lex(lval *yySymType) int
	Error(s string)
}

type yyParser interface {
	Parse(yyLexer) int
	Lookahead() int
}

type yyParserImpl struct {
	lookahead func() int
}

func (p *yyParserImpl) Lookahead() int {
	return p.lookahead()
}

func yyNewParser() yyParser {
	p := &yyParserImpl{
		lookahead: func() int { return -1 },
	}
	return p
}

const yyFlag = -1000

func yyTokname(c int) string {
	if c >= 1 && c-1 < len(yyToknames) {
		if yyToknames[c-1] != "" {
			return yyToknames[c-1]
		}
	}
	return __yyfmt__.Sprintf("tok-%v", c)
}

func yyStatname(s int) string {
	if s >= 0 && s < len(yyStatenames) {
		if yyStatenames[s] != "" {
			return yyStatenames[s]
		}
	}
	return __yyfmt__.Sprintf("state-%v", s)
}

func yyErrorMessage(state, lookAhead int) string {
	const TOKSTART = 4

	if !yyErrorVerbose {
		return "syntax error"
	}

	for _, e := range yyErrorMessages {
		if e.state == state && e.token == lookAhead {
			return "syntax error: " + e.msg
		}
	}

	res := "syntax error: unexpected " + yyTokname(lookAhead)

	// To match Bison, suggest at most four expected tokens.
	expected := make([]int, 0, 4)

	// Look for shiftable tokens.
	base := yyPact[state]
	for tok := TOKSTART; tok-1 < len(yyToknames); tok++ {
		if n := base + tok; n >= 0 && n < yyLast && yyChk[yyAct[n]] == tok {
			if len(expected) == cap(expected) {
				return res
			}
			expected = append(expected, tok)
		}
	}

	if yyDef[state] == -2 {
		i := 0
		for yyExca[i] != -1 || yyExca[i+1] != state {
			i += 2
		}

		// Look for tokens that we accept or reduce.
		for i += 2; yyExca[i] >= 0; i += 2 {
			tok := yyExca[i]
			if tok < TOKSTART || yyExca[i+1] == 0 {
				continue
			}
			if len(expected) == cap(expected) {
				return res
			}
			expected = append(expected, tok)
		}

		// If the default action is to accept or reduce, give up.
		if yyExca[i+1] != 0 {
			return res
		}
	}

	for i, tok := range expected {
		if i == 0 {
			res += ", expecting "
		} else {
			res += " or "
		}
		res += yyTokname(tok)
	}
	return res
}

func yylex1(lex yyLexer, lval *yySymType) (char, token int) {
	token = 0
	char = lex.Lex(lval)
	if char <= 0 {
		token = yyTok1[0]
		goto out
	}
	if char < len(yyTok1) {
		token = yyTok1[char]
		goto out
	}
	if char >= yyPrivate {
		if char < yyPrivate+len(yyTok2) {
			token = yyTok2[char-yyPrivate]
			goto out
		}
	}
	for i := 0; i < len(yyTok3); i += 2 {
		token = yyTok3[i+0]
		if token == char {
			token = yyTok3[i+1]
			goto out
		}
	}

out:
	if token == 0 {
		token = yyTok2[1] /* unknown char */
	}
	if yyDebug >= 3 {
		__yyfmt__.Printf("lex %s(%d)\n", yyTokname(token), uint(char))
	}
	return char, token
}

func yyParse(yylex yyLexer) int {
	return yyNewParser().Parse(yylex)
}

func (yyrcvr *yyParserImpl) Parse(yylex yyLexer) int {
	var yyn int
	var yylval yySymType
	var yyVAL yySymType
	var yyDollar []yySymType
	_ = yyDollar // silence set and not used
	yyS := make([]yySymType, yyMaxDepth)

	Nerrs := 0   /* number of errors */
	Errflag := 0 /* error recovery flag */
	yystate := 0
	yychar := -1
	yytoken := -1 // yychar translated into internal numbering
	yyrcvr.lookahead = func() int { return yychar }
	defer func() {
		// Make sure we report no lookahead when not parsing.
		yystate = -1
		yychar = -1
		yytoken = -1
	}()
	yyp := -1
	goto yystack

ret0:
	return 0

ret1:
	return 1

yystack:
	/* put a state and value onto the stack */
	if yyDebug >= 4 {
		__yyfmt__.Printf("char %v in %v\n", yyTokname(yytoken), yyStatname(yystate))
	}

	yyp++
	if yyp >= len(yyS) {
		nyys := make([]yySymType, len(yyS)*2)
		copy(nyys, yyS)
		yyS = nyys
	}
	yyS[yyp] = yyVAL
	yyS[yyp].yys = yystate

yynewstate:
	yyn = yyPact[yystate]
	if yyn <= yyFlag {
		goto yydefault /* simple state */
	}
	if yychar < 0 {
		yychar, yytoken = yylex1(yylex, &yylval)
	}
	yyn += yytoken
	if yyn < 0 || yyn >= yyLast {
		goto yydefault
	}
	yyn = yyAct[yyn]
	if yyChk[yyn] == yytoken { /* valid shift */
		yychar = -1
		yytoken = -1
		yyVAL = yylval
		yystate = yyn
		if Errflag > 0 {
			Errflag--
		}
		goto yystack
	}

yydefault:
	/* default state action */
	yyn = yyDef[yystate]
	if yyn == -2 {
		if yychar < 0 {
			yychar, yytoken = yylex1(yylex, &yylval)
		}

		/* look through exception table */
		xi := 0
		for {
			if yyExca[xi+0] == -1 && yyExca[xi+1] == yystate {
				break
			}
			xi += 2
		}
		for xi += 2; ; xi += 2 {
			yyn = yyExca[xi+0]
			if yyn < 0 || yyn == yytoken {
				break
			}
		}
		yyn = yyExca[xi+1]
		if yyn < 0 {
			goto ret0
		}
	}
	if yyn == 0 {
		/* error ... attempt to resume parsing */
		switch Errflag {
		case 0: /* brand new error */
			yylex.Error(yyErrorMessage(yystate, yytoken))
			Nerrs++
			if yyDebug >= 1 {
				__yyfmt__.Printf("%s", yyStatname(yystate))
				__yyfmt__.Printf(" saw %s\n", yyTokname(yytoken))
			}
			fallthrough

		case 1, 2: /* incompletely recovered error ... try again */
			Errflag = 3

			/* find a state where "error" is a legal shift action */
			for yyp >= 0 {
				yyn = yyPact[yyS[yyp].yys] + yyErrCode
				if yyn >= 0 && yyn < yyLast {
					yystate = yyAct[yyn] /* simulate a shift of "error" */
					if yyChk[yystate] == yyErrCode {
						goto yystack
					}
				}

				/* the current p has no shift on "error", pop stack */
				if yyDebug >= 2 {
					__yyfmt__.Printf("error recovery pops state %d\n", yyS[yyp].yys)
				}
				yyp--
			}
			/* there is no state on the stack with an error shift ... abort */
			goto ret1

		case 3: /* no shift yet; clobber input char */
			if yyDebug >= 2 {
				__yyfmt__.Printf("error recovery discards %s\n", yyTokname(yytoken))
			}
			if yytoken == yyEofCode {
				goto ret1
			}
			yychar = -1
			yytoken = -1
			goto yynewstate /* try again in the same state */
		}
	}

	/* reduction by production yyn */
	if yyDebug >= 2 {
		__yyfmt__.Printf("reduce %v in:\n\t%v\n", yyn, yyStatname(yystate))
	}

	yynt := yyn
	yypt := yyp
	_ = yypt // guard against "declared and not used"

	yyp -= yyR2[yyn]
	// yyp is now the index of $0. Perform the default action. Iff the
	// reduced production is ε, $1 is possibly out of range.
	if yyp+1 >= len(yyS) {
		nyys := make([]yySymType, len(yyS)*2)
		copy(nyys, yyS)
		yyS = nyys
	}
	yyVAL = yyS[yyp+1]

	/* consult goto table to find next state */
	yyn = yyR1[yyn]
	yyg := yyPgo[yyn]
	yyj := yyg + yyS[yyp].yys + 1

	if yyj >= yyLast {
		yystate = yyAct[yyg]
	} else {
		yystate = yyAct[yyj]
		if yyChk[yystate] != -yyn {
			yystate = yyAct[yyg]
		}
	}
	// dummy call; replaced with literal code
	switch yynt {

	case 1:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line y.y:47
		{
			yyVAL.val = yyDollar[1].val
			yyVAL.val.lexer.root <- yyVAL.val
			yyVAL.val.lexer.debug("yacc: root -> list")
		}
	case 2:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line y.y:55
		{
			yyVAL.val = &Value{
				Kind:   reflect.Struct,
				Fields: make(map[string]*Value),
				lexer:  yyDollar[1].fld.val.lexer,
			}
			if yyDollar[1].fld.imp != "" {
				yyVAL.val.Import(yyDollar[1].fld.imp)
			} else {
				yyVAL.val.SetField(yyDollar[1].fld.pth, yyDollar[1].fld.val)
			}
			yyVAL.val.lexer.debug("yacc: list -> field")
		}
	case 3:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line y.y:69
		{
			yyVAL.val = yyDollar[1].val
			if yyDollar[3].fld.imp != "" {
				yyVAL.val.Import(yyDollar[3].fld.imp)
			} else {
				yyVAL.val.SetField(yyDollar[3].fld.pth, yyDollar[3].fld.val)
			}
			yyVAL.val.lexer.debug("yacc: list -> list NEWLINE field")
		}
	case 4:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line y.y:79
		{
			yyVAL.val = yyDollar[1].val
			yyVAL.val.lexer.debug("yacc: list -> list NEWLINE")
		}
	case 5:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line y.y:86
		{
			yyVAL.fld.pth = yyDollar[1].pth.pth
			yyVAL.fld.val = yyDollar[2].val
			yyVAL.fld.lxr = yyDollar[1].pth.lxr
			yyVAL.fld.lxr.debug("yacc: field -> field_path assign")
		}
	case 6:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line y.y:93
		{
			yyVAL.fld.imp = yyDollar[2].tkn.str
			yyVAL.fld.lxr = yyDollar[1].tkn.lxr
			yyVAL.fld.lxr.debug("yacc: field -> IMPORT STRING : %s", yyDollar[2].tkn)
		}
	case 7:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line y.y:101
		{
			yyVAL.pth = yyDollar[1].pth
			yyVAL.pth.lxr.debug("yacc: field_path -> field_name")
		}
	case 8:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line y.y:106
		{
			yyVAL.pth = yyDollar[3].pth
			yyVAL.pth.pth = yyDollar[1].str + "." + yyVAL.pth.pth
			yyVAL.pth.lxr.debug("yacc: field_path -> field_path_head '.' field_name")
		}
	case 9:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line y.y:114
		{
			yyVAL.str = yyDollar[1].tkn.str
			yyDollar[1].tkn.lxr.debug("yacc: field_path_head -> IDENTIFIER : %s", yyDollar[1].tkn)
		}
	case 10:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line y.y:119
		{
			yyVAL.str = yyDollar[1].str + "." + yyDollar[3].tkn.str
			yyDollar[3].tkn.lxr.debug("yacc: field_path_head -> field_path_head '.' IDENTIFIER : %s", yyDollar[3].tkn)
		}
	case 11:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line y.y:126
		{
			yyVAL.pth.pth = yyDollar[1].tkn.str
			yyVAL.pth.lxr = yyDollar[1].tkn.lxr
			yyVAL.pth.lxr.debug("yacc: field_name -> IDENTIFIER : %s", yyDollar[1].tkn)
		}
	case 12:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line y.y:132
		{
			yyVAL.pth.pth = yyDollar[1].tkn.str
			yyVAL.pth.key = yyDollar[3].val
			yyVAL.pth.lxr = yyDollar[1].tkn.lxr
			yyVAL.pth.lxr.debug("yacc: field_name -> IDENTIFIER '[' primary_value ']' : %s, %s", yyDollar[1].tkn, yyDollar[3].val)
		}
	case 13:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line y.y:139
		{
			yyVAL.pth.pth = yyDollar[1].tkn.str
			yyVAL.pth.arr = true
			yyVAL.pth.lxr = yyDollar[1].tkn.lxr
			yyVAL.pth.lxr.debug("yacc: field_name -> IDENTIFIER '[' ']' : %s", yyDollar[1].tkn)
		}
	case 14:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line y.y:148
		{
			yyVAL.val = yyDollar[2].val
			yyVAL.val.lexer.debug("yacc: assign -> ':' non_object")
		}
	case 15:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line y.y:153
		{
			yyVAL.val = yyDollar[4].val
			yyDollar[4].val.Type = yyDollar[2].str
			yyDollar[4].val.Tags = yyDollar[3].tag
			yyVAL.val.lexer.debug("yacc: assign -> ':' object")
		}
	case 16:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line y.y:160
		{
			yyVAL.val = &Value{
				Kind:    reflect.Ptr,
				Primary: yyDollar[3].tkn.str,
				lexer:   yyDollar[3].tkn.lxr,
			}
			yyVAL.val.lexer.debug("yacc: assign -> ':' '@' IDENTIFIER : %s", yyDollar[3].tkn)
		}
	case 17:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line y.y:171
		{
		}
	case 18:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line y.y:174
		{
			yyVAL.str = yyDollar[2].tkn.str
			yyDollar[2].tkn.lxr.debug("yacc: assign -> '@' STRING : %s", yyDollar[2].tkn)
		}
	case 19:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line y.y:181
		{
		}
	case 20:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line y.y:184
		{
			yyVAL.tag = yyDollar[2].tag
		}
	case 21:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line y.y:190
		{
			yyVAL.tag = yyDollar[1].tag
			yyDollar[1].tag[yyDollar[3].tkn.str] = yyDollar[5].tkn.str
			yyDollar[3].tkn.lxr.debug("yacc: tag_list -> tag_list ',' IDENTIFIER '=' STRING : %s, %s", yyDollar[3].tkn, yyDollar[5].tkn)
		}
	case 22:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line y.y:196
		{
			yyVAL.tag = make(map[string]string)
			yyVAL.tag[yyDollar[1].tkn.str] = yyDollar[3].tkn.str
			yyDollar[1].tkn.lxr.debug("yacc: tag_list -> IDENTIFIER '=' STRING : %s, %s", yyDollar[1].tkn, yyDollar[3].tkn)
		}
	case 23:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line y.y:204
		{
			ival, _ := strconv.ParseInt(yyDollar[1].tkn.str, 10, 64)
			yyVAL.val = &Value{
				Kind:    reflect.Int64,
				Primary: ival,
				lexer:   yyDollar[1].tkn.lxr,
			}
			yyVAL.val.lexer.debug("yacc: primary_value -> DEC: %d", ival)
		}
	case 24:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line y.y:214
		{
			ival, _ := strconv.ParseInt(yyDollar[1].tkn.str[1:], 8, 64)
			yyVAL.val = &Value{
				Kind:    reflect.Int64,
				Primary: ival,
				lexer:   yyDollar[1].tkn.lxr,
			}
			yyVAL.val.lexer.debug("yacc: primary_value -> OCT: %d(%s)", ival, yyDollar[1].tkn.str)
		}
	case 25:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line y.y:224
		{
			ival, _ := strconv.ParseInt(yyDollar[1].tkn.str[2:], 16, 64)
			yyVAL.val = &Value{
				Kind:    reflect.Int64,
				Primary: ival,
				lexer:   yyDollar[1].tkn.lxr,
			}
			yyVAL.val.lexer.debug("yacc: primary_value -> HEX: %d(%s)", ival, yyDollar[1].tkn.str)
		}
	case 26:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line y.y:234
		{
			fval, _ := strconv.ParseFloat(yyDollar[1].tkn.str, 64)
			yyVAL.val = &Value{
				Kind:    reflect.Int64,
				Primary: fval,
				lexer:   yyDollar[1].tkn.lxr,
			}
		}
	case 27:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line y.y:243
		{
			yyVAL.val = &Value{
				Kind:    reflect.String,
				Primary: yyDollar[1].tkn.str,
				lexer:   yyDollar[1].tkn.lxr,
			}
			yyVAL.val.lexer.debug("yacc: primary_value -> STRING: %s", yyDollar[1].tkn)
		}
	case 28:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line y.y:252
		{
			bval, _ := strconv.ParseBool(yyDollar[1].tkn.str)
			yyVAL.val = &Value{
				Kind:    reflect.String,
				Primary: bval,
				lexer:   yyDollar[1].tkn.lxr,
			}
			yyVAL.val.lexer.debug("yacc: primary_value -> BOOL: %s", yyDollar[1].tkn)
		}
	case 29:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line y.y:262
		{
			yyVAL.val = &Value{
				Kind:    reflect.Int32,
				Primary: yyDollar[1].tkn.run,
				lexer:   yyDollar[1].tkn.lxr,
			}
			yyVAL.val.lexer.debug("yacc: primary_value -> RUNE: %q", yyDollar[1].tkn.run)
		}
	case 30:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line y.y:273
		{
			yyVAL.val = yyDollar[1].val
			yyVAL.val.lexer.debug("yacc: non_object -> primary_value")
		}
	case 31:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line y.y:278
		{
			yyVAL.val = yyDollar[1].val
			yyVAL.val.lexer.debug("yacc: non_object -> array_literal")
		}
	case 32:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line y.y:283
		{
			yyVAL.val = &Value{
				Kind:    reflect.Struct,
				Type:    "time.Time",
				Primary: yyDollar[2].tkn.str,
			}
			yyVAL.val.lexer.debug("yacc: non_object -> TIME STRING : %s", yyDollar[2].tkn)
		}
	case 33:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line y.y:292
		{
			yyVAL.val = &Value{
				Kind:    reflect.Int64,
				Type:    "time.Duration",
				Primary: yyDollar[2].tkn.str,
			}
			yyVAL.val.lexer.debug("yacc: non_object -> DURATION STRING : %s", yyDollar[2].tkn)
		}
	case 34:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line y.y:303
		{
			yyVAL.val = yyDollar[2].val
			yyVAL.val.lexer.debug("yacc: array_literal -> SQUARE_OPEN array_inside SQUARE_CLOSE")
		}
	case 35:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line y.y:310
		{
			yyVAL.val = &Value{
				Kind:  reflect.Slice,
				Items: make([]*Value, 0, 4),
				lexer: yyDollar[1].val.lexer,
			}
			yyVAL.val.Append("", yyDollar[1].val)
			yyVAL.val.lexer.debug("yacc: array_inside -> non_object")
		}
	case 36:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line y.y:320
		{
			yyVAL.val = yyDollar[1].val
			yyVAL.val.Append("", yyDollar[3].val)
			yyVAL.val.lexer.debug("yacc: array_inside -> array_inside ',' non_object")
		}
	case 37:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line y.y:328
		{
			yyVAL.val = yyDollar[1].val
			yyVAL.val.lexer.debug("yacc: object -> suite")
		}
	case 38:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line y.y:335
		{
			yyVAL.val = yyDollar[3].val
			yyVAL.val.lexer.debug("yacc: suite -> NEWLINE INDENT list DEDENT")
		}
	}
	goto yystack /* stack new state and value */
}
