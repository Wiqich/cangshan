%{
package jaml

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

%}

%union {
    val *Value
    fld fieldInfo
    str string
    pth fieldPath
    tkn token
    tag map[string]string
}

%type <val> root list object suite
%type <str> field_path_head class
%type <pth> field_name field_path
%type <val> assign non_object array_inside array_literal primary_value
%type <tag> tags tag_list
%type <fld> field

%token <tkn> DEC OCT HEX FLOAT STRING RUNE BOOL IDENTIFIER NEWLINE SQUARE_OPEN SQUARE_CLOSE INDENT DEDENT IMPORT TIME DURATION

%%

root            : list
                    {
                        $$ = $1
                        $$.lexer.root <- $$
                        $$.lexer.debug("yacc: root -> list")
                    }
                ;

list            : field
                    {
                        $$ = &Value{
                            Kind:   reflect.Struct,
                            Fields: make(map[string]*Value),
                            lexer:  $1.val.lexer,
                        }
                        if $1.imp != "" {
                            $$.Import($1.imp)
                        } else {
                            $$.SetField($1.pth, $1.val)
                        }
                        $$.lexer.debug("yacc: list -> field")
                    }
                | list NEWLINE field
                    {
                        $$ = $1
                        if $3.imp != "" {
                            $$.Import($3.imp)
                        } else {
                            $$.SetField($3.pth, $3.val)
                        }
                        $$.lexer.debug("yacc: list -> list NEWLINE field")
                    }
                | list NEWLINE
                    {
                        $$ = $1
                        $$.lexer.debug("yacc: list -> list NEWLINE")
                    }
                ;

field           : field_path assign
                    {
                        $$.pth = $1.pth
                        $$.val = $2
                        $$.lxr = $1.lxr
                        $$.lxr.debug("yacc: field -> field_path assign")
                    }
                | IMPORT STRING
                    {
                        $$.imp = $2.str
                        $$.lxr = $1.lxr
                        $$.lxr.debug("yacc: field -> IMPORT STRING : %s", $2)
                    }
                ;

field_path      : field_name
                    {
                        $$ = $1
                        $$.lxr.debug("yacc: field_path -> field_name")
                    }
                | field_path_head '.' field_name
                    {
                        $$ = $3
                        $$.pth = $1 + "." + $$.pth
                        $$.lxr.debug("yacc: field_path -> field_path_head '.' field_name")
                    }
                ;

field_path_head : IDENTIFIER
                    {
                        $$ = $1.str
                        $1.lxr.debug("yacc: field_path_head -> IDENTIFIER : %s", $1)
                    }
                | field_path_head '.' IDENTIFIER
                    {
                        $$ = $1 + "." + $3.str
                        $3.lxr.debug("yacc: field_path_head -> field_path_head '.' IDENTIFIER : %s", $3)
                    }
                ;

field_name      : IDENTIFIER
                    {
                        $$.pth = $1.str
                        $$.lxr = $1.lxr
                        $$.lxr.debug("yacc: field_name -> IDENTIFIER : %s", $1)
                    }
                | IDENTIFIER '[' primary_value ']'
                    {
                        $$.pth = $1.str
                        $$.key = $3
                        $$.lxr = $1.lxr
                        $$.lxr.debug("yacc: field_name -> IDENTIFIER '[' primary_value ']' : %s, %s", $1, $3)
                    }
                | IDENTIFIER '[' ']'
                    {
                        $$.pth = $1.str
                        $$.arr = true
                        $$.lxr = $1.lxr
                        $$.lxr.debug("yacc: field_name -> IDENTIFIER '[' ']' : %s", $1)
                    }
                ;

assign          : ':' non_object
                    {
                        $$ = $2
                        $$.lexer.debug("yacc: assign -> ':' non_object")
                    }
                | ':' class tags object
                    {
                        $$ = $4
                        $4.Type = $2
                        $4.Tags = $3
                        $$.lexer.debug("yacc: assign -> ':' object")
                    }
                | ':' '@' IDENTIFIER
                    {
                        $$ = &Value{
                            Kind:    reflect.Ptr,
                            Primary: $3.str,
                            lexer:   $3.lxr,
                        }
                        $$.lexer.debug("yacc: assign -> ':' '@' IDENTIFIER : %s", $3)
                    }
                ;

class           : /* empty */
                    {
                    }
                | '@' STRING
                    {
                        $$ = $2.str
                        $2.lxr.debug("yacc: assign -> '@' STRING : %s", $2)
                    }
                ;

tags            : /* empty */
                    {
                    }
                | '`' tag_list '`'
                    {
                        $$ = $2
                    }
                ;

