package jaml

import (
	"errors"
	"fmt"
	"reflect"
	"runtime"
	"strconv"
	"testing"
	"time"
)

var (
	defaultTimeout = time.Millisecond * 50
)

func fetchToken(tokens <-chan token, timeout time.Duration, f func()) (tok token, err error) {
	if f != nil {
		go f()
	}
	select {
	case tok = <-tokens:
		break
	case <-time.After(timeout):
		err = fmt.Errorf("timeout: %s", timeout)
	}
	return
}

func fetchTokenOrError(tokens <-chan token, timeout time.Duration, errFunc func() error) (tok token, err error) {
	errCh := make(chan error)
	go func() {
		errCh <- errFunc()
	}()
	if err = <-errCh; err != nil {
		return
	}
	select {
	case tok = <-tokens:
		break
	case <-time.After(timeout):
		err = fmt.Errorf("timeout: %s", timeout)
	}
	return
}

func equalStateFunc(f1, f2 stateFunc) bool {
	return reflect.ValueOf(f1).Pointer() == reflect.ValueOf(f2).Pointer()
}

func funcName(sf stateFunc) string {
	if sf == nil {
		return "nil"
	}
	return runtime.FuncForPC(reflect.ValueOf(sf).Pointer()).Name()
}

func Test_newLexer(t *testing.T) {
	l := newLexer("", nil)
	if l == nil || l.start != 0 || l.pos != 0 || l.line != 1 || l.col != 1 {
		t.Error("create lexer fail")
	} else {
		t.Log("lexer creation test pass")
	}
}

func Test_lexerDebug(t *testing.T) {
	debugChan := make(chan string, 1)
	l := newLexer("", debugChan)
	go l.debug("test")
	select {
	case text := <-debugChan:
		if text != "test" {
			t.Error("lexer debug test fail: %q != %q", text, "test")
		} else {
			t.Log("lexer debug test pass")
		}
	case <-time.After(defaultTimeout):
		t.Error("lexer debug test timeout: %s", defaultTimeout)
	}
}

func Test_readAndUnread(t *testing.T) {
	l := newSilentLexer("test", nil)
	if l.read(); l.pos != 1 {
		t.Error("read fail: pos=%d", l.pos)
	} else if l.unread(); l.pos != 0 {
		t.Error("unread fail: pos=%d", l.pos)
	} else {
		t.Log("read and unread test pass")
	}
}

func Test_emitChars(t *testing.T) {
	l := newSilentLexer(",+", nil)
	l.pos = 2
	if token, err := fetchToken(l.tokens, defaultTimeout, l.emitChars); err != nil {
		t.Errorf("emit chars error: %s", err.Error())
		return
	} else if token.typ != tokenChar || token.run != ',' || token.lxr != l {
		t.Errorf("emit chars wrong: %s", token)
		return
	}
	if token, err := fetchToken(l.tokens, defaultTimeout, l.emitChars); err != nil {
		t.Errorf("emit chars error: %s", err.Error())
		return
	} else if token.typ != tokenChar || token.run != '+' || token.lxr != l {
		t.Errorf("emit chars wrong: %s", token)
		return
	}
	t.Log("chars emition test pass")
}

func Test_emitRune(t *testing.T) {
	l := newSilentLexer("'\\n'''", nil)
	// 正常Rune测试
	l.pos = 4
	if token, err := fetchTokenOrError(l.tokens, defaultTimeout, l.emitRune); err != nil {
		t.Errorf("emit rune error: %s", err.Error())
		return
	} else if token.typ != tokenRune || token.run != '\n' || token.lxr != l || l.line != 1 || l.col != 5 {
		t.Errorf("emit rune wrong: %s (%d:%d)", token, l.line, l.col)
		return
	}
	// 错误Rune测试
	l.pos = 6
	if _, err := fetchTokenOrError(l.tokens, defaultTimeout, l.emitRune); err == nil {
		t.Errorf("emit invalid rune success: ''")
		return
	}
	t.Log("rune emition test pass")
}

func Test_emitSymbol(t *testing.T) {
	// 非换行符号测试
	l := newSilentLexer("[\n", nil)
	l.pos = 1
	if token, err := fetchToken(l.tokens, defaultTimeout, func() { l.emitSymbol(tokenSquareOpen) }); err != nil {
		t.Errorf("emit symbol SQUARE_OPEN error: %s", err.Error())
		return
	} else if token.typ != tokenSquareOpen || token.lxr != l || l.line != 1 || l.col != 2 {
		t.Errorf("emit symbol SQUARE_OPEN wrong: %s (%d:%d)", token, l.line, l.col)
		return
	}
	// 换行符号测试
	l.pos = 2
	if token, err := fetchToken(l.tokens, defaultTimeout, func() { l.emitSymbol(tokenNewline) }); err != nil {
		t.Errorf("emit symbol NEWLINE error: %s", err.Error())
		return
	} else if token.typ != tokenNewline || token.lxr != l || l.line != 2 || l.col != 1 {
		t.Errorf("emit symbol NEWLINE wrong: %s (%d:%d)", token, l.line, l.col)
		return
	}
	t.Log("emit symbol test pass")
}

