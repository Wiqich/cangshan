package jaml

import (
	"fmt"
	"io"
	"strconv"
	"unicode"
)

type tokenType int

const (
	tokenEOF tokenType = iota
	tokenError
	tokenIdentifier
	tokenDec
	tokenOct
	tokenHex
	tokenFloat
	tokenString
	tokenBool
	tokenIndent
	tokenDedent
	tokenChar
	tokenRune
	tokenNewline
	tokenSquareOpen
	tokenSquareClose
	tokenImport
	tokenTime
	tokenDuration
)

var (
	tokenNameMap = map[tokenType]string{
		tokenEOF:         "eof",
		tokenError:       "error",
		tokenIdentifier:  "identifier",
		tokenDec:         "dec",
		tokenOct:         "oct",
		tokenHex:         "hex",
		tokenFloat:       "float",
		tokenString:      "string",
		tokenBool:        "bool",
		tokenIndent:      "indent",
		tokenDedent:      "dedent",
		tokenRune:        "rune",
		tokenChar:        "char",
		tokenNewline:     "newline",
		tokenSquareOpen:  "square_open",
		tokenSquareClose: "square_close",
		tokenTime:        "time",
		tokenDuration:    "duration",
	}
)

func (typ tokenType) String() string {
	return tokenNameMap[typ]
}

type token struct {
	typ tokenType
	str string
	run rune
	err error
	lxr *lexer
}

func (i token) String() string {
	switch i.typ {
	case tokenEOF, tokenNewline, tokenSquareOpen, tokenSquareClose, tokenIndent, tokenDedent:
		return i.typ.String()
	case tokenError:
		if i.err == nil {
			return i.typ.String()
		}
		return fmt.Sprintf("%s:%s", i.typ, i.err.Error())
	case tokenRune, tokenChar:
		return fmt.Sprintf("%s:%q", i.typ, i.run)
	default:
		return fmt.Sprintf("%s:%s", i.typ, i.str)
	}
}

type lexer struct {
	text   []rune
	start  int
	pos    int
	line   int
	col    int
	lv     int
	arrlv  int
	err    error
	tokens chan token
	debugs chan<- string
	root   chan *Value
}

func newLexer(text string, debugs chan<- string) *lexer {
	l := newSilentLexer(text, debugs)
	go l.run()
	return l
}

func newSilentLexer(text string, debugs chan<- string) *lexer {
	return &lexer{
		text:   []rune(text),
		line:   1,
		col:    1,
		tokens: make(chan token, 16),
		debugs: debugs,
		root:   make(chan *Value, 1),
	}
}

type stateFunc func() stateFunc

func (l *lexer) run() {
	for state := l.lexIndent; state != nil; {
		state = state()
	}
	close(l.tokens)
}

func (l *lexer) debug(pattern string, params ...interface{}) {
	if l.debugs != nil {
		l.debugs <- fmt.Sprintf(pattern, params...)
	}
}

/*
NUMBER  : DEC
        | OCT
        | HEX
        ;

RUNE    : '\'' CHAR '\''

STRING  : '"' CHARS '"'
        ;

BOOL    : true
        | false
        ;

IDENTIFER   : [A-Z_][A-Za-z_0-9]*
            ;

DEC : -?[0-9]+(\.[0-9]+)?
    ;

OCT : 0[0-7]+
    ;

HEX : 0x[0-9a-fA-F]+
    ;

NON-QUOTE-CHAR  : [^\\"]
                | \.
                ;

INDENT  : ^\.*
        ;

SPACE   : [ \t]+
        ;

NEWLINE : \n

IMPORT  : import
*/

func (l *lexer) lexAny() stateFunc {
	r, err := l.read()
	if err == io.EOF {
		return nil
	} else if err != nil {
		l.emitError(fmt.Errorf("lexAny fail: %s", err.Error()))
		return nil
	}
	switch r {
	case '-':
		return l.lexMinus
	case '"':
		return l.lexString
	case '\'':
		return l.lexRune
	case ' ', '\t':
		l.ignore()
		return l.lexAny
	case '\n':
		l.ignore()
		if l.arrlv > 0 {
			return l.lexAny
		} else {
			l.emitSymbol(tokenNewline)
			return l.lexIndent
		}
	case '0':
		return l.lexZero
	case '1', '2', '3', '4', '5', '6', '7', '8', '9':
		return l.lexInteger
	case '[':
		l.emitSymbol(tokenSquareOpen)
		l.arrlv++
		return l.lexAny
	case ']':
		l.emitSymbol(tokenSquareClose)
		l.arrlv--
		return l.lexAny
	}
	if unicode.IsLower(r) {
		return l.lexReserved
	}
	if unicode.IsUpper(r) || r == '_' {
		return l.lexIdentifier
	}
	l.emitChars()
	return l.lexAny
}

