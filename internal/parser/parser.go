// Package parser builds an AST from source (milestone M1).
//
// Expression mode uses a Pratt parser (binding powers in precedence.go).
// Pattern mode (pattern.go) is a recursive-descent sub-parser. The two call
// each other at '[', '<', '(' and '$'.
//
// Syntax errors are recovered per statement: the parser records a *SyntaxError,
// skips to the end of the offending statement, and keeps going, so one block
// can report every mistake at once. Parse returns the statements it managed to
// build plus an ErrorList; because a statement that failed is simply absent,
// callers must not apply anything unless the error is nil.
package parser

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/you/loom/internal/ast"
	"github.com/you/loom/internal/lexer"
)

// Parse parses one submission (a REPL block or a whole file). Statements that
// fail to parse are dropped and their errors collected.
func Parse(src string) ([]ast.Stmt, error) {
	p := newParser(lexer.Tokenize(src))
	stmts, errs := p.parseProgram()
	if len(errs) == 0 {
		return stmts, nil
	}
	return stmts, ErrorList(errs)
}

// Incomplete reports whether parsing src failed only because the input ended
// early, so more text could still make it valid (an unclosed bracket, a
// trailing operator, an unfinished synth body, ...). A REPL uses this to
// decide whether to keep reading. It is false for empty input and for errors
// that more input cannot fix.
func Incomplete(src string) bool {
	p := newParser(lexer.Tokenize(src))
	eof := p.toks[len(p.toks)-1].Pos
	_, errs := p.parseProgram()
	if len(errs) == 0 {
		return false
	}
	for _, err := range errs {
		se, ok := err.(*SyntaxError)
		if !ok || se.Pos != eof {
			return false
		}
	}
	return true
}

// SyntaxError is one syntax error at a source position.
type SyntaxError struct {
	Pos lexer.Pos
	Msg string
}

func (e *SyntaxError) Error() string {
	return fmt.Sprintf("%d:%d: %s", e.Pos.Line, e.Pos.Col, e.Msg)
}

// ErrorList is the error returned when one or more statements fail to parse.
type ErrorList []error

func (e ErrorList) Error() string {
	if len(e) == 1 {
		return e[0].Error()
	}
	var b strings.Builder
	fmt.Fprintf(&b, "%d parse errors:", len(e))
	for _, err := range e {
		b.WriteString("\n  ")
		b.WriteString(err.Error())
	}
	return b.String()
}

// Unwrap exposes the individual errors to errors.Is/As.
func (e ErrorList) Unwrap() []error { return e }

type parser struct {
	toks  []lexer.Token
	pos   int
	depth int // open ( [ < nesting; newlines inside are not statement ends
}

func newParser(toks []lexer.Token) *parser {
	return &parser{toks: toks}
}

func (p *parser) cur() lexer.Token { return p.toks[p.pos] }

func (p *parser) peek() lexer.Token {
	if p.pos+1 < len(p.toks) {
		return p.toks[p.pos+1]
	}
	return p.toks[len(p.toks)-1]
}

// next returns the current token and advances (never past EOF).
func (p *parser) next() lexer.Token {
	t := p.toks[p.pos]
	if p.pos < len(p.toks)-1 {
		p.pos++
	}
	return t
}

func (p *parser) errorf(pos lexer.Pos, format string, args ...any) {
	panic(&SyntaxError{Pos: pos, Msg: fmt.Sprintf(format, args...)})
}

func (p *parser) expect(k lexer.Kind) lexer.Token {
	if p.cur().Kind != k {
		p.errorf(p.cur().Pos, "expected %s, got %s", k, p.cur().Kind)
	}
	return p.next()
}

// spanFrom returns a Span from start to the last consumed token.
func (p *parser) spanFrom(start lexer.Pos) ast.Span {
	end := start
	if p.pos > 0 {
		end = p.toks[p.pos-1].Pos
	}
	return ast.Span{Start: start, End: end}
}

// skipNewlines consumes newline tokens. It is used after an operator or an
// opening bracket, where a line break does not end the statement.
func (p *parser) skipNewlines() {
	for p.cur().Kind == lexer.Newline {
		p.next()
	}
}

// skipContinuationNewlines consumes newlines that do not end the statement:
// any newline inside unbalanced ( [ < brackets, and otherwise a newline whose
// next non-newline token continues the expression with |> or >>.
func (p *parser) skipContinuationNewlines() {
	if p.cur().Kind != lexer.Newline {
		return
	}
	if p.depth > 0 {
		p.skipNewlines()
		return
	}
	i := p.pos
	for i < len(p.toks) && p.toks[i].Kind == lexer.Newline {
		i++
	}
	if i < len(p.toks) {
		switch p.toks[i].Kind {
		case lexer.Pipe, lexer.RAngle:
			p.pos = i
		}
	}
}

// --- statements ---

