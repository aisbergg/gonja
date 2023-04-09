package parse

import (
	"fmt"
	"strconv"

	u "github.com/aisbergg/gonja/pkg/gonja/utils"
)

// Node represents a token (or series of tokens) with associated data.
//
// All nodes contain position information marking the beginning of the
// corresponding source text segment; it is accessible via the Pos accessor
// method. Nodes may contain additional position info for language constructs
// where comments may be found between parts of the construct (typically any
// larger, parenthesized subpart). That position information is needed to
// properly position comments when printing the construct.
type Node interface {
	fmt.Stringer

	// Position returns the start token of the Node.
	Position() *Token
}

// Expression represents an evaluable expression part `{{ expr }}`.
type Expression interface {
	Node
}

// Statement represents a statement block `{% stmt %}`.
type Statement interface {
	Node
}

// -----------------------------------------------------------------------------

// TemplateNode is the root node of any template AST.
type TemplateNode struct {
	Name   string
	Nodes  []Node
	Blocks BlockSet
	Macros map[string]*MacroNode
	Parent *TemplateNode
}

// Position returns the start token of the Node.
func (tpl *TemplateNode) Position() *Token { return tpl.Nodes[0].Position() }
func (tpl *TemplateNode) String() string {
	tok := tpl.Position()
	return fmt.Sprintf("Template(Name=%s Line=%d Col=%d)", tpl.Name, tok.Line, tok.Col)
}

// GetBlocks returns the blocks with the given name.
func (tpl *TemplateNode) GetBlocks(name string) []*WrapperNode {
	var blocks []*WrapperNode
	if tpl.Parent != nil {
		blocks = tpl.Parent.GetBlocks(name)
	} else {
		blocks = []*WrapperNode{}
	}
	block, exists := tpl.Blocks[name]
	if exists {
		blocks = append([]*WrapperNode{block}, blocks...)
	}
	return blocks
}

type Trim struct {
	Left  bool
	Right bool
}

// DataNode represents a raw data (non-template text) a node.
type DataNode struct {
	Data *Token // data token
}

func (d *DataNode) Position() *Token { return d.Data }

func (d *DataNode) String() string {
	return fmt.Sprintf("Data(text=%s Line=%d Col=%d)",
		u.Ellipsis(d.Data.Val, 20), d.Data.Line, d.Data.Col)
}

// CommentNode represents a single comment node `{# comment #}`.
type CommentNode struct {
	Start *Token // Opening token
	Text  string // Comment text
	End   *Token // Closing token
	Trim  *Trim
}

// Position returns the start token of the Node.
func (c *CommentNode) Position() *Token { return c.Start }

// func (c *Comment) End() token.Pos { return token.Pos(int(c.Slash) + len(c.Text)) }
func (c *CommentNode) String() string {
	return fmt.Sprintf("Comment(text=%s Line=%d Col=%d)",
		u.Ellipsis(c.Text, 20), c.Start.Line, c.Start.Col)
}

// -----------------------------------------------------------------------------

// OutputNode represents a printable expression node `{{ expr }}`.
type OutputNode struct {
	Start      *Token
	Expression Expression
	End        *Token
	Trim       *Trim
}

// Position returns the start token of the Node.
func (o *OutputNode) Position() *Token { return o.Start }
func (o *OutputNode) String() string {
	return fmt.Sprintf("Output(Expression=%s Line=%d Col=%d)",
		o.Expression, o.Start.Line, o.End.Col)
}

type FilteredExpression struct {
	Expression Expression
	Filters    []*FilterCall
}

func (expr *FilteredExpression) Position() *Token {
	return expr.Expression.Position()
}
func (expr *FilteredExpression) String() string {
	t := expr.Expression.Position()

	return fmt.Sprintf("FilteredExpression(Expression=%s Line=%d Col=%d)",
		expr.Expression, t.Line, t.Col)
	// return fmt.Sprintf("<FilteredExpression Expression=%s", expr.Expression)
}

type FilterCall struct {
	Token *Token

	Name   string
	Args   []Expression
	Kwargs map[string]Expression

	// filterFunc FilterFunction
}

type TestExpression struct {
	Expression Expression
	Test       *TestCall
}

func (expr *TestExpression) String() string {
	t := expr.Position()

	return fmt.Sprintf("TestExpression(Expression=%s Test=%s Line=%d Col=%d)",
		expr.Expression, expr.Test, t.Line, t.Col)
	// return fmt.Sprintf("TestExpression(Expression=%s Test=%s)",
	// 	expr.Expression, expr.Test)
}
func (expr *TestExpression) Position() *Token {
	return expr.Expression.Position()
}

type TestCall struct {
	Token *Token

	Name   string
	Args   []Expression
	Kwargs map[string]Expression

	// testFunc TestFunction
}

func (tc *TestCall) String() string {
	return fmt.Sprintf("TestCall(name=%s Line=%d Col=%d)",
		tc.Name, tc.Token.Line, tc.Token.Col)
}