func (l *lexer) read() (r rune, err error) {
	if l.err != nil {
		err = l.err
		return
	}
	if l.pos >= len(l.text) {
		err = io.EOF
	} else {
		r = l.text[l.pos]
		l.pos++
	}
	return
}

func (l *lexer) unread() {
	if l.pos > l.start {
		l.pos--
	}
}

func (l *lexer) ignore() {
	l.col += l.pos - l.start
	l.start = l.pos
}

func (l *lexer) emitChars() {
	for l.start < l.pos {
		l.tokens <- token{
			typ: tokenChar,
			run: l.text[l.start],
			lxr: l,
		}
		l.start++
		l.col++
	}
}

func (l *lexer) emitRune() error {
	r, _, _, err := strconv.UnquoteChar(string(l.text[l.start+1:l.pos]), '\'')
	if err != nil {
		return err
	}
	l.tokens <- token{
		typ: tokenRune,
		run: r,
		lxr: l,
	}
	l.col += l.pos - l.start
	l.start = l.pos
	return nil
}

func (l *lexer) emitSymbol(typ tokenType) {
	l.tokens <- token{
		typ: typ,
		lxr: l,
	}
	if typ == tokenNewline {
		l.line++
		l.col = 1
	} else {
		l.col += l.pos - l.start
	}
	l.start = l.pos
}

func (l *lexer) emitString() error {
	multiline := false
	for _, r := range l.text[l.start:l.pos] {
		if r == '\n' {
			multiline = true
			break
		}
	}
	var str string
	var err error
	if multiline {
		str, err = strconv.Unquote("`" + string(l.text[l.start+1:l.pos-1]) + "`")
	} else {
		str, err = strconv.Unquote(string(l.text[l.start:l.pos]))
	}
	if err != nil {
		return err
	}
	l.tokens <- token{
		typ: tokenString,
		str: str,
		lxr: l,
	}
	for i := l.start; i < l.pos; i++ {
		if l.text[i] == '\n' {
			l.line++
			l.col = 1
		} else {
			l.col++
		}
	}
	l.start = l.pos
	return nil
}

func (l *lexer) emitWord(typ tokenType) {
	l.tokens <- token{
		typ: typ,
		str: string(l.text[l.start:l.pos]),
		lxr: l,
	}
	l.col += l.pos - l.start
	l.start = l.pos
}

func (l *lexer) emitIndents(lv int) {
	for l.lv < lv {
		l.emitSymbol(tokenIndent)
		l.lv++
	}
	for l.lv > lv {
		l.emitSymbol(tokenDedent)
		l.emitSymbol(tokenNewline)
		l.lv--
	}
	l.start = l.pos
}

func (l *lexer) emitLexer(typ tokenType) {
	l.tokens <- token{
		typ: typ,
		lxr: l,
	}
	l.start = l.pos
}

func (l *lexer) emitError(err error) {
	l.tokens <- token{
		typ: tokenError,
		err: err,
		lxr: l,
	}
	l.err = err
}

func (l *lexer) lexMinus() stateFunc {
	r, err := l.read()
	if err == io.EOF {
		l.emitChars()
		return nil
	} else if err != nil {
		l.emitError(fmt.Errorf("lexMinux fail: %s", err.Error()))
		return nil
	}
	if unicode.IsDigit(r) {
		return l.lexInteger
	}
	l.unread()
	l.emitChars()
	return l.lexAny
}

func (l *lexer) lexInteger() stateFunc {
	r, err := l.read()
	if err == io.EOF {
		l.emitWord(tokenDec)
		return nil
	} else if err != nil {
		l.emitError(fmt.Errorf("lexInteger fail: %s", err.Error()))
		return nil
	}
	if unicode.IsDigit(r) {
		return l.lexInteger
	}
	if r == '.' {
		return l.lexFloatPoint
	}
	l.unread()
	l.emitWord(tokenDec)
	return l.lexAny
}