tag_list        : tag_list ',' IDENTIFIER '=' STRING
                    {
                        $$ = $1
                        $1[$3.str] = $5.str
                        $3.lxr.debug("yacc: tag_list -> tag_list ',' IDENTIFIER '=' STRING : %s, %s", $3, $5)
                    }
                | IDENTIFIER '=' STRING
                    {
                        $$ = make(map[string]string)
                        $$[$1.str] = $3.str
                        $1.lxr.debug("yacc: tag_list -> IDENTIFIER '=' STRING : %s, %s", $1, $3)
                    }
                ;

primary_value   : DEC
                    {
                        ival, _ := strconv.ParseInt($1.str, 10, 64)
                        $$ = &Value{
                            Kind:    reflect.Int64,
                            Primary: ival,
                            lexer:   $1.lxr,
                        }
                        $$.lexer.debug("yacc: primary_value -> DEC: %d", ival)
                    }
                | OCT
                    {
                        ival, _ := strconv.ParseInt($1.str[1:], 8, 64)
                        $$ = &Value{
                            Kind:    reflect.Int64,
                            Primary: ival,
                            lexer:   $1.lxr,
                        }
                        $$.lexer.debug("yacc: primary_value -> OCT: %d(%s)", ival, $1.str)
                    }
                | HEX
                    {
                        ival, _ := strconv.ParseInt($1.str[2:], 16, 64)
                        $$ = &Value{
                            Kind:    reflect.Int64,
                            Primary: ival,
                            lexer:   $1.lxr,
                        }
                        $$.lexer.debug("yacc: primary_value -> HEX: %d(%s)", ival, $1.str)
                    }
                | FLOAT
                    {
                        fval, _ := strconv.ParseFloat($1.str, 64)
                        $$ = &Value{
                            Kind:    reflect.Int64,
                            Primary: fval,
                            lexer:   $1.lxr,
                        }
                    }
                | STRING
                    {
                        $$ = &Value{
                            Kind:    reflect.String,
                            Primary: $1.str,
                            lexer:   $1.lxr,
                        }
                        $$.lexer.debug("yacc: primary_value -> STRING: %s", $1)
                    }
                | BOOL
                    {
                        bval, _ := strconv.ParseBool($1.str)
                        $$ = &Value{
                            Kind:    reflect.String,
                            Primary: bval,
                            lexer:   $1.lxr,
                        }
                        $$.lexer.debug("yacc: primary_value -> BOOL: %s", $1)
                    }
                | RUNE
                    {
                        $$ = &Value{
                            Kind:    reflect.Int32,
                            Primary: $1.run,
                            lexer:   $1.lxr,
                        }
                        $$.lexer.debug("yacc: primary_value -> RUNE: %q", $1.run)
                    }
                ;

non_object      : primary_value
                    {
                        $$ = $1
                        $$.lexer.debug("yacc: non_object -> primary_value")
                    }
                | array_literal
                    {
                        $$ = $1
                        $$.lexer.debug("yacc: non_object -> array_literal")
                    }
                | TIME STRING
                    {
                        $$ = &Value{
                            Kind:    reflect.Struct,
                            Type:    "time.Time",
                            Primary: $2.str,
                        }
                        $$.lexer.debug("yacc: non_object -> TIME STRING : %s", $2)
                    }
                | DURATION STRING
                    {
                        $$ = &Value{
                            Kind:    reflect.Int64,
                            Type:    "time.Duration",
                            Primary: $2.str,
                        }
                        $$.lexer.debug("yacc: non_object -> DURATION STRING : %s", $2)
                    }
                ;

array_literal   : SQUARE_OPEN array_inside SQUARE_CLOSE
                    {
                        $$ = $2
                        $$.lexer.debug("yacc: array_literal -> SQUARE_OPEN array_inside SQUARE_CLOSE")
                    }
                ;

array_inside    : non_object
                    {
                        $$ = &Value {
                            Kind:  reflect.Slice,
                            Items: make([]*Value, 0, 4),
                            lexer: $1.lexer,
                        }
                        $$.Append("", $1)
                        $$.lexer.debug("yacc: array_inside -> non_object")
                    }
                | array_inside ',' non_object
                    {
                        $$ = $1
                        $$.Append("", $3)
                        $$.lexer.debug("yacc: array_inside -> array_inside ',' non_object")
                    }
                ;

object          : suite
                    {
                        $$ = $1
                        $$.lexer.debug("yacc: object -> suite")
                    }
                ;

suite           : NEWLINE INDENT list DEDENT
                    {
                        $$ = $3
                        $$.lexer.debug("yacc: suite -> NEWLINE INDENT list DEDENT")
                    }
                ;

%%

var (
    tokenTypeMap = map[tokenType]int {
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