func Test_emitString(t *testing.T) {
	// 成功调用测试
	l := newSilentLexer("\"\\\"string\\\"\"\"\\p\"\"\n\"", nil)
	l.pos = 12
	if token, err := fetchTokenOrError(l.tokens, defaultTimeout, l.emitString); err != nil {
		t.Errorf("emit string error: %s", err.Error())
		return
	} else if token.typ != tokenString || token.str != "\"string\"" || token.lxr != l || l.line != 1 || l.col != 13 {
		t.Errorf("emit string wrong: %q (%d:%d)", token.str, l.line, l.col)
		return
	}
	// 错误转义测试
	l.pos = 16
	if _, err := fetchTokenOrError(l.tokens, defaultTimeout, l.emitString); err == nil {
		t.Errorf("emit invalid string success: \"\\p\"")
		return
	}
	// 多行测试
	l = newSilentLexer("\"\n\"", nil)
	l.pos = 3
	if token, err := fetchTokenOrError(l.tokens, defaultTimeout, l.emitString); err != nil {
		t.Errorf("emit multi-line string error: %s (%s)", err.Error(), strconv.Quote(string(l.text[l.start:l.pos])))
		return
	} else if token.typ != tokenString || token.str != "\n" || token.lxr != l || l.line != 2 || l.col != 2 {
		t.Errorf("emit multi-line string wrong: %s (%d:%d)", token, l.line, l.col)
		return
	}
	t.Log("emit string test pass")
}

func Test_emitWord(t *testing.T) {
	l := newSilentLexer("ID", nil)
	l.pos = 2
	if token, err := fetchToken(l.tokens, defaultTimeout, func() { l.emitWord(tokenIdentifier) }); err != nil {
		t.Errorf("emit word error: %s", err.Error())
		return
	} else if token.typ != tokenIdentifier || token.str != "ID" || token.lxr != l || l.line != 1 || l.col != 3 {
		t.Errorf("emit word wrong: %s (%d:%d)", token, l.line, l.col)
		return
	}
	t.Log("emit word test pass")
}

func Test_emitIndentAndDedent(t *testing.T) {
	l := newSilentLexer("", nil)
	if token, err := fetchToken(l.tokens, defaultTimeout, func() { l.emitIndents(1) }); err != nil {
		t.Errorf("emit indent error: %s", err.Error())
		return
	} else if token.typ != tokenIndent {
		t.Errorf("emit indent wrong: %s", token)
	}
	if token, err := fetchToken(l.tokens, defaultTimeout, func() { l.emitIndents(2) }); err != nil {
		t.Errorf("emit indent error: %s", err.Error())
		return
	} else if token.typ != tokenIndent {
		t.Errorf("emit indent wrong: %s", token)
	}
	if _, err := fetchToken(l.tokens, defaultTimeout, func() { l.emitIndents(2) }); err == nil {
		t.Errorf("emit no indent fail")
		return
	}
	if token, err := fetchToken(l.tokens, defaultTimeout, func() { l.emitIndents(0) }); err != nil {
		t.Errorf("emit dedent error: %s", err.Error())
		return
	} else if token.typ != tokenDedent {
		t.Errorf("emit dedent wrong: %s", token)
		return
	}
	if token, err := fetchToken(l.tokens, defaultTimeout, nil); err != nil {
		t.Errorf("emit newline after dedent error: %s", err.Error())
		return
	} else if token.typ != tokenNewline {
		t.Errorf("emit newline after dedent wrong: %s", token)
		return
	}
	if token, err := fetchToken(l.tokens, defaultTimeout, nil); err != nil {
		t.Errorf("emit dedent 2 error: %s", err.Error())
		return
	} else if token.typ != tokenDedent {
		t.Errorf("emit dedent 2 wrong: %s", token)
		return
	}
	if token, err := fetchToken(l.tokens, defaultTimeout, nil); err != nil {
		t.Errorf("emit newline after dedent 2 error: %s", err.Error())
		return
	} else if token.typ != tokenNewline {
		t.Errorf("emit newline after dedent 2 wrong: %s", token)
		return
	}
	t.Log("emit indent and dedent test pass")
}

func Test_emitLexer(t *testing.T) {
	l := newSilentLexer("", nil)
	if token, err := fetchToken(l.tokens, defaultTimeout, func() { l.emitLexer(tokenImport) }); err != nil {
		t.Errorf("emit lexer error: %s", err.Error())
		return
	} else if token.typ != tokenImport || token.lxr != l {
		t.Errorf("emit lexer wrong: %s", token)
		return
	}
	t.Log("emit lexer test pass")
}

