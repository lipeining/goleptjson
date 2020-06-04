package leptjson

import (
	"bytes"
	"errors"
	"math"
	"strconv"
)

var (
	// ErrReachEnd the input string is reach to end
	ErrReachEnd = errors.New("json string reach end")
	// ErrUnexpectChar expect function fail
	ErrUnexpectChar = errors.New("get an unexpect char")
)

// define some global parse events
const (
	// LeptParseOK just ok
	LeptParseOK int = iota
	// LeptParseExpectValue expect value
	LeptParseExpectValue
	// LeptParseInvalidValue invalid value
	LeptParseInvalidValue
	// LeptParseRootNotSingular root not singular
	LeptParseRootNotSingular
	// LeptParseNumberTooBig number is to big
	LeptParseNumberTooBig
	// LeptParseMissQuotationMark
	LeptParseMissQuotationMark
	// LeptParseInvalidStringEscape
	LeptParseInvalidStringEscape
	// LeptParseInvalidStringChar
	LeptParseInvalidStringChar
)

// LeptType enums of json type
type LeptType int

const (
	// LeptNULL parse to nil
	LeptNULL LeptType = iota
	// LeptFALSE parse to false
	LeptFALSE
	// LeptTRUE parse to true
	LeptTRUE
	// LeptNUMBER parse to number like int float
	LeptNUMBER
	// LeptSTRING parse to string
	LeptSTRING
	// LeptARRAY parse to array
	LeptARRAY
	// LeptOBJECT parse to map
	LeptOBJECT
)

// LeptValue hold the value
type LeptValue struct {
	typ LeptType
	n   float64
	s   string
}

// NewLeptValue return a init LeptValue
func NewLeptValue() *LeptValue {
	return &LeptValue{
		typ: LeptFALSE,
	}
}

// LeptContext hold the input string
type LeptContext struct {
	json string
}

// NewLeptContext return a init LeptContext
func NewLeptContext(json string) *LeptContext {
	return &LeptContext{
		json: json,
	}
}

func expect(c *LeptContext, ch byte) {
	if len(c.json) == 0 {
		panic(ErrReachEnd)
	}
	first := c.json[0]
	if first != ch {
		panic(ErrUnexpectChar)
	}
	c.json = c.json[1:]
}

// LeptParseWhitespace use to parse white space like '\t' '\n' '\r' ' '
func LeptParseWhitespace(c *LeptContext) {
	i := 0
	n := len(c.json)
	for i < n && (c.json[i] == ' ' || c.json[i] == '\t' || c.json[i] == '\n' || c.json[i] == '\r') {
		i++
	}
	c.json = c.json[i:]
}

// LeptParseNull use to parse "null"
func LeptParseNull(c *LeptContext, v *LeptValue) int {
	expect(c, 'n')
	n := len(c.json)
	want := 4
	if n < want-1 {
		return LeptParseInvalidValue
	}
	if c.json[0] != 'u' || c.json[1] != 'l' || c.json[2] != 'l' {
		return LeptParseInvalidValue
	}
	c.json = c.json[want-1:]
	v.typ = LeptNULL
	return LeptParseOK
}

// LeptParseTrue use to parse "true"
func LeptParseTrue(c *LeptContext, v *LeptValue) int {
	expect(c, 't')
	n := len(c.json)
	want := 4
	if n < want-1 {
		return LeptParseInvalidValue
	}
	if c.json[0] != 'r' || c.json[1] != 'u' || c.json[2] != 'e' {
		return LeptParseInvalidValue
	}
	c.json = c.json[want-1:]
	v.typ = LeptTRUE
	return LeptParseOK
}

// LeptParseFalse use to parse "false"
func LeptParseFalse(c *LeptContext, v *LeptValue) int {
	expect(c, 'f')
	n := len(c.json)
	want := 5
	if n < want-1 {
		return LeptParseInvalidValue
	}
	if c.json[0] != 'a' || c.json[1] != 'l' || c.json[2] != 's' || c.json[3] != 'e' {
		return LeptParseInvalidValue
	}
	c.json = c.json[want-1:]
	v.typ = LeptFALSE
	return LeptParseOK
}

// LeptParseLiteral merge null true false
func LeptParseLiteral(c *LeptContext, v *LeptValue, literal string, typ LeptType) int {
	expect(c, literal[0])
	n := len(c.json)
	want := len(literal)
	if n < want-1 {
		return LeptParseInvalidValue
	}
	for i := 0; i < want-1; i++ {
		if c.json[i] != literal[i+1] {
			return LeptParseInvalidValue
		}
	}
	c.json = c.json[want-1:]
	v.typ = typ
	return LeptParseOK
}

// LeptParseNumber use to parse "Number"
func LeptParseNumber(c *LeptContext, v *LeptValue) int {
	var end string
	var err error
	v.n, end, err = strtod(c.json)
	if err != nil {
		return LeptParseInvalidValue
	}
	c.json = end
	v.typ = LeptNUMBER
	return LeptParseOK
}