func (p *parser) parseProgram() ([]ast.Stmt, []error) {
	var stmts []ast.Stmt
	var errs []error
	for {
		p.skipNewlines()
		if p.cur().Kind == lexer.EOF {
			break
		}
		stmt, err := p.parseOneStatement()
		if err != nil {
			errs = append(errs, err)
			p.recoverStatement()
			continue
		}
		stmts = append(stmts, stmt)
	}
	return stmts, errs
}

// parseOneStatement parses a statement and its line terminator, converting a
// *SyntaxError panic into an ordinary error.
func (p *parser) parseOneStatement() (stmt ast.Stmt, err error) {
	defer func() {
		if r := recover(); r != nil {
			se, ok := r.(*SyntaxError)
			if !ok {
				panic(r)
			}
			stmt, err = nil, se
		}
	}()

	stmt = p.parseStatement()
	switch p.cur().Kind {
	case lexer.Newline:
		p.next()
	case lexer.EOF:
	default:
		p.errorf(p.cur().Pos, "expected end of line, got %s", p.cur().Kind)
	}
	return stmt, nil
}

// recoverStatement skips tokens up to and including the next newline that is
// not inside ( [ { nesting, so the next statement starts clean.
func (p *parser) recoverStatement() {
	p.depth = 0
	depth := 0
	for p.cur().Kind != lexer.EOF {
		switch p.cur().Kind {
		case lexer.LParen, lexer.LBracket, lexer.LBrace:
			depth++
		case lexer.RParen, lexer.RBracket, lexer.RBrace:
			if depth > 0 {
				depth--
			}
		case lexer.Newline:
			if depth == 0 {
				p.next()
				return
			}
		}
		p.next()
	}
}

func (p *parser) parseStatement() ast.Stmt {
	start := p.cur().Pos

	var quant float64
	hasQuant := false
	if p.cur().Kind == lexer.At {
		p.next()
		n := p.expect(lexer.Number)
		quant = parseFloat(n)
		hasQuant = true
	}

	left := p.parseExpr(0)

	switch {
	case p.cur().Kind == lexer.Equals:
		target := p.pathOf(left)
		p.next()
		value := p.parseExpr(0)
		return &ast.Assign{Sp: p.spanFrom(start), Target: target, Value: value}

	case p.cur().Kind == lexer.LAngle && p.peek().Kind == lexer.Minus:
		target := p.pathOf(left)
		p.next() // <
		p.next() // -
		value := p.parseExpr(0)
		return &ast.Attach{
			Sp: p.spanFrom(start), Target: target, Value: value,
			Quant: quant, HasQuant: hasQuant,
		}

	default:
		return &ast.ExprStmt{
			Sp: p.spanFrom(start), Value: left,
			Quant: quant, HasQuant: hasQuant,
		}
	}
}

// pathOf extracts a dotted path from an Ident/Field chain.
func (p *parser) pathOf(e ast.Expr) []string {
	switch n := e.(type) {
	case *ast.Ident:
		return []string{n.Name}
	case *ast.Field:
		return append(p.pathOf(n.X), n.Name)
	}
	p.errorf(e.Span().Start, "invalid target: not a dotted path")
	return nil
}

// --- expressions (Pratt) ---

// parseExpr parses an expression, stopping at any operator with a binding
// power below minBP.
func (p *parser) parseExpr(minBP int) ast.Expr {
	left := p.parsePrefix()
	for {
		p.skipContinuationNewlines()

		if p.startsAtom() && bpApply >= minBP {
			arg := p.parseAccess()
			left = applyArg(left, arg)
			continue
		}

		op, lbp, ok := p.peekInfix()
		if !ok || lbp < minBP {
			break
		}
		p.next()
		if op == "|>" || op == ">>" {
			p.next()
		}
		p.skipNewlines()
		start := left.Span().Start
		right := p.parseExpr(lbp + 1)
		left = &ast.Binary{Sp: ast.Span{Start: start, End: right.Span().End}, Op: op, X: left, Y: right}
	}
	return left
}

func (p *parser) parsePrefix() ast.Expr {
	switch p.cur().Kind {
	case lexer.Backslash:
		return p.parseLambda()
	case lexer.Minus:
		start := p.next().Pos
		x := p.parseExpr(bpUnary)
		return &ast.Unary{Sp: ast.Span{Start: start, End: x.Span().End}, Op: "-", X: x}
	default:
		return p.parseAccess()
	}
}

// parseLambda parses `\x y -> body`.
func (p *parser) parseLambda() ast.Expr {
	start := p.expect(lexer.Backslash).Pos
	var params []string
	for p.cur().Kind == lexer.Ident {
		params = append(params, p.next().Text)
	}
	if len(params) == 0 {
		p.errorf(p.cur().Pos, "lambda has no parameters")
	}
	p.expect(lexer.Minus)
	p.expect(lexer.RAngle)
	body := p.parseExpr(0)
	return &ast.Lambda{Sp: p.spanFrom(start), Params: params, Body: body}
}