func Test_emitError(t *testing.T) {
	l := newSilentLexer("", nil)
	if token, err := fetchToken(l.tokens, defaultTimeout, func() { l.emitError(errors.New("error")) }); err != nil {
		t.Errorf("emit error error: %s", err.Error())
		return
	} else if token.typ != tokenError || token.lxr != l {
		t.Errorf("emit error wrong: %s", token)
		return
	}
	t.Log("emit error test pass")
}

func Test_lexAny(t *testing.T) {
	var l *lexer
	// EOF
	l = newSilentLexer("", nil)
	if sf := l.lexAny(); sf != nil {
		t.Errorf("lex any with eof fail: %s", funcName(sf))
		return
	}
	// read error
	l = newSilentLexer("", nil)
	l.err = errors.New("error")
	if sf := l.lexAny(); sf != nil {
		t.Errorf("lex any with error fail: %s", funcName(sf))
		return
	} else if token, err := fetchToken(l.tokens, defaultTimeout, nil); err != nil {
		t.Errorf("lex any with error error: %s", err.Error())
		return
	} else if token.typ != tokenError || token.err != l.err {
		t.Errorf("lex any with error wrong: %s", token)
		return
	}
	// minus
	l = newSilentLexer("-", nil)
	if sf := l.lexAny(); !equalStateFunc(sf, l.lexMinus) {
		t.Errorf("lex any with minus fail: %s", funcName(sf))
		return
	}
	// quote
	l = newSilentLexer("\"", nil)
	if sf := l.lexAny(); !equalStateFunc(sf, l.lexString) {
		t.Errorf("lex any with quote fail: %s", funcName(sf))
		return
	}
	// single-quote
	l = newSilentLexer("'", nil)
	if sf := l.lexAny(); !equalStateFunc(sf, l.lexRune) {
		t.Errorf("lex any with single-quote fail: %s", funcName(sf))
		return
	}
	// space
	l = newSilentLexer(" ", nil)
	if sf := l.lexAny(); !equalStateFunc(sf, l.lexAny) {
		t.Errorf("lex any with space fail: %s", funcName(sf))
		return
	} else if l.start != l.pos || l.pos != 1 {
		t.Errorf("lex any with space wrong: [%d:%d]", l.start, l.pos)
		return
	}
	// tabular
	l = newSilentLexer("\t", nil)
	if sf := l.lexAny(); !equalStateFunc(sf, l.lexAny) {
		t.Errorf("lex any with tabular fail: %s", funcName(sf))
		return
	} else if l.start != l.pos || l.pos != 1 {
		t.Errorf("lex any with tabular wrong: [%d:%d]", l.start, l.pos)
		return
	}
	// newline in array
	l = newSilentLexer("\n", nil)
	l.arrlv = 1
	if sf := l.lexAny(); !equalStateFunc(sf, l.lexAny) {
		t.Errorf("lex any with newline in array fail: %s", funcName(sf))
		return
	} else if l.start != l.pos || l.pos != 1 {
		t.Errorf("lex any with newline in array wrong: [%d:%d]", l.start, l.pos)
		return
	}
	// newline out array
	l = newSilentLexer("\n", nil)
	l.arrlv = 0
	if sf := l.lexAny(); !equalStateFunc(sf, l.lexIndent) {
		t.Errorf("lex any with newline out array fail: %s", funcName(sf))
		return
	} else if token, err := fetchToken(l.tokens, defaultTimeout, nil); err != nil {
		t.Errorf("lex any with newline out array error: %s", err.Error())
		return
	} else if token.typ != tokenNewline {
		t.Errorf("lex any with newline out array wrong: %s", token)
		return
	}
	// zero
	l = newSilentLexer("0", nil)
	if sf := l.lexAny(); !equalStateFunc(sf, l.lexZero) {
		t.Errorf("lex any with zero fail: %s", funcName(sf))
		return
	}
	// non-zero digit
	l = newSilentLexer("1", nil)
	if sf := l.lexAny(); !equalStateFunc(sf, l.lexInteger) {
		t.Errorf("lex any with non-zero digit fail: %s", funcName(sf))
		return
	}
	// open square
	l = newSilentLexer("[", nil)
	if sf := l.lexAny(); !equalStateFunc(sf, l.lexAny) {
		t.Errorf("lex any with open square fail: %s", funcName(sf))
		return
	} else if token, err := fetchToken(l.tokens, defaultTimeout, nil); err != nil {
		t.Errorf("lex any with open square error: %s", err.Error())
		return
	} else if token.typ != tokenSquareOpen {
		t.Errorf("lex any with open square wrong: %s", token)
		return
	} else if l.arrlv != 1 {
		t.Errorf("lex any with open square wrong: arrlv=%d", l.arrlv)
		return
	}
	// close square
	l = newSilentLexer("]", nil)
	l.arrlv = 2
	if sf := l.lexAny(); !equalStateFunc(sf, l.lexAny) {
		t.Errorf("lex any with close square fail: %s", funcName(sf))
		return
	} else if token, err := fetchToken(l.tokens, defaultTimeout, nil); err != nil {
		t.Errorf("lex any with close square error: %s", err.Error())
		return
	} else if token.typ != tokenSquareClose {
		t.Errorf("lex any with close square wrong: %s", token)
		return
	} else if l.arrlv != 1 {
		t.Errorf("lex any with close square wrong: arrlv=%d", l.arrlv)
		return
	}
	// lower letter
	l = newSilentLexer("a", nil)
	if sf := l.lexAny(); !equalStateFunc(sf, l.lexReserved) {
		t.Errorf("lex any with lower letter fail: %s", funcName(sf))
		return
	}
	// upper letter
	l = newSilentLexer("A", nil)
	if sf := l.lexAny(); !equalStateFunc(sf, l.lexIdentifier) {
		t.Errorf("lex any with upper letter fail: %s", funcName(sf))
		return
	}
	// other char
	l = newSilentLexer(",", nil)
	if sf := l.lexAny(); !equalStateFunc(sf, l.lexAny) {
		t.Errorf("lex any with other char fail: %s", funcName(sf))
		return
	} else if token, err := fetchToken(l.tokens, defaultTimeout, nil); err != nil {
		t.Errorf("lex any with other char error: %s", err.Error())
		return
	} else if token.typ != tokenChar || token.run != ',' {
		t.Errorf("lex any with other char wrong: %s", token)
		return
	}
	t.Log("lex any test pass")
}

