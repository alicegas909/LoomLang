// Package lexer turns source text into tokens.
//
// The lexer is deliberately dumb. It emits raw punctuation one character at a
// time, each token carrying SpaceBefore/SpaceAfter flags, and leaves every
// context-dependent decision to the parser: multi-character operators
// (|>, >>, <-, ->), pattern modifiers, signs on numbers, and whether an
// Ident is a sample WORD. Two classifications happen here because they need
// only the word itself: keywords (synth, param) and note literals.
//
// Comments (# to end of line) and spaces, tabs and carriage returns are
// skipped. Newlines are tokens, because the parser applies the statement-end
// rules. The lexer never panics, always makes progress, and always ends with
// EOF (repeated calls after the end keep returning EOF).
package lexer

// Lexer scans a source string.
type Lexer struct {
	src  string
	pos  int // byte offset of the next unread byte
	line int
	col  int
}

// New returns a lexer for src.
func New(src string) *Lexer {
	return &Lexer{src: src, line: 1, col: 1}
}

// Tokenize returns every token in src, including the final EOF.
func Tokenize(src string) []Token {
	l := New(src)
	var toks []Token
	for {
		t := l.Next()
		toks = append(toks, t)
		if t.Kind == EOF {
			return toks
		}
	}
}

// Next returns the next token.
func (l *Lexer) Next() Token {
	sawSpace := l.skipSpace()
	start := l.pos
	pos := Pos{Line: l.line, Col: l.col}
	spaceBefore := sawSpace || start == 0 || l.src[start-1] == '\n'

	if l.atEOF() {
		return Token{Kind: EOF, Pos: pos, SpaceBefore: spaceBefore, SpaceAfter: true}
	}

	kind := l.scan()
	return Token{
		Kind:        kind,
		Text:        l.src[start:l.pos],
		Pos:         pos,
		SpaceBefore: spaceBefore,
		SpaceAfter:  l.spaceAfter(),
	}
}

func (l *Lexer) atEOF() bool {
	return l.pos >= len(l.src)
}

// advance consumes one byte and updates the line and column.
func (l *Lexer) advance() {
	c := l.src[l.pos]
	l.pos++
	switch {
	case c == '\n':
		l.line++
		l.col = 1
	case c&0xC0 == 0x80:
		// UTF-8 continuation byte: still the same character, same column.
	default:
		l.col++
	}
}

// skipSpace skips spaces, tabs, carriage returns and comments, and reports
// whether it skipped anything. It stops before a newline.
func (l *Lexer) skipSpace() bool {
	skipped := false
	for !l.atEOF() {
		switch l.src[l.pos] {
		case ' ', '\t', '\r':
			l.advance()
		case '#':
			for !l.atEOF() && l.src[l.pos] != '\n' {
				l.advance()
			}
		default:
			return skipped
		}
		skipped = true
	}
	return skipped
}

// spaceAfter reports whether the next byte separates this token from the
// following one: whitespace, a newline, a comment, or the end of input.
func (l *Lexer) spaceAfter() bool {
	if l.atEOF() {
		return true
	}
	switch l.src[l.pos] {
	case ' ', '\t', '\r', '\n', '#':
		return true
	}
	return false
}

// scan consumes one token starting at l.pos and returns its kind. It always
// consumes at least one byte.
func (l *Lexer) scan() Kind {
	c := l.src[l.pos]
	switch {
	case isLetter(c):
		return l.scanWord()
	case isDigit(c):
		return l.scanNumber()
	case c == '"':
		return l.scanString()
	}

	l.advance()
	switch c {
	case '\n':
		return Newline
	case '[':
		return LBracket
	case ']':
		return RBracket
	case '<':
		return LAngle
	case '>':
		return RAngle
	case '(':
		return LParen
	case ')':
		return RParen
	case '{':
		return LBrace
	case '}':
		return RBrace
	case ',':
		return Comma
	case ';':
		return Semicolon
	case '.':
		return Dot
	case '$':
		return Dollar
	case '~':
		return Tilde
	case '!':
		return Bang
	case '@':
		return At
	case '?':
		return Question
	case '+':
		return Plus
	case '-':
		return Minus
	case '*':
		return Star
	case '/':
		return Slash
	case '=':
		return Equals
	case '|':
		return Pipe
	case '\\':
		return Backslash
	}

	// Unknown byte. Swallow the rest of a multi-byte UTF-8 character so that
	// one character produces one Illegal token.
	if c >= 0x80 {
		for !l.atEOF() && l.src[l.pos]&0xC0 == 0x80 {
			l.advance()
		}
	}
	return Illegal
}

// scanWord reads an identifier and classifies it as keyword, note or ident.
func (l *Lexer) scanWord() Kind {
	start := l.pos
	for !l.atEOF() && (isLetter(l.src[l.pos]) || isDigit(l.src[l.pos])) {
		l.advance()
	}
	word := l.src[start:l.pos]
	switch word {
	case "synth":
		return KwSynth
	case "param":
		return KwParam
	}
	if isNote(word) {
		return Note
	}
	return Ident
}

// scanNumber reads digits, plus a fractional part if a digit follows the dot.
// That keeps 0.25 whole but leaves the dot in lead.gain alone.
func (l *Lexer) scanNumber() Kind {
	for !l.atEOF() && isDigit(l.src[l.pos]) {
		l.advance()
	}
	if l.pos+1 < len(l.src) && l.src[l.pos] == '.' && isDigit(l.src[l.pos+1]) {
		l.advance() // the dot
		for !l.atEOF() && isDigit(l.src[l.pos]) {
			l.advance()
		}
	}
	return Number
}

// scanString reads from the opening quote to the closing quote. There are no
// escapes yet. A string cut off by a newline or EOF is Illegal; the newline
// itself is left for the next token.
func (l *Lexer) scanString() Kind {
	l.advance() // opening quote
	for !l.atEOF() {
		c := l.src[l.pos]
		if c == '\n' {
			return Illegal
		}
		l.advance()
		if c == '"' {
			return String
		}
	}
	return Illegal
}

func isLetter(c byte) bool {
	return c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c == '_'
}

func isDigit(c byte) bool {
	return c >= '0' && c <= '9'
}

// isNote reports whether w is a note literal: [a-g], an optional s or f, then
// one digit (c4, fs3, ef2). Words like bd4, a2b and bf are not notes.
func isNote(w string) bool {
	if len(w) < 2 || len(w) > 3 {
		return false
	}
	if w[0] < 'a' || w[0] > 'g' {
		return false
	}
	if len(w) == 3 && w[1] != 's' && w[1] != 'f' {
		return false
	}
	return isDigit(w[len(w)-1])
}