func (l *lexer) lexFloatPoint() stateFunc {
	r, err := l.read()
	if err == io.EOF {
		l.emitError(fmt.Errorf("invalid number at line %d column %d: %s", l.line, l.col, string(l.text[l.start:l.pos])))
		return nil
	} else if err != nil {
		l.emitError(fmt.Errorf("lexFloatPoint fail: %s", err.Error()))
		return nil
	}
	if unicode.IsDigit(r) {
		return l.lexFloat
	}
	l.emitError(fmt.Errorf("invalid number at line %d column %d: %s", l.line, l.col, string(l.text[l.start:l.pos])))
	return nil
}

func (l *lexer) lexFloat() stateFunc {
	r, err := l.read()
	if err == io.EOF {
		l.emitWord(tokenFloat)
		return nil
	} else if err != nil {
		l.emitError(fmt.Errorf("lexFloat fail: %s", err.Error()))
		return nil
	}
	if unicode.IsDigit(r) {
		return l.lexFloat
	}
	l.unread()
	l.emitWord(tokenFloat)
	return l.lexAny
}

func (l *lexer) lexString() stateFunc {
	r, err := l.read()
	if err == io.EOF {
		l.emitError(fmt.Errorf("unfinished string at line %d column %d", l.line, l.col))
		return nil
	} else if err != nil {
		l.emitError(fmt.Errorf("lexString fail: %s", err.Error()))
		return nil
	}
	if r == '"' {
		if err := l.emitString(); err != nil {
			l.emitError(fmt.Errorf("emitString fail: %s", err.Error()))
			return nil
		}
		return l.lexAny
	} else if r == '\\' {
		return l.lexEscapeChar
	}
	return l.lexString
}

func (l *lexer) lexEscapeChar() stateFunc {
	r, err := l.read()
	if err == io.EOF {
		l.emitError(fmt.Errorf("unfinished string at line %d column %d", l.line, l.col))
		return nil
	} else if err != nil {
		l.emitError(fmt.Errorf("lexEscapeChar fail: %s", err.Error()))
		return nil
	}
	if r == 'a' || r == 'b' || r == 'f' || r == 'n' || r == 'r' || r == 't' || r == 'v' ||
		r == '\\' || r == '\'' || r == '"' {
		return l.lexString
	}
	l.emitError(fmt.Errorf("unknown escaped char at line %d column %d: '\\%c'", l.line, l.col, r))
	return nil
}

func (l *lexer) lexRune() stateFunc {
	r, err := l.read()
	if err == io.EOF {
		l.emitError(fmt.Errorf("unfinished rune at line %d column %d", l.line, l.col))
		return nil
	} else if err != nil {
		l.emitError(fmt.Errorf("lexRune fail: %s", err.Error()))
		return nil
	}
	if r == '\'' {
		l.emitError(fmt.Errorf("empty rune at line %d column %d", l.line, l.col))
		return nil
	} else if r == '\\' {
		return l.lexEscapeRune
	} else if r == '\r' || r == '\n' {
		l.emitError(fmt.Errorf("new line in rune at line %d column %d", l.line, l.col))
		return nil
	}
	return l.lexRuneEnd
}

func (l *lexer) lexEscapeRune() stateFunc {
	r, err := l.read()
	if err == io.EOF {
		l.emitError(fmt.Errorf("unfinished rune at line %d column %d", l.line, l.col))
		return nil
	} else if err != nil {
		l.emitError(fmt.Errorf("lexEscapeRune fail: %s", err.Error()))
		return nil
	}
	if r == 'a' || r == 'b' || r == 'f' || r == 'n' || r == 'r' || r == 't' || r == 'v' ||
		r == '\\' || r == '\'' || r == '"' {
		return l.lexRuneEnd
	}
	l.emitError(fmt.Errorf("unknown escaped char at line %d column %d: '\\%c'", l.line, l.col, r))
	return nil
}

func (l *lexer) lexRuneEnd() stateFunc {
	r, err := l.read()
	if err == io.EOF {
		l.emitError(fmt.Errorf("unfinished rune at line %d column %d", l.line, l.col))
		return nil
	} else if err != nil {
		l.emitError(fmt.Errorf("lexRuneEnd fail: %s", err.Error()))
		return nil
	}
	if r == '\'' {
		l.emitRune()
		return l.lexAny
	}
	l.emitError(fmt.Errorf("unfinished rune at line %d column %d", l.line, l.col))
	return nil
}

