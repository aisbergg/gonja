package parse

import (
	log "github.com/aisbergg/gonja/internal/log/parse"
)

// ParseMath parses a math expression.
func (p *Parser) ParseMath() Expression {
	if log.Enabled {
		fm := log.FuncMarker()
		defer fm.End()
	}
	log.Print("parse: %s", p.Current())

	expr := p.parseConcat()
	for p.Peek(TokenAdd, TokenSub) != nil {
		tok := p.Pop()
		right := p.parseConcat()
		var opType BinOperatorType
		switch tok.Val {
		case "+":
			opType = OperatorAdd
		case "-":
			opType = OperatorSub
		}
		expr = &BinaryExpressionNode{
			Left:  expr,
			Right: right,
			Operator: &BinOperatorNode{
				Token: tok,
				Type:  opType,
			},
		}
	}

	log.Print("parsed expression: %s", expr)
	return expr
}

// parseConcat parses a concatenation expression.
func (p *Parser) parseConcat() Expression {
	if log.Enabled {
		fm := log.FuncMarker()
		defer fm.End()
	}
	log.Print("parse: %s", p.Current())

	expr := p.parseMathPrioritary()
	for p.Peek(TokenTilde) != nil {
		tok := p.Pop()
		right := p.parseMathPrioritary()
		expr = &BinaryExpressionNode{
			Left:  expr,
			Right: right,
			Operator: &BinOperatorNode{
				Token: tok,
				Type:  OperatorConcat,
			},
		}
	}

	log.Print("parsed expression: %s", expr)
	return expr
}

// parseMathPrioritary parses a math expression with priority.
func (p *Parser) parseMathPrioritary() Expression {
	if log.Enabled {
		fm := log.FuncMarker()
		defer fm.End()
	}
	log.Print("parse: %s", p.Current())

	expr := p.parseUnary()

	for p.Peek(TokenMul, TokenDiv, TokenFloordiv, TokenMod) != nil {
		tok := p.Pop()
		right := p.parseUnary()
		var opType BinOperatorType
		switch tok.Val {
		case "*":
			opType = OperatorMul
		case "/":
			opType = OperatorDiv
		case "//":
			opType = OperatorFloordiv
		case "%":
			opType = OperatorMod
		}
		expr = &BinaryExpressionNode{
			Left:  expr,
			Right: right,
			Operator: &BinOperatorNode{
				Token: tok,
				Type:  opType,
			},
		}
	}

	log.Print("parsed expression: %s", expr)
	return expr
}

// parseUnary parses a unary expression.
func (p *Parser) parseUnary() Expression {
	if log.Enabled {
		fm := log.FuncMarker()
		defer fm.End()
	}
	log.Print("parse: %s", p.Current())

	sign := p.Match(TokenAdd, TokenSub)
	expr := p.parsePower()

	if sign != nil {
		expr = &UnaryExpressionNode{
			Location: sign,
			Term:     expr,
			Negative: sign.Val == "-",
		}
	}

	log.Print("parsed expression: %s", expr)
	return expr
}

// parsePower parses a power expression.
func (p *Parser) parsePower() Expression {
	if log.Enabled {
		fm := log.FuncMarker()
		defer fm.End()
	}
	log.Print("parse: %s", p.Current())

	expr := p.ParseVariableOrLiteral()
	for p.Peek(TokenPow) != nil {
		tok := p.Pop()
		right := p.ParseVariableOrLiteral()
		expr = &BinaryExpressionNode{
			Left:  expr,
			Right: right,
			Operator: &BinOperatorNode{
				Token: tok,
				Type:  OperatorPower,
			},
		}
	}

	log.Print("parsed expression: %s", expr)
	return expr
}