// parseParen parses `( expr )`, where newlines are insignificant.
func (p *parser) parseParen() ast.Expr {
	p.expect(lexer.LParen)
	p.depth++
	p.skipNewlines()
	e := p.parseExpr(0)
	p.skipNewlines()
	p.expect(lexer.RParen)
	p.depth--
	return e
}

// parseAccess parses a primary followed by dotted fields.
func (p *parser) parseAccess() ast.Expr {
	e := p.parseAtom()
	for p.cur().Kind == lexer.Dot {
		p.next()
		name := p.expect(lexer.Ident)
		e = &ast.Field{Sp: ast.Span{Start: e.Span().Start, End: name.Pos}, X: e, Name: name.Text}
	}
	return e
}

func (p *parser) parseAtom() ast.Expr {
	t := p.cur()
	switch t.Kind {
	case lexer.Number:
		p.next()
		return &ast.Number{Sp: ast.Span{Start: t.Pos, End: t.Pos}, Text: t.Text, Value: parseFloat(t)}
	case lexer.Note:
		p.next()
		return &ast.Note{Sp: ast.Span{Start: t.Pos, End: t.Pos}, Text: t.Text}
	case lexer.String:
		p.next()
		return &ast.String{Sp: ast.Span{Start: t.Pos, End: t.Pos}, Value: t.Text[1 : len(t.Text)-1]}
	case lexer.Ident:
		p.next()
		return &ast.Ident{Sp: ast.Span{Start: t.Pos, End: t.Pos}, Name: t.Text}
	case lexer.LParen:
		return p.parseParen()
	case lexer.LBracket, lexer.LAngle:
		start := t.Pos
		pat := p.parsePattern()
		pat = p.parseModifiers(pat)
		return &ast.PatternExpr{Sp: p.spanFrom(start), Pat: pat}
	case lexer.KwSynth:
		return p.parseSynth()
	default:
		p.errorf(t.Pos, "unexpected %s", t.Kind)
		return nil
	}
}

// parseSynth parses `synth { name = expr ... }`.
func (p *parser) parseSynth() ast.Expr {
	start := p.expect(lexer.KwSynth).Pos
	p.expect(lexer.LBrace)
	p.skipNewlines()

	var stmts []ast.SynthStmt
	for p.cur().Kind != lexer.RBrace && p.cur().Kind != lexer.EOF {
		stmtStart := p.cur().Pos
		param := false
		if p.cur().Kind == lexer.KwParam {
			param = true
			p.next()
		}
		name := p.expect(lexer.Ident)
		p.expect(lexer.Equals)
		value := p.parseExpr(0)
		stmts = append(stmts, ast.SynthStmt{
			Sp: p.spanFrom(stmtStart), Param: param, Name: name.Text, Value: value,
		})

		switch p.cur().Kind {
		case lexer.Newline, lexer.Semicolon:
			for p.cur().Kind == lexer.Newline || p.cur().Kind == lexer.Semicolon {
				p.next()
			}
		case lexer.RBrace:
		default:
			p.errorf(p.cur().Pos, "expected end of synth statement, got %s", p.cur().Kind)
		}
	}
	p.expect(lexer.RBrace)
	return &ast.SynthDef{Sp: p.spanFrom(start), Stmts: stmts}
}

// startsAtom reports whether the current token can begin an access, i.e. an
// application argument. `<` does not count when it is really the `<-` attach.
func (p *parser) startsAtom() bool {
	switch p.cur().Kind {
	case lexer.Number, lexer.Note, lexer.String, lexer.Ident,
		lexer.LParen, lexer.LBracket, lexer.KwSynth:
		return true
	case lexer.LAngle:
		return p.peek().Kind != lexer.Minus
	}
	return false
}

// applyArg appends arg to an Apply node, or starts one.
func applyArg(fn, arg ast.Expr) ast.Expr {
	if a, ok := fn.(*ast.Apply); ok {
		a.Args = append(a.Args, arg)
		a.Sp.End = arg.Span().End
		return a
	}
	return &ast.Apply{
		Sp:   ast.Span{Start: fn.Span().Start, End: arg.Span().End},
		Func: fn, Args: []ast.Expr{arg},
	}
}

// peekInfix returns the infix operator at the cursor and its binding power.
func (p *parser) peekInfix() (op string, lbp int, ok bool) {
	switch p.cur().Kind {
	case lexer.Plus:
		return "+", bpAdd, true
	case lexer.Minus:
		if p.peek().Kind == lexer.RAngle {
			return "", 0, false
		}
		return "-", bpAdd, true
	case lexer.Star:
		return "*", bpMul, true
	case lexer.Slash:
		return "/", bpMul, true
	case lexer.Pipe:
		if p.peek().Kind == lexer.RAngle {
			return "|>", bpPipe, true
		}
	case lexer.RAngle:
		if p.peek().Kind == lexer.RAngle {
			return ">>", bpCompose, true
		}
	}
	return "", 0, false
}

func parseFloat(t lexer.Token) float64 {
	f, err := strconv.ParseFloat(t.Text, 64)
	if err != nil {
		return 0
	}
	return f
}
