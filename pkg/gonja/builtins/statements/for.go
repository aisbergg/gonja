package statements

import (
	"fmt"
	"math"

	"github.com/aisbergg/gonja/pkg/gonja/errors"
	"github.com/aisbergg/gonja/pkg/gonja/exec"
	"github.com/aisbergg/gonja/pkg/gonja/parse"
)

type ForStmt struct {
	key string
	// value is only used for maps: for key, value in map
	value           string
	objectEvaluator parse.Expression
	ifCondition     parse.Expression

	bodyWrapper  *parse.WrapperNode
	emptyWrapper *parse.WrapperNode
}

func (stmt *ForStmt) Position() *parse.Token { return stmt.bodyWrapper.Position() }
func (stmt *ForStmt) String() string {
	t := stmt.Position()
	return fmt.Sprintf("ForStmt(Line=%d Col=%d)", t.Line, t.Col)
}

type LoopInfos struct {
	Index      int         `gonja:"index"`
	Index0     int         `gonja:"index0"`
	RevIndex   int         `gonja:"revindex"`
	RevIndex0  int         `gonja:"revindex0"`
	First      bool        `gonja:"first"`
	Last       bool        `gonja:"last"`
	Length     int         `gonja:"length"`
	Depth      int         `gonja:"depth"`
	Depth0     int         `gonja:"depth0"`
	PrevItem   *exec.Value `gonja:"previtem"`
	NextItem   *exec.Value `gonja:"nextitem"`
	_lastValue *exec.Value
}

func (li *LoopInfos) Cycle(va *exec.VarArgs) *exec.Value {
	return va.Args[int(math.Mod(float64(li.Index0), float64(len(va.Args))))]
}

func (li *LoopInfos) Changed(value *exec.Value) bool {
	same := li._lastValue != nil && value.EqualValueTo(li._lastValue)
	li._lastValue = value
	return !same
}

func (stmt *ForStmt) Execute(r *exec.Renderer, tag *parse.StatementBlockNode) {
	r.Current = stmt
	obj := r.Eval(stmt.objectEvaluator)

	// Create loop struct
	items := exec.NewDict()

	// First iteration: filter values to ensure proper LoopInfos
	obj.Iterate(func(idx, count int, key, value *exec.Value) bool {
		sub := r.Inherit()
		ctx := sub.Ctx
		pair := &exec.Pair{}

		// There's something to iterate over (correct type and at least 1 item)
		// Update loop infos and public context
		if stmt.value != "" && !key.IsString() && key.Len() == 2 {
			key.Iterate(func(idx, count int, key, value *exec.Value) bool {
				switch idx {
				case 0:
					ctx.Set(stmt.key, key)
					pair.Key = key
				case 1:
					ctx.Set(stmt.value, key)
					pair.Value = key
				}
				return true
			}, func() {})
		} else {
			ctx.Set(stmt.key, key)
			pair.Key = key
			if value != nil {
				ctx.Set(stmt.value, value)
				pair.Value = value
			}
		}

		if stmt.ifCondition != nil {
			if !sub.Eval(stmt.ifCondition).IsTrue() {
				return true
			}
		}
		items.Pairs = append(items.Pairs, pair)
		return true
	}, func() {
		// Nothing to iterate over (maybe wrong type or no items)
		if stmt.emptyWrapper != nil {
			sub := r.Inherit()
			err := sub.ExecuteWrapper(stmt.emptyWrapper)
			if err != nil {
				// pass error up the execution stack
				panic(err)
			}
		}
	})

	// 2nd pass: all values are defined, render
	length := len(items.Pairs)
	loop := &LoopInfos{
		First:  true,
		Index0: -1,
	}
	for idx, pair := range items.Pairs {
		r.EndTag(tag.Trim)
		sub := r.Inherit()
		ctx := sub.Ctx

		ctx.Set(stmt.key, pair.Key)
		if pair.Value != nil {
			ctx.Set(stmt.value, pair.Value)
		}

		ctx.Set("loop", loop)
		loop.Index0 = idx
		loop.Index = loop.Index0 + 1
		if idx == 1 {
			loop.First = false
		}
		if idx+1 == length {
			loop.Last = true
		}
		loop.RevIndex = length - idx
		loop.RevIndex0 = length - (idx + 1)

		if idx == 0 {
			loop.PrevItem = exec.AsValue(nil)
		} else {
			pp := items.Pairs[idx-1]
			if pp.Value != nil {
				loop.PrevItem = exec.AsValue([2]*exec.Value{pp.Key, pp.Value})
			} else {
				loop.PrevItem = pp.Key
			}
		}

		if idx == length-1 {
			loop.NextItem = exec.AsValue(nil)
		} else {
			np := items.Pairs[idx+1]
			if np.Value != nil {
				loop.NextItem = exec.AsValue([2]*exec.Value{np.Key, np.Value})
			} else {
				loop.NextItem = np.Key
			}
		}

		// Render elements with updated context
		err := sub.ExecuteWrapper(stmt.bodyWrapper)
		if err != nil {
			// pass error up the execution stack
			panic(err)
		}
	}
}

func forParser(p *parse.Parser, args *parse.Parser) parse.Statement {
	stmt := &ForStmt{}

	// Arguments parsing
	var valueToken *parse.Token
	keyToken := args.Match(parse.TokenName)
	if keyToken == nil {
		errors.ThrowSyntaxError(parse.AsErrorToken(p.Current()), "expected an key identifier as first argument for 'for'-tag")
	}

	if args.Match(parse.TokenComma) != nil {
		// Value name is provided
		valueToken = args.Match(parse.TokenName)
		if valueToken == nil {
			errors.ThrowSyntaxError(parse.AsErrorToken(p.Current()), "value name must be an identifier")
		}
	}

	if args.MatchName("in") == nil {
		errors.ThrowSyntaxError(parse.AsErrorToken(p.Current()), "expected keyword 'in' after key name")
	}

	objectEvaluator := args.ParseExpression()
	stmt.objectEvaluator = objectEvaluator
	stmt.key = keyToken.Val
	if valueToken != nil {
		stmt.value = valueToken.Val
	}

	if args.MatchName("if") != nil {
		var ifCondition = args.ParseExpression()
		stmt.ifCondition = ifCondition
	}

	if !args.End() {
		errors.ThrowSyntaxError(parse.AsErrorToken(p.Current()), "malformed for-loop args")
	}

	// Body wrapping
	wrapper, endargs := p.WrapUntil("else", "endfor")
	stmt.bodyWrapper = wrapper

	if !endargs.End() {
		errors.ThrowSyntaxError(parse.AsErrorToken(p.Current()), "arguments not allowed here")
	}

	if wrapper.EndTag == "else" {
		// if there's an else in the if-statement, we need the else-Block as well
		wrapper, endargs = p.WrapUntil("endfor")
		stmt.emptyWrapper = wrapper

		if !endargs.End() {
			errors.ThrowSyntaxError(parse.AsErrorToken(p.Current()), "arguments not allowed here")
		}
	}

	return stmt
}

func init() {
	All.MustRegister("for", forParser)
}
