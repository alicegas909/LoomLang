package lexer

import (
	"fmt"
	"strconv"
)

// Kind identifies a token type.
type Kind int

const (
	// Special.
	Illegal Kind = iota // unknown byte or rune, or an unterminated string
	EOF                 // end of input
	Newline             // \n (a \r before it is skipped as whitespace)

	// Literals. Token.Text holds the raw source text.
	Ident  // lead, bd, freq: letters, digits, underscore; not a note or keyword
	Note   // c4, fs3, ef2: [a-g](s|f)?[0-9]
	Number // 42, 0.25 (unsigned; a leading minus is a separate Minus token)
	String // "kick.wav": raw text including the quotes

	// Keywords.
	KwSynth // synth
	KwParam // param

	// Brackets.
	LBracket // [
	RBracket // ]
	LAngle   // <
	RAngle   // >
	LParen   // (
	RParen   // )
	LBrace   // {
	RBrace   // }

	// Separators.
	Comma     // ,
	Semicolon // ;
	Dot       // .

	// Pattern punctuation.
	Dollar   // $
	Tilde    // ~
	Bang     // !
	At       // @
	Question // ?

	// Operators. Multi-character operators (|>, >>, <-, ->) are assembled by
	// the parser from adjacent single-character tokens, never by the lexer.
	Plus      // +
	Minus     // -
	Star      // *
	Slash     // /
	Equals    // =
	Pipe      // |
	Backslash // \

	kindCount // number of kinds; keep last
)

var kindNames = [kindCount]string{
	Illegal:   "Illegal",
	EOF:       "EOF",
	Newline:   "Newline",
	Ident:     "Ident",
	Note:      "Note",
	Number:    "Number",
	String:    "String",
	KwSynth:   "KwSynth",
	KwParam:   "KwParam",
	LBracket:  "LBracket",
	RBracket:  "RBracket",
	LAngle:    "LAngle",
	RAngle:    "RAngle",
	LParen:    "LParen",
	RParen:    "RParen",
	LBrace:    "LBrace",
	RBrace:    "RBrace",
	Comma:     "Comma",
	Semicolon: "Semicolon",
	Dot:       "Dot",
	Dollar:    "Dollar",
	Tilde:     "Tilde",
	Bang:      "Bang",
	At:        "At",
	Question:  "Question",
	Plus:      "Plus",
	Minus:     "Minus",
	Star:      "Star",
	Slash:     "Slash",
	Equals:    "Equals",
	Pipe:      "Pipe",
	Backslash: "Backslash",
}

// String returns the kind's name, for tests and error messages.
func (k Kind) String() string {
	if k >= 0 && k < kindCount && kindNames[k] != "" {
		return kindNames[k]
	}
	return "Kind(" + strconv.Itoa(int(k)) + ")"
}

// Pos is a 1-based position in the source. Columns count runes, not bytes.
type Pos struct {
	Line, Col int
}

// Token is one lexical unit.
type Token struct {
	Kind        Kind
	Text        string // raw source text; empty for EOF
	Pos         Pos
	SpaceBefore bool // whitespace, a comment, or the start of a line immediately before this token
	SpaceAfter  bool // whitespace, a comment, a newline or EOF immediately after this token
}

// String formats a token for test failures, e.g. Ident("lead")@1:1.
func (t Token) String() string {
	return fmt.Sprintf("%s(%q)@%d:%d", t.Kind, t.Text, t.Pos.Line, t.Pos.Col)
}