// strtod use to parse input string to a number
func strtod(input string) (float64, string, error) {
	// number = [ "-" ] int [ frac ] [ exp ]
	// int = "0" / digit1-9 *digit
	// frac = "." 1*digit
	// exp = ("e" / "E") ["-" / "+"] 1*digit
	first := input[0]
	neg := false
	if first == '-' {
		neg = true
		input = input[1:]
	}
	var ret float64 = 0
	var integer int = 0
	var decimal int = 0
	var exp int = 0
	var err error
	n := len(input)
	var IllegalInput = errors.New("illegal input")
	if n == 0 {
		// no more charater
		return ret, "", IllegalInput
	}
	// take care of 0.0 0.12120
	if input[0] == '0' && n == 1 {
		// start with zero illegal like 0123
		return ret, "", nil
	}
	if input[0] == '0' && n > 1 && !(input[1] == '.' || input[1] == 'e' || input[1] == 'E') {
		// start with zero illegal like 0123
		return ret, "", IllegalInput
	}
	if !isDigit(input[0]) {
		// 非法开头字符
		return ret, "", IllegalInput
	}
	input, integer, err = parseInteger(input)
	if err != nil {
		return ret, "", err
	}
	n = len(input)
	if n <= 0 {
		// end just integer
		ret = float64(integer)
		if neg {
			return -ret, "", nil
		}
		return ret, "", nil
	}
	// frac or exp
	ret = float64(integer)
	if input[0] == '.' {
		// should be frac
		input, decimal, err = parseFrac(input)
		if err != nil {
			return ret, "", err
		}
		var frac int = 1
		for j := n - 1; j > len(input); j-- {
			frac *= 10
		}
		ret += float64(decimal) / float64(frac)
		if len(input) == 0 {
			if neg {
				return -ret, "", nil
			}
			return ret, "", nil
		}
		if !(input[0] == 'e' || input[0] == 'E') {
			// following is not exp
			return ret, "", IllegalInput
		}
		input, exp, err = parseExp(input)
		if err != nil || len(input) != 0 {
			// illegal next char
			return ret, "", IllegalInput
		}
		ret *= float64(math.Pow10(exp))
		if neg {
			return -ret, "", nil
		}
		return ret, "", nil
	} else if input[0] == 'e' || input[0] == 'E' {
		// should be exp
		// get exp
		input, exp, err = parseExp(input)
		if err != nil {
			return ret, "", err
		}
		if len(input) != 0 {
			// follow illegal char
			return ret, "", IllegalInput
		}
		ret *= float64(math.Pow10(exp))
		if neg {
			return -ret, "", nil
		}
		return ret, "", nil
	} else {
		// illegal next
		return ret, "", IllegalInput
	}
}

func parseExp(input string) (string, int, error) {
	if input[0] == 'e' || input[0] == 'E' {
		// should be exp
		if len(input) == 1 {
			// just e E illegal
			return "", 0, errors.New("input is not a exp")
		}
		expNeg := false
		if input[1] == '-' || input[1] == '+' {
			expNeg = input[1] == '-'
			input = input[2:]
		} else {
			input = input[1:]
		}
		// get exp
		input, exp, err := parseInteger(input)
		if err != nil {
			return "", 0, err
		}
		if expNeg {
			return input, -exp, err
		}
		return input, exp, err
	}
	return "", 0, errors.New("input is not a exp")
}
func parseFrac(input string) (string, int, error) {
	if input[0] == '.' {
		// should be frac
		return parseInteger(input[1:])
	}
	return "", 0, errors.New("input is not a frac")
}
func parseInteger(input string) (string, int, error) {
	i := 0
	n := len(input)
	for i < n && isDigit(input[i]) {
		// get the integer
		i++
	}
	integer, err := strconv.Atoi(input[:i])
	if err != nil {
		return "", 0, err
	}
	return input[i:], integer, nil
}

func isDigit(char byte) bool {
	return char >= '0' && char <= '9'
}

func isDigit1to9(char byte) bool {
	return char >= '1' && char <= '9'
}

