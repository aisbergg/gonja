package parse

import (
	log "github.com/aisbergg/gonja/internal/log/parse"
)

var compareOps = []TokenType{
	TokenEq, TokenNe,
	TokenGt, TokenGteq,
	TokenLt, TokenLteq,
}

// parseLogicalExpression parses a logical expression.
func (p *Parser) parseLogicalExpression() Expression {
	if log.Enabled {
		fm := log.FuncMarker()
		defer fm.End()
	}
	log.Print("parse: %s", p.Current())
	return p.parseOr()
}

// parseOr parses an 'or' expression.
func (p *Parser) parseOr() Expression {
	if log.Enabled {
		fm := log.FuncMarker()
		defer fm.End()
	}
	log.Print("parse: %s", p.Current())

	expr := p.parseAnd()
	for p.PeekName("or") != nil {
		tok := p.Pop()
		right := p.parseAnd()
		expr = &BinaryExpressionNode{
			Left:  expr,
			Right: right,
			Operator: &BinOperatorNode{
				Token: tok,
				Type:  OperatorOr,
			},
		}
	}

	log.Print("parsed expression: %s", expr)
	return expr
}

// parseAnd parses an 'and' expression.
func (p *Parser) parseAnd() Expression {
	if log.Enabled {
		fm := log.FuncMarker()
		defer fm.End()
	}
	log.Print("parse: %s", p.Current())

	expr := p.parseNot()
	for p.PeekName("and") != nil {
		tok := p.Pop()
		right := p.parseNot()
		expr = &BinaryExpressionNode{
			Left:  expr,
			Right: right,
			Operator: &BinOperatorNode{
				Token: tok,
				Type:  OperatorAnd,
			},
		}
	}

	log.Print("parsed expression: %s", expr)
	return expr
}

// parseNot parses a 'not' expression.
func (p *Parser) parseNot() Expression {
	if log.Enabled {
		fm := log.FuncMarker()
		defer fm.End()
	}
	log.Print("parse: %s", p.Current())

	op := p.MatchName("not")
	expr := p.parseCompare()

	if op != nil {
		expr = &NegationNode{
			Operator: op,
			Term:     expr,
		}
	}

	log.Print("parsed expression: %s", expr)
	return expr
}

// parseCompare parses a comparison expression.
func (p *Parser) parseCompare() Expression {
	if log.Enabled {
		fm := log.FuncMarker()
		defer fm.End()
	}
	log.Print("parse: %s", p.Current())

	expr := p.ParseMath()
	for p.Peek(compareOps...) != nil || p.PeekName("in", "not") != nil {
		tok := p.Pop()
		var opType BinOperatorType
		switch tok.Val {
		case "not":
			opType = OperatorNot
		case "in":
			opType = OperatorIn
		case "==":
			opType = OperatorEq
		case "!=", "<>":
			opType = OperatorNe
		case ">":
			opType = OperatorGt
		case ">=":
			opType = OperatorGteq
		case "<":
			opType = OperatorLt
		case "<=":
			opType = OperatorLteq
		}

		right := p.ParseMath()
		if right != nil {
			expr = &BinaryExpressionNode{
				Left: expr,
				Operator: &BinOperatorNode{
					Token: tok,
					Type:  opType,
				},
				Right: right,
			}
		}
	}

	expr = p.ParseTest(expr)

	log.Print("parsed expression: %s", expr)
	return expr
}
