package parse

// Config is a configuration for the parser and lexer.
type Config struct {
	// BlockStartString marks the beginning of a block. Defaults to '{%'
	BlockStartString string

	// BlockEndString marks the end of a block. Defaults to '%}'.
	BlockEndString string

	// VariableStartString marks the the beginning of a print statement. Defaults to '{{'.
	VariableStartString string

	// VariableEndString marks the end of a print statement. Defaults to '}}'.
	VariableEndString string

	// CommentStartString marks the beginning of a comment. Defaults to '{#'.
	CommentStartString string

	// CommentEndString marks the end of a comment. Defaults to '#}'.
	CommentEndString string

	// LineStatementPrefix will be used as prefix for line based statements, if
	// given and a string.
	LineStatementPrefix string

	// LineCommentPrefix will be used as prefix for line based comments, if given
	// and a string.
	LineCommentPrefix string
}

func NewConfig() *Config {
	return &Config{
		BlockStartString:    "{%",
		BlockEndString:      "%}",
		VariableStartString: "{{",
		VariableEndString:   "}}",
		CommentStartString:  "{#",
		CommentEndString:    "#}",
	}
}

func (cfg Config) Inherit() *Config {
	return &Config{
		BlockStartString:    cfg.BlockStartString,
		BlockEndString:      cfg.BlockEndString,
		VariableStartString: cfg.VariableStartString,
		VariableEndString:   cfg.VariableEndString,
		CommentStartString:  cfg.CommentStartString,
		CommentEndString:    cfg.CommentEndString,
	}
}