// LeptParseString use to parse string include \u
// string = quotation-mark *char quotation-mark
// char = unescaped /
//    escape (
//        %x22 /          ; "    quotation mark  U+0022
//        %x5C /          ; \    reverse solidus U+005C
//        %x2F /          ; /    solidus         U+002F
//        %x62 /          ; b    backspace       U+0008
//        %x66 /          ; f    form feed       U+000C
//        %x6E /          ; n    line feed       U+000A
//        %x72 /          ; r    carriage return U+000D
//        %x74 /          ; t    tab             U+0009
//        %x75 4HEXDIG )  ; uXXXX                U+XXXX
// escape = %x5C          ; \
// quotation-mark = %x22  ; "
// unescaped = %x20-21 / %x23-5B / %x5D-10FFFF
func LeptParseString(c *LeptContext, v *LeptValue) int {
	expect(c, '"')
	var stack bytes.Buffer
	for i, n := 0, len(c.json); i < n; i++ {
		ch := c.json[i]
		switch ch {
		case '"':
			LeptSetString(v, stack.String())
			stack.Truncate(0)
			c.json = c.json[i+1:]
			return LeptParseOK
		case '\\':
			// 遇到第一个转义符号，需要连续匹配两个 \
			if i+1 >= n {
				return LeptParseInvalidStringEscape
			}
			switch c.json[i+1] {
			case '"':
				stack.WriteString("\"")
			case '\\':
				stack.WriteString("\\")
			case 'b':
				stack.WriteString("\b")
			case 'f':
				stack.WriteString("\f")
			case 'n':
				stack.WriteString("\n")
			case 'r':
				stack.WriteString("\r")
			case 't':
				stack.WriteString("\t")
			case '/':
				stack.WriteString("/")
			default:
				return LeptParseInvalidStringEscape
			}
			i++
		default:
			// 	unescaped = %x20-21 / %x23-5B / %x5D-10FFFF
			// 当中空缺的 %x22 是双引号，%x5C 是反斜线，都已经处理。所以不合法的字符是 %x00 至 %x1F。
			if ch < 0x20 {
				return LeptParseInvalidStringChar
			}
			stack.WriteByte(ch)
		}
	}
	// reach end of string becase the string has no \"
	return LeptParseMissQuotationMark
}

// LeptParseValue use to parse value switch to spec func
func LeptParseValue(c *LeptContext, v *LeptValue) int {
	n := len(c.json)
	if n == 0 {
		return LeptParseExpectValue
	}
	switch c.json[0] {
	case 'n':
		return LeptParseNull(c, v)
	case 't':
		return LeptParseTrue(c, v)
	case 'f':
		return LeptParseFalse(c, v)
	case '"':
		return LeptParseString(c, v)
	default:
		return LeptParseNumber(c, v)
	}
}

// LeptParse use to parse value the enter
func LeptParse(v *LeptValue, json string) int {
	if v == nil {
		panic("LeptParse v is nil")
	}
	c := NewLeptContext(json)
	v.typ = LeptNULL
	LeptParseWhitespace(c)
	if ret := LeptParseValue(c, v); ret != LeptParseOK {
		return ret
	}
	LeptParseWhitespace(c)
	if len(c.json) != 0 {
		return LeptParseRootNotSingular
	}
	return LeptParseOK
}

// LeptGetType use to get the type of value
func LeptGetType(v *LeptValue) LeptType {
	if v == nil {
		panic("LeptGetType v is nil")
	}
	return v.typ
}

// LeptSetNull use to set the type of null
func LeptSetNull(v *LeptValue) {
	if v == nil {
		panic("LeptGetNumber v is nil or typ is not LeptNUMBER")
	}
	v.typ = LeptNULL
}

// LeptGetNumber use to get the type of value
func LeptGetNumber(v *LeptValue) float64 {
	if v == nil || v.typ != LeptNUMBER {
		panic("LeptGetNumber v is nil or typ is not LeptNUMBER")
	}
	return v.n
}

// LeptSetNumber use to set the type of value
func LeptSetNumber(v *LeptValue, n float64) {
	if v == nil {
		panic("LeptSetNumber v is nil ")
	}
	v.n = n
	v.typ = LeptNUMBER
}

// LeptGetBoolean use to get the type of value
func LeptGetBoolean(v *LeptValue) int {
	if v == nil || !(v.typ == LeptFALSE || v.typ == LeptTRUE) {
		panic("LeptGetBoolean v is nil or typ is not boolean")
	}
	if v.typ == LeptFALSE {
		return 0
	}
	return 1
}

// LeptSetBoolean use to set the type of value
func LeptSetBoolean(v *LeptValue, n int) {
	if v == nil {
		panic("LeptSetBoolean v is nil ")
	}
	if n == 0 {
		v.typ = LeptFALSE
	} else {
		v.typ = LeptTRUE
	}
}

// LeptGetStringLength use to get the type of value
func LeptGetStringLength(v *LeptValue) int {
	if v == nil || v.typ != LeptSTRING {
		panic("LeptGetStringLength v is nil or typ is not string")
	}
	return len(v.s)
}

// LeptGetString use to get the type of value
func LeptGetString(v *LeptValue) string {
	if v == nil || v.typ != LeptSTRING {
		panic("LeptGetString v is nil or typ is not string")
	}
	return v.s
}

// LeptSetString use to get the type of value
func LeptSetString(v *LeptValue, s string) {
	if v == nil {
		panic("LeptSetString v is nil")
	}
	v.s = s
	v.typ = LeptSTRING
}