func Test_lexMinus(t *testing.T) {
	var l *lexer
	// digit
	l = newSilentLexer("-1", nil)
	l.pos = 1
	if sf := l.lexMinus(); !equalStateFunc(sf, l.lexInteger) {
		t.Errorf("lex minus with integer fail: %s", funcName(sf))
		return
	}
	// EOF
	l = newSilentLexer("-", nil)
	l.pos = 1
	if sf := l.lexMinus(); sf != nil {
		t.Errorf("lex minus with eof fail: %s", funcName(sf))
		return
	} else if token, err := fetchToken(l.tokens, defaultTimeout, nil); err != nil {
		t.Errorf("lex minus with eof error: %s", err.Error())
		return
	} else if token.typ != tokenChar || token.run != '-' {
		t.Errorf("lex minus with eof wrong: %s", token)
		return
	}
	// read error
	l = newSilentLexer("-", nil)
	l.err = errors.New("error")
	l.pos = 1
	if sf := l.lexMinus(); sf != nil {
		t.Errorf("lex minus with eof fail: %s", funcName(sf))
		return
	} else if token, err := fetchToken(l.tokens, defaultTimeout, nil); err != nil {
		t.Errorf("lex minus with error error: %s", err.Error())
		return
	} else if token.typ != tokenError || token.err != l.err {
		t.Errorf("lex minus with error wrong: %s", token)
		return
	}
	// non-digit
	l = newSilentLexer("-x", nil)
	l.pos = 1
	if sf := l.lexMinus(); !equalStateFunc(sf, l.lexAny) {
		t.Errorf("lex minus with non-digit fail: %s", funcName(sf))
		return
	} else if token, err := fetchToken(l.tokens, defaultTimeout, nil); err != nil {
		t.Errorf("lex minus with non-digit error: %s", err.Error())
		return
	} else if token.typ != tokenChar || token.run != '-' {
		t.Errorf("lex minus with non-digit wrong: %s", token)
		return
	}
	t.Log("lex minus test pass")
}

func Test_lexInteger(t *testing.T) {
	var l *lexer
	// digit
	l = newSilentLexer("10", nil)
	l.pos = 1
	if sf := l.lexInteger(); !equalStateFunc(sf, l.lexInteger) {
		t.Errorf("lex integer with digit fail: %s", funcName(sf))
		return
	}
	// EOF
	l = newSilentLexer("1", nil)
	l.pos = 1
	if sf := l.lexInteger(); sf != nil {
		t.Errorf("lex integer with eof fail: %s", funcName(sf))
		return
	} else if token, err := fetchToken(l.tokens, defaultTimeout, nil); err != nil {
		t.Errorf("lex integer with eof error: %s", err.Error())
		return
	} else if token.typ != tokenDec || token.str != "1" {
		t.Errorf("lex integer with eof wrong: %s", token)
		return
	}
	// read error
	l = newSilentLexer("10", nil)
	l.pos = 1
	l.err = errors.New("error")
	if sf := l.lexInteger(); sf != nil {
		t.Errorf("lex integer with error fail: %s", funcName(sf))
		return
	} else if token, err := fetchToken(l.tokens, defaultTimeout, nil); err != nil {
		t.Errorf("lex integer with error error: %s", err.Error())
		return
	} else if token.typ != tokenError {
		t.Errorf("lex integer with error wrong: %s", token)
		return
	}
	// float point
	l = newSilentLexer("1.0", nil)
	l.pos = 1
	if sf := l.lexInteger(); !equalStateFunc(sf, l.lexFloatPoint) {
		t.Errorf("lex integer with float point fail: %s", funcName(sf))
		return
	}
	// non-digit
	l = newSilentLexer("1 ", nil)
	l.pos = 1
	if sf := l.lexInteger(); !equalStateFunc(sf, l.lexAny) {
		t.Errorf("lex integer with non-digit fail: %s", funcName(sf))
		return
	} else if token, err := fetchToken(l.tokens, defaultTimeout, nil); err != nil {
		t.Errorf("lex integer with non-digit error: %s", err.Error())
		return
	} else if token.typ != tokenDec || token.str != "1" {
		t.Errorf("lex integer with non-digit wrong: %s", token)
		return
	}
	t.Log("lex integer test pass")
}