type StringNode struct {
	Location *Token
	Val      string
}

func (s *StringNode) Position() *Token { return s.Location }
func (s *StringNode) String() string   { return s.Location.Val }

// -----------------------------------------------------------------------------

// IntegerNode represents an integer literal node `{{ 1 }}`.
type IntegerNode struct {
	Location *Token
	Val      int
}

// Position returns the start token of the Node.
func (i *IntegerNode) Position() *Token { return i.Location }
func (i *IntegerNode) String() string   { return i.Location.Val }

// -----------------------------------------------------------------------------

// FloatNode represents a float literal node `{{ 1.0 }}`.
type FloatNode struct {
	Location *Token
	Val      float64
}

// Position returns the start token of the Node.
func (f *FloatNode) Position() *Token { return f.Location }
func (f *FloatNode) String() string   { return f.Location.Val }

// -----------------------------------------------------------------------------

// BoolNode represents a boolean literal node `{{ true }}`.
type BoolNode struct {
	Location *Token
	Val      bool
}

// Position returns the start token of the Node.
func (b *BoolNode) Position() *Token { return b.Location }
func (b *BoolNode) String() string   { return b.Location.Val }

// -----------------------------------------------------------------------------

// NameNode represents a variable name node `{{ my_var }}`.
type NameNode struct {
	Name *Token
}

// Position returns the start token of the Node.
func (n *NameNode) Position() *Token { return n.Name }
func (n *NameNode) String() string {
	t := n.Position()
	return fmt.Sprintf("Name(Val=%s Line=%d Col=%d)", t.Val, t.Line, t.Col)
}

// -----------------------------------------------------------------------------

// ListNode represents a list node `{{ [1, 2, 3] }}`.
type ListNode struct {
	Location *Token
	Val      []Expression
}

// Position returns the start token of the Node.
func (l *ListNode) Position() *Token { return l.Location }
func (l *ListNode) String() string   { return l.Location.Val }

// -----------------------------------------------------------------------------

// TupleNode represents a tuple literal node `{{ ('a', 'b') }}`.
type TupleNode struct {
	Location *Token
	Val      []Expression
}

// Position returns the start token of the Node.
func (t *TupleNode) Position() *Token { return t.Location }
func (t *TupleNode) String() string   { return t.Location.Val }

// -----------------------------------------------------------------------------

// DictNode represents a dictionary literal node `{{ {'k1': 'v', 'k2': 'v'} }}`.
type DictNode struct {
	Token *Token
	Pairs []*PairNode
}

// Position returns the start token of the Node.
func (d *DictNode) Position() *Token { return d.Token }
func (d *DictNode) String() string   { return d.Token.Val }

// -----------------------------------------------------------------------------

// PairNode represents a key/value pair node `{{ {'key': 'value'} }}`
type PairNode struct {
	Key   Expression
	Value Expression
}

// Position returns the start token of the Node.
func (p *PairNode) Position() *Token { return p.Key.Position() }
func (p *PairNode) String() string {
	t := p.Position()
	return fmt.Sprintf("Pair(Key=%s Value=%s Line=%d Col=%d)", p.Key, p.Value, t.Line, t.Col)
}

// -----------------------------------------------------------------------------

// CallNode represents a function call node `{{ func() }}`.
type CallNode struct {
	Location *Token
	Func     Node
	Args     []Expression
	Kwargs   map[string]Expression
}

// Position returns the start token of the Node.
func (c *CallNode) Position() *Token { return c.Location }
func (c *CallNode) String() string {
	t := c.Position()
	return fmt.Sprintf("Call(Args=%s Kwargs=%s Line=%d Col=%d)", c.Args, c.Kwargs, t.Line, t.Col)
}

// -----------------------------------------------------------------------------

// GetItemNode represents a node for looking up items from a list or dictionary
// `{{ obj[key] }}`.
type GetItemNode struct {
	Location *Token
	Node     Node
	Arg      string
	Index    int
}

// Position returns the start token of the Node.
func (g *GetItemNode) Position() *Token { return g.Location }
func (g *GetItemNode) String() string {
	t := g.Position()
	var param string
	if g.Arg != "" {
		param = fmt.Sprintf("Arg=%s", g.Arg)
	} else {
		param = fmt.Sprintf("Index=%s", strconv.Itoa(g.Index))
	}
	return fmt.Sprintf("GetItem(Node=%s %s Line=%d Col=%d)", g.Node, param, t.Line, t.Col)
}

// -----------------------------------------------------------------------------

// NegationNode represents a unary negation node `{{ not true }}`.
type NegationNode struct {
	Term     Expression
	Operator *Token
}

// Position returns the start token of the Node.
func (n *NegationNode) Position() *Token { return n.Operator }
func (n *NegationNode) String() string {
	t := n.Operator
	return fmt.Sprintf("Negation(term=%s Line=%d Col=%d)", n.Term, t.Line, t.Col)
}

