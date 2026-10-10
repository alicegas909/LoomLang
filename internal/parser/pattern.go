package parser

import (
	"strconv"

	"github.com/you/loom/internal/ast"
	"github.com/you/loom/internal/lexer"
)

// parsePattern parses one pattern literal: `[layers]` or `<sequence>`.
func (p *parser) parsePattern() ast.Pat {
	switch p.cur().Kind {
	case lexer.LBracket:
		return p.parseStack()
	case lexer.LAngle:
		return p.parseAlt()
	}
	p.errorf(p.cur().Pos, "expected pattern, got %s", p.cur().Kind)
	return nil
}

// parseStack parses `[sequence { , sequence }]`.
func (p *parser) parseStack() ast.Pat {
	start := p.expect(lexer.LBracket).Pos
	p.depth++
	p.skipNewlines()

	var layers []ast.Pat
	for {
		layers = append(layers, p.parseSequence())
		if p.cur().Kind == lexer.Comma {
			p.next()
			p.skipNewlines()
			continue
		}
		break
	}
	p.skipNewlines()
	p.expect(lexer.RBracket)
	p.depth--

	if len(layers) == 1 {
		return layers[0]
	}
	return &ast.Stack{Sp: p.spanFrom(start), Layers: layers}
}

// parseAlt parses `<sequence>`.
func (p *parser) parseAlt() ast.Pat {
	start := p.expect(lexer.LAngle).Pos
	p.depth++
	p.skipNewlines()
	seq := p.parseSequence()
	p.skipNewlines()
	p.expect(lexer.RAngle)
	p.depth--
	return &ast.Alt{Sp: p.spanFrom(start), Steps: seq.(*ast.Seq).Steps}
}

// parseSequence parses steps until a closer, comma, newline or EOF.
func (p *parser) parseSequence() ast.Pat {
	start := p.cur().Pos
	var steps []ast.Pat
	for {
		if p.depth > 0 {
			p.skipNewlines()
		}
		switch p.cur().Kind {
		case lexer.RBracket, lexer.RAngle, lexer.Comma, lexer.Newline,
			lexer.Semicolon, lexer.EOF:
			return &ast.Seq{Sp: p.spanFrom(start), Steps: steps}
		}
		steps = append(steps, p.parseStep())
	}
}

// parseStep parses one element plus any tight modifiers.
func (p *parser) parseStep() ast.Pat {
	return p.parseModifiers(p.parseElement())
}

func (p *parser) parseElement() ast.Pat {
	t := p.cur()
	switch t.Kind {
	case lexer.Note, lexer.Number, lexer.Ident:
		p.next()
		return &ast.Atom{Sp: ast.Span{Start: t.Pos, End: t.Pos}, Kind: t.Kind, Text: t.Text}
	case lexer.Tilde:
		p.next()
		return &ast.Rest{Sp: ast.Span{Start: t.Pos, End: t.Pos}}
	case lexer.Dollar:
		p.next()
		name := p.expect(lexer.Ident)
		return &ast.Splice{
			Sp:   ast.Span{Start: t.Pos, End: name.Pos},
			Expr: &ast.Ident{Sp: ast.Span{Start: name.Pos, End: name.Pos}, Name: name.Text},
		}
	case lexer.LParen:
		e := p.parseParen()
		return &ast.Splice{Sp: p.spanFrom(t.Pos), Expr: e}
	case lexer.LBracket, lexer.LAngle:
		return p.parsePattern()
	}
	p.errorf(t.Pos, "unexpected %s in pattern", t.Kind)
	return nil
}

// parseModifiers applies every modifier that touches pat (no space before it).
func (p *parser) parseModifiers(pat ast.Pat) ast.Pat {
	for !p.cur().SpaceBefore {
		switch p.cur().Kind {
		case lexer.Star:
			p.next()
			pat = &ast.Fast{Sp: p.spanFrom(pat.Span().Start), Pat: pat, Factor: p.parseFactor()}
		case lexer.Slash:
			p.next()
			pat = &ast.Slow{Sp: p.spanFrom(pat.Span().Start), Pat: pat, Factor: p.parseFactor()}
		case lexer.Bang:
			p.next()
			count := 2
			if p.cur().Kind == lexer.Number && !p.cur().SpaceBefore {
				count = atoi(p.next())
			}
			pat = &ast.Repeat{Sp: p.spanFrom(pat.Span().Start), Pat: pat, Count: count}
		case lexer.At:
			p.next()
			amount := parseFloat(p.expect(lexer.Number))
			pat = &ast.Weight{Sp: p.spanFrom(pat.Span().Start), Pat: pat, Amount: amount}
		case lexer.Question:
			p.next()
			amount := 0.5
			if p.cur().Kind == lexer.Number && !p.cur().SpaceBefore {
				amount = parseFloat(p.next())
			}
			pat = &ast.Degrade{Sp: p.spanFrom(pat.Span().Start), Pat: pat, Amount: amount}
		case lexer.LParen:
			pat = p.parseEuclid(pat)
		default:
			return pat
		}
	}
	return pat
}

// parseEuclid parses `(steps, pulses [, rotate])`.
func (p *parser) parseEuclid(pat ast.Pat) ast.Pat {
	start := pat.Span().Start
	p.expect(lexer.LParen)
	steps := atoi(p.expect(lexer.Number))
	p.expect(lexer.Comma)
	pulses := atoi(p.expect(lexer.Number))
	rotate := 0
	if p.cur().Kind == lexer.Comma {
		p.next()
		rotate = atoi(p.expect(lexer.Number))
	}
	p.expect(lexer.RParen)
	return &ast.Euclid{Sp: p.spanFrom(start), Pat: pat, Steps: steps, Pulses: pulses, Rotate: rotate}
}

// parseFactor parses the argument of a fast/slow modifier.
func (p *parser) parseFactor() ast.Factor {
	t := p.cur()
	switch t.Kind {
	case lexer.Number:
		p.next()
		return ast.Factor{
			Sp:   ast.Span{Start: t.Pos, End: t.Pos},
			Expr: &ast.Number{Sp: ast.Span{Start: t.Pos, End: t.Pos}, Text: t.Text, Value: parseFloat(t)},
		}
	case lexer.Dollar:
		p.next()
		name := p.expect(lexer.Ident)
		return ast.Factor{
			Sp: ast.Span{Start: t.Pos, End: name.Pos},
			Pat: &ast.Splice{
				Sp:   ast.Span{Start: t.Pos, End: name.Pos},
				Expr: &ast.Ident{Sp: ast.Span{Start: name.Pos, End: name.Pos}, Name: name.Text},
			},
		}
	case lexer.LParen:
		e := p.parseParen()
		return ast.Factor{Sp: p.spanFrom(t.Pos), Expr: e}
	case lexer.LAngle:
		pat := p.parseAlt()
		return ast.Factor{Sp: ast.Span{Start: t.Pos, End: pat.Span().End}, Pat: pat}
	}
	p.errorf(t.Pos, "expected modifier factor, got %s", t.Kind)
	return ast.Factor{}
}

func atoi(t lexer.Token) int {
	n, err := strconv.Atoi(t.Text)
	if err != nil {
		return 0
	}
	return n
}