func Test_lexFloatPoint(t *testing.T) {
	var l *lexer
	// digit
	l = newSilentLexer("1.0", nil)
	l.pos = 2
	if sf := l.lexFloatPoint(); !equalStateFunc(sf, l.lexFloat) {
		t.Errorf("lex float point with digit fail: %s", funcName(sf))
		return
	}
	// EOF
	l = newSilentLexer("1.", nil)
	l.pos = 2
	if sf := l.lexFloatPoint(); sf != nil {
		t.Errorf("lex float point with eof fail: %s", funcName(sf))
		return
	} else if token, err := fetchToken(l.tokens, defaultTimeout, nil); err != nil {
		t.Errorf("lex float point with eof error: %s", err.Error())
		return
	} else if token.typ != tokenError {
		t.Errorf("lex float point with eof wrong: %s", token)
		return
	}
	// read error
	l = newSilentLexer("1.0", nil)
	l.pos = 2
	l.err = errors.New("error")
	if sf := l.lexFloatPoint(); sf != nil {
		t.Errorf("lex float point with error fail: %s", funcName(sf))
		return
	} else if token, err := fetchToken(l.tokens, defaultTimeout, nil); err != nil {
		t.Errorf("lex float point with error error: %s", err.Error())
		return
	} else if token.typ != tokenError || token.err != l.err {
		t.Errorf("lex float point with error wrong: %s", token)
		return
	}
	// non-digit
	l = newSilentLexer("1.x", nil)
	l.pos = 2
	if sf := l.lexFloatPoint(); sf != nil {
		t.Errorf("lex float point with non-digit fail: %s", funcName(sf))
		return
	} else if token, err := fetchToken(l.tokens, defaultTimeout, nil); err != nil {
		t.Errorf("lex float point with non-digit error: %s", err.Error())
		return
	} else if token.typ != tokenError {
		t.Errorf("lex float point with non-digit wrong: %s", token)
		return
	}
	t.Log("lex float point test pass")
}

func Test_lexFloat(t *testing.T) {
	var l *lexer
	// digit
	l = newSilentLexer("1.0", nil)
	l.pos = 2
	if sf := l.lexFloat(); !equalStateFunc(sf, l.lexFloat) {
		t.Errorf("lex float with digit fail: %s", funcName(sf))
		return
	}
	// EOF
	l = newSilentLexer("1.0", nil)
	l.pos = 3
	if sf := l.lexFloat(); sf != nil {
		t.Errorf("lex float with eof fail: %s", funcName(sf))
		return
	} else if token, err := fetchToken(l.tokens, defaultTimeout, nil); err != nil {
		t.Errorf("lex float with eof error: %s", err.Error())
		return
	} else if token.typ != tokenFloat || token.str != "1.0" {
		t.Errorf("lex float with eof wrong: %s", token)
		return
	}
	// read error
	l = newSilentLexer("1.00", nil)
	l.pos = 3
	l.err = errors.New("error")
	if sf := l.lexFloat(); sf != nil {
		t.Errorf("lex float with error fail: %s", funcName(sf))
		return
	} else if token, err := fetchToken(l.tokens, defaultTimeout, nil); err != nil {
		t.Errorf("lex float with error error: %s", err.Error())
		return
	} else if token.typ != tokenError || token.err != l.err {
		t.Errorf("lex float with error wrong: %s", token)
		return
	}
	// non-digit
	l = newSilentLexer("1.0 ", nil)
	l.pos = 3
	if sf := l.lexFloat(); !equalStateFunc(sf, l.lexAny) {
		t.Errorf("lex float with non-digit fail: %s", funcName(sf))
		return
	} else if token, err := fetchToken(l.tokens, defaultTimeout, nil); err != nil {
		t.Errorf("lex float with non-digit error: %s", err.Error())
		return
	} else if token.typ != tokenFloat || token.str != "1.0" {
		t.Errorf("lex float with non-digit wrong: %s", token)
		return
	}
	t.Log("lex float test pass")
}