// -----------------------------------------------------------------------------

// UnaryExpressionNode represents a unary expression node `{{ -1 }}`.
type UnaryExpressionNode struct {
	Location *Token
	Term     Expression
	Negative bool
}

// Position returns the start token of the Node.
func (ue *UnaryExpressionNode) Position() *Token { return ue.Location }
func (ue *UnaryExpressionNode) String() string {
	t := ue.Location
	return fmt.Sprintf("UnaryExpression(sign=%s term=%s Line=%d Col=%d)",
		t.Val, ue.Term, t.Line, t.Col)
}

// -----------------------------------------------------------------------------

// BinaryExpressionNode represents a binary expression node `{{ 1 + 2 }}`.
type BinaryExpressionNode struct {
	Left     Expression
	Right    Expression
	Operator *BinOperatorNode
}

// Position returns the start token of the Node.
func (be *BinaryExpressionNode) Position() *Token { return be.Left.Position() }
func (be *BinaryExpressionNode) String() string {
	t := be.Position()
	return fmt.Sprintf("BinaryExpression(operator=%s left=%s right=%s Line=%d Col=%d)", be.Operator.Token.Val, be.Left, be.Right, t.Line, t.Col)
}

type BinOperatorType int

const (
	OperatorInvalid BinOperatorType = iota
	OperatorAnd
	OperatorOr
	OperatorNot
	OperatorIs
	OperatorIn
	OperatorEq
	OperatorNe
	OperatorGt
	OperatorGteq
	OperatorLt
	OperatorLteq
	OperatorAdd
	OperatorSub
	OperatorMul
	OperatorDiv
	OperatorFloordiv
	OperatorMod
	OperatorPower
	OperatorConcat
)

func (t BinOperatorType) String() string {
	names := map[BinOperatorType]string{
		OperatorAnd:      "And",
		OperatorOr:       "Or",
		OperatorNot:      "Not",
		OperatorIs:       "Is",
		OperatorIn:       "In",
		OperatorEq:       "Eq",
		OperatorNe:       "Ne",
		OperatorGt:       "Gt",
		OperatorGteq:     "Gteq",
		OperatorLt:       "Lt",
		OperatorLteq:     "Lteq",
		OperatorAdd:      "Add",
		OperatorSub:      "Sub",
		OperatorMul:      "Mul",
		OperatorDiv:      "Div",
		OperatorFloordiv: "Floordiv",
		OperatorMod:      "Mod",
		OperatorPower:    "Power",
		OperatorConcat:   "Concat",
	}
	if name, ok := names[t]; ok {
		return name
	}

	return "Invalid"
}

// BinOperatorNode represents a binary operator node `{{ 1 + 2 }}`.
type BinOperatorNode struct {
	Token *Token
	Type  BinOperatorType
}

// Position returns the start token of the Node.
func (bo BinOperatorNode) Position() *Token { return bo.Token }
func (bo BinOperatorNode) String() string   { return bo.Token.String() }

// -----------------------------------------------------------------------------

// StatementBlockNode represents a statement block node `{% 1; 2; 3 %}`.
type StatementBlockNode struct {
	Location *Token
	Name     string
	Stmt     Statement
	Trim     *Trim
	LStrip   bool
}

// Position returns the start token of the Node.
func (sb StatementBlockNode) Position() *Token { return sb.Location }
func (sb StatementBlockNode) String() string {
	t := sb.Position()
	return fmt.Sprintf("StatementBlock(Name=%s Impl=%s Line=%d Col=%d)",
		sb.Name, sb.Stmt, t.Line, t.Col)
}

// -----------------------------------------------------------------------------

// WrapperNode wraps one or more nodes.
type WrapperNode struct {
	Location *Token
	Nodes    []Node
	EndTag   string
	Trim     *Trim
	LStrip   bool
}

// Position returns the start token of the Node.
func (w WrapperNode) Position() *Token { return w.Location }
func (w WrapperNode) String() string {
	t := w.Position()
	return fmt.Sprintf("Wrapper(Nodes=%s EndTag=%s Line=%d Col=%d)",
		w.Nodes, w.EndTag, t.Line, t.Col)
}

// -----------------------------------------------------------------------------

// MacroNode represents a macro definition node `{{% macro foo() }}`.
type MacroNode struct {
	Location *Token
	Name     string
	Args     []string
	Kwargs   []*PairNode
	Wrapper  *WrapperNode
}

// Position returns the start token of the Node.
func (m *MacroNode) Position() *Token { return m.Location }
func (m *MacroNode) String() string {
	t := m.Position()
	return fmt.Sprintf("Macro(Name=%s Args=%s Kwargs=%s Line=%d Col=%d)", m.Name, m.Args, m.Kwargs, t.Line, t.Col)
}