func (l *lexer) lexIndent() stateFunc {
	r, err := l.read()
	if err == io.EOF {
		l.emitIndents(l.pos - l.start)
		return nil
	} else if err != nil {
		l.emitError(fmt.Errorf("lexIndent fail: %s", err.Error()))
		return nil
	}
	if r == '.' {
		return l.lexIndent
	}
	l.unread()
	l.emitIndents(l.pos - l.start)
	return l.lexAny
}

func (l *lexer) lexZero() stateFunc {
	r, err := l.read()
	if err == io.EOF {
		l.emitWord(tokenDec)
		return nil
	} else if err != nil {
		l.emitError(fmt.Errorf("lexZero fail: %s", err.Error()))
		return nil
	}
	if unicode.IsDigit(r) {
		l.unread()
		return l.lexOct
	}
	if r == 'x' {
		return l.lexHexPrefix
	}
	l.unread()
	l.emitWord(tokenDec)
	return l.lexAny
}

func (l *lexer) lexOct() stateFunc {
	r, err := l.read()
	if err == io.EOF {
		l.emitWord(tokenOct)
		return nil
	} else if err != nil {
		l.emitError(fmt.Errorf("lexOct fail: %s", err.Error()))
		return nil
	}
	if unicode.IsDigit(r) {
		if r >= '0' && r <= '7' {
			return l.lexOct
		} else {
			l.emitError(fmt.Errorf("invalid oct int at line %d column %d", l.line, l.col))
			return nil
		}
	}
	l.unread()
	l.emitWord(tokenOct)
	return l.lexAny
}

func (l *lexer) lexHexPrefix() stateFunc {
	r, err := l.read()
	if err == io.EOF {
		l.emitError(fmt.Errorf("invalid hex number at line %d column %d: %s",
			l.line, l.col, l.text[l.start:l.pos]))
		return nil
	} else if err != nil {
		l.emitError(fmt.Errorf("lexHexPrefix fail: %s", err.Error()))
		return nil
	}
	if unicode.IsDigit(r) || r >= 'A' && r <= 'F' || r >= 'a' && r <= 'f' {
		return l.lexHex
	}
	l.emitError(fmt.Errorf("invalid hex number at line %d column %d: %s",
		l.line, l.col, l.text[l.start:l.pos]))
	return nil
}

func (l *lexer) lexHex() stateFunc {
	r, err := l.read()
	if err == io.EOF {
		l.emitWord(tokenHex)
		return nil
	} else if err != nil {
		l.emitError(fmt.Errorf("lexHex fail: %s", err.Error()))
		return nil
	}
	if unicode.IsDigit(r) || r >= 'A' && r <= 'F' || r >= 'a' && r <= 'f' {
		return l.lexHex
	}
	l.unread()
	l.emitWord(tokenHex)
	return l.lexAny
}

func (l *lexer) lexReserved() stateFunc {
	r, err := l.read()
	if err == io.EOF {
		text := string(l.text[l.start:l.pos])
		if text == "true" || text == "false" {
			l.emitWord(tokenBool)
			return l.lexAny
		}
		l.emitError(fmt.Errorf("invalid reserved word at line %d column %d: %s",
			l.line, l.col, text))
		return nil
	} else if err != nil {
		l.emitError(fmt.Errorf("lexReserved fail: %s", err.Error()))
		return nil
	}
	if unicode.IsLower(r) {
		return l.lexReserved
	}
	l.unread()
	text := string(l.text[l.start:l.pos])
	switch text {
	case "true", "false":
		l.emitWord(tokenBool)
		return l.lexAny
	case "import":
		l.emitLexer(tokenImport)
		return l.lexAny
	case "time":
		l.emitLexer(tokenTime)
		return l.lexAny
	case "duration":
		l.emitLexer(tokenDuration)
		return l.lexAny
	}
	l.emitError(fmt.Errorf("invalid reserved word at line %d column %d: %s",
		l.line, l.col, text))
	return nil
}

func (l *lexer) lexIdentifier() stateFunc {
	r, err := l.read()
	if err == io.EOF {
		l.emitWord(tokenIdentifier)
		return l.lexAny
	} else if err != nil {
		l.emitError(fmt.Errorf("lexIdentifier fail: %s", err.Error()))
		return nil
	}
	if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' {
		return l.lexIdentifier
	}
	l.unread()
	l.emitWord(tokenIdentifier)
	return l.lexAny
}