func Test_lexString(t *testing.T) {
	var l *lexer
	// non-quote
	l = newSilentLexer("\"a\"", nil)
	l.pos = 1
	if sf := l.lexString(); !equalStateFunc(sf, l.lexString) {
		t.Errorf("lex string with non-quote fail: %s", funcName(sf))
		return
	}
	// EOF
	l = newSilentLexer("\"a", nil)
	l.pos = 2
	if sf := l.lexString(); sf != nil {
		t.Errorf("lex string with eof fail: %s", funcName(sf))
		return
	} else if token, err := fetchToken(l.tokens, defaultTimeout, nil); err != nil {
		t.Errorf("lex string with eof error: %s", err.Error())
		return
	} else if token.typ != tokenError {
		t.Errorf("lex string with eof wrong: %s", token)
		return
	}
	// read error
	l = newSilentLexer("\"a\"", nil)
	l.pos = 2
	l.err = errors.New("error")
	if sf := l.lexString(); sf != nil {
		t.Errorf("lex string with error fail: %s", funcName(sf))
		return
	} else if token, err := fetchToken(l.tokens, defaultTimeout, nil); err != nil {
		t.Errorf("lex string with error error: %s", err.Error())
		return
	} else if token.typ != tokenError || token.err != l.err {
		t.Errorf("lex string with error wrong: %s", token)
		return
	}
	// escape char
	l = newSilentLexer("\"\\\\\"", nil)
	l.pos = 1
	if sf := l.lexString(); !equalStateFunc(sf, l.lexEscapeChar) {
		t.Errorf("lex string with escape char fail: %s", funcName(sf))
		return
	}
	// quote
	l = newSilentLexer("\"a\"", nil)
	l.pos = 2
	if sf := l.lexString(); !equalStateFunc(sf, l.lexAny) {
		t.Errorf("lex string with quote fail: %s", funcName(sf))
		return
	} else if token, err := fetchToken(l.tokens, defaultTimeout, nil); err != nil {
		t.Errorf("lex string with quote error: %s", err.Error())
		return
	} else if token.typ != tokenString || token.str != "a" {
		t.Errorf("lex string with quote wrong: %s", token)
	}
	// invalid quote
	l = newSilentLexer("\"a\\\"", nil)
	l.pos = 3
	if sf := l.lexString(); sf != nil {
		t.Errorf("lex string with invalid quote fail: %s", funcName(sf))
		return
	} else if token, err := fetchToken(l.tokens, defaultTimeout, nil); err != nil {
		t.Errorf("lex string with invalid quote error: %s", err.Error())
		return
	} else if token.typ != tokenError {
		t.Errorf("lex string with invalid quote wrong: %s", token)
		return
	}
	t.Log("lex string test pass")
}

func Test_lexEscapeChar(t *testing.T) {
	var l *lexer
	// valid char
	l = newSilentLexer("\"\\n\"", nil)
	l.pos = 2
	if sf := l.lexEscapeChar(); !equalStateFunc(sf, l.lexString) {
		t.Errorf("lex escape char fail: %s", funcName(sf))
		return
	}
	// EOF
	l = newSilentLexer("\"\\", nil)
	l.pos = 2
	if sf := l.lexEscapeChar(); sf != nil {
		t.Errorf("lex escape char with eof fail: %s", funcName(sf))
		return
	} else if token, err := fetchToken(l.tokens, defaultTimeout, nil); err != nil {
		t.Errorf("lex escape char with eof error: %s", err.Error())
		return
	} else if token.typ != tokenError {
		t.Errorf("lex escape char with eof wrong: %s", token)
		return
	}
	// read error
	l = newSilentLexer("\"\\n\"", nil)
	l.pos = 2
	l.err = errors.New("error")
	if sf := l.lexEscapeChar(); sf != nil {
		t.Errorf("lex escape char with error fail: %s", funcName(sf))
		return
	} else if token, err := fetchToken(l.tokens, defaultTimeout, nil); err != nil {
		t.Errorf("lex escape char with error error: %s", err.Error())
		return
	} else if token.typ != tokenError || token.err != l.err {
		t.Errorf("lex escape char with error wrong: %s", token)
		return
	}
	// invalid char
	l = newSilentLexer("\"\\c\"", nil)
	l.pos = 2
	if sf := l.lexEscapeChar(); sf != nil {
		t.Errorf("lex escape char with invalid char fail: %s", funcName(sf))
		return
	} else if token, err := fetchToken(l.tokens, defaultTimeout, nil); err != nil {
		t.Errorf("lex escape char with invalid char error: %s", err.Error())
		return
	} else if token.typ != tokenError {
		t.Errorf("lex escape char with invalid char wrong: %s", token)
		return
	}
	t.Log("lex escape char test pass")
}

func Test_lexRune(t *testing.T) {
	var l *lexer
	// normal char
	l = newSilentLexer("'a'", nil)
	l.pos = 1
	if sf := l.lexRune(); !equalStateFunc(sf, l.lexRuneEnd) {
		t.Errorf("lex rune with normal char fail: %s", funcName(sf))
		return
	}
	// single-quote
	l = newSilentLexer("''", nil)
	l.pos = 1
	if sf := l.lexRune(); sf != nil {
		t.Errorf("lex rune with eof fail: %s", funcName(sf))
		return
	} else if token, err := fetchToken(l.tokens, defaultTimeout, nil); err != nil {
		t.Errorf("lex rune with eof error: %s", err.Error())
		return
	} else if token.typ != tokenError {
		t.Errorf("lex rune with eof wrong: %s", token)
		return
	}
	// EOF
	l = newSilentLexer("'", nil)
	l.pos = 1
	if sf := l.lexRune(); sf != nil {
		t.Errorf("lex rune with eof fail: %s", funcName(sf))
		return
	} else if token, err := fetchToken(l.tokens, defaultTimeout, nil); err != nil {
		t.Errorf("lex rune with eof error: %s", err.Error())
		return
	} else if token.typ != tokenError {
		t.Errorf("lex rune with eof wrong: %s", token)
		return
	}
	// read error
	l = newSilentLexer("'a'", nil)
	l.pos = 1
	l.err = errors.New("error")
	if sf := l.lexRune(); sf != nil {
		t.Errorf("lex rune with error fail: %s", funcName(sf))
		return
	} else if token, err := fetchToken(l.tokens, defaultTimeout, nil); err != nil {
		t.Errorf("lex rune with error error: %s", err.Error())
		return
	} else if token.typ != tokenError || token.err != l.err {
		t.Errorf("lex rune with error wrong: %s", token)
		return
	}
	// escape char
	l = newSilentLexer("'\\n'", nil)
	l.pos = 1
	if sf := l.lexRune(); !equalStateFunc(sf, l.lexEscapeRune) {
		t.Errorf("lex rune with escape char fail: %s", funcName(sf))
		return
	}
	// newline
	l = newSilentLexer("'\n'", nil)
	l.pos = 1
	if sf := l.lexRune(); sf != nil {
		t.Errorf("lex rune with newline fail: %s", funcName(sf))
		return
	} else if token, err := fetchToken(l.tokens, defaultTimeout, nil); err != nil {
		t.Errorf("lex rune with newline error: %s", err.Error())
		return
	} else if token.typ != tokenError {
		t.Errorf("lex rune with newline wrong: %s", token)
		return
	}
	// return
	l = newSilentLexer("'\r'", nil)
	l.pos = 1
	if sf := l.lexRune(); sf != nil {
		t.Errorf("lex rune with return fail: %s", funcName(sf))
		return
	} else if token, err := fetchToken(l.tokens, defaultTimeout, nil); err != nil {
		t.Errorf("lex rune with return error: %s", err.Error())
		return
	} else if token.typ != tokenError {
		t.Errorf("lex rune with return wrong: %s", token)
		return
	}
	t.Log("lex rune test pass")
}

func Test_lexEscapeRune(t *testing.T) {
	var l *lexer
	// valid char
	l = newSilentLexer("\"\\n\"", nil)
	l.pos = 2
	if sf := l.lexEscapeRune(); !equalStateFunc(sf, l.lexRuneEnd) {
		t.Errorf("lex escape rune fail: %s", funcName(sf))
		return
	}
	// EOF
	l = newSilentLexer("\"\\", nil)
	l.pos = 2
	if sf := l.lexEscapeRune(); sf != nil {
		t.Errorf("lex escape rune with eof fail: %s", funcName(sf))
		return
	} else if token, err := fetchToken(l.tokens, defaultTimeout, nil); err != nil {
		t.Errorf("lex escape rune with eof error: %s", err.Error())
		return
	} else if token.typ != tokenError {
		t.Errorf("lex escape rune with eof wrong: %s", token)
		return
	}
	// read error
	l = newSilentLexer("\"\\n\"", nil)
	l.pos = 2
	l.err = errors.New("error")
	if sf := l.lexEscapeRune(); sf != nil {
		t.Errorf("lex escape rune with error fail: %s", funcName(sf))
		return
	} else if token, err := fetchToken(l.tokens, defaultTimeout, nil); err != nil {
		t.Errorf("lex escape rune with error error: %s", err.Error())
		return
	} else if token.typ != tokenError || token.err != l.err {
		t.Errorf("lex escape rune with error wrong: %s", token)
		return
	}
	// invalid char
	l = newSilentLexer("\"\\c\"", nil)
	l.pos = 2
	if sf := l.lexEscapeRune(); sf != nil {
		t.Errorf("lex escape rune with invalid char fail: %s", funcName(sf))
		return
	} else if token, err := fetchToken(l.tokens, defaultTimeout, nil); err != nil {
		t.Errorf("lex escape rune with invalid char error: %s", err.Error())
		return
	} else if token.typ != tokenError {
		t.Errorf("lex escape rune with invalid char wrong: %s", token)
		return
	}
	t.Log("lex escape rune test pass")
}

func Test_lexRuneEnd(t *testing.T) {
	var l *lexer
	// single quote
	l = newSilentLexer("'a'", nil)
	l.pos = 2
	if sf := l.lexRuneEnd(); !equalStateFunc(sf, l.lexAny) {
		t.Errorf("lex rune end with single-quote fail: %s", funcName(sf))
		return
	} else if token, err := fetchToken(l.tokens, defaultTimeout, nil); err != nil {
		t.Errorf("lex rune end with single-quote error: %s", err.Error())
		return
	} else if token.typ != tokenRune || token.run != 'a' {
		t.Errorf("lex rune end with single-quote wrong: %s", token)
		return
	}
	// EOF
	l = newSilentLexer("'a", nil)
	l.pos = 2
	if sf := l.lexRuneEnd(); sf != nil {
		t.Errorf("lex rune end with eof fail: %s", funcName(sf))
		return
	} else if token, err := fetchToken(l.tokens, defaultTimeout, nil); err != nil {
		t.Errorf("lex rune end with eof error: %s", err.Error())
		return
	} else if token.typ != tokenError {
		t.Errorf("lex rune end with eof wrong: %s", token)
		return
	}
	// read error
	l = newSilentLexer("'a'", nil)
	l.pos = 2
	l.err = errors.New("error")
	if sf := l.lexRuneEnd(); sf != nil {
		t.Errorf("lex rune end with error fail: %s", funcName(sf))
		return
	} else if token, err := fetchToken(l.tokens, defaultTimeout, nil); err != nil {
		t.Errorf("lex rune end with error error: %s", err.Error())
		return
	} else if token.typ != tokenError {
		t.Errorf("lex rune end with error wrong: %s", token)
		return
	}
	// other char
	l = newSilentLexer("'ab'", nil)
	l.pos = 2
	if sf := l.lexRuneEnd(); sf != nil {
		t.Errorf("lex rune end with other char fail: %s", funcName(sf))
		return
	} else if token, err := fetchToken(l.tokens, defaultTimeout, nil); err != nil {
		t.Errorf("lex rune end with other char error: %s", err.Error())
		return
	} else if token.typ != tokenError {
		t.Errorf("lex rune end with other char wrong: %s", token)
		return
	}
	t.Log("lex rune end test pass")
}

func Test_lexIndent(t *testing.T) {
	var l *lexer
	// dot
	l = newSilentLexer(".", nil)
	if sf := l.lexIndent(); !equalStateFunc(sf, l.lexIndent) {
		t.Errorf("lex indent with dot fail: %s", funcName(sf))
		return
	}
	// EOF
	l = newSilentLexer("", nil)
	l.lv = 1
	if sf := l.lexIndent(); sf != nil {
		t.Errorf("lex indent with eof fail: %s", funcName(sf))
		return
	} else if token, err := fetchToken(l.tokens, defaultTimeout, nil); err != nil {
		t.Errorf("lex indent with eof error: %s", err.Error())
		return
	} else if token.typ != tokenDedent {
		t.Errorf("lex indent with eof wrong: %s", token)
		return
	} else if l.lv != 0 {
		t.Errorf("lex indent with eof wrong: lv=%d", l.lv)
		return
	}
	// read error
	l = newSilentLexer("", nil)
	l.err = errors.New("error")
	if sf := l.lexIndent(); sf != nil {
		t.Errorf("lex indent with error fail: %s", funcName(sf))
		return
	} else if token, err := fetchToken(l.tokens, defaultTimeout, nil); err != nil {
		t.Errorf("lex indent with error error: %s", err.Error())
		return
	} else if token.typ != tokenError {
		t.Errorf("lex indent with error wrong: %s", token)
		return
	}
	// other char
	l = newSilentLexer("ab", nil)
	l.lv = 1
	if sf := l.lexIndent(); !equalStateFunc(sf, l.lexAny) {
		t.Errorf("lex indent with other char fail: %s", funcName(sf))
		return
	} else if token, err := fetchToken(l.tokens, defaultTimeout, nil); err != nil {
		t.Errorf("lex indent with other char error: %s", err.Error())
		return
	} else if token.typ != tokenDedent {
		t.Errorf("lex indent with other char wrong: %s", token)
		return
	} else if l.lv != 0 {
		t.Errorf("lex indent with eof wrong: lv=%d", l.lv)
		return
	}
	t.Log("lex indent test pass")
}
