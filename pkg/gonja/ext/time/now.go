package time

import (
	"fmt"
	"strconv"
	"strings"

	arrow "github.com/bmuller/arrow/lib"

	"github.com/aisbergg/gonja/pkg/gonja/errors"
	"github.com/aisbergg/gonja/pkg/gonja/exec"
	"github.com/aisbergg/gonja/pkg/gonja/parse"
)

type Offset struct {
	Years   int
	Months  int
	Days    int
	Hours   int
	Minutes int
	Seconds int
}

type NowStmt struct {
	Location *parse.Token
	TZ       string
	Format   string
	Offset   *Offset
}

func (stmt *NowStmt) Position() *parse.Token { return stmt.Location }
func (stmt *NowStmt) String() string {
	t := stmt.Position()
	return fmt.Sprintf("NowStmt(Line=%d Col=%d)", t.Line, t.Col)
}

func (stmt *NowStmt) Execute(r *exec.Renderer, tag *parse.StatementBlockNode) {
	r.Current = stmt
	var now arrow.Arrow

	cfg := r.ExtensionConfig["time"].(*Config)
	format := cfg.DatetimeFormat

	if cfg.Now != nil {
		now = *cfg.Now
	} else {
		now = arrow.Now()
	}

	if stmt.Format != "" {
		format = stmt.Format
	}

	now = now.InTimezone(stmt.TZ)

	if stmt.Offset != nil {
		offset := stmt.Offset
		if offset.Years != 0 || offset.Months != 0 || offset.Days != 0 {
			now = arrow.New(now.AddDate(offset.Years, offset.Months, offset.Days))
		}
		if offset.Hours != 0 {
			now = now.AddHours(offset.Hours)
		}
		if offset.Minutes != 0 {
			now = now.AddMinutes(offset.Minutes)
		}
		if offset.Seconds != 0 {
			now = now.AddSeconds(offset.Seconds)
		}
	}

	r.WriteString(now.CFormat(format))
}

func nowParser(p *parse.Parser, args *parse.Parser) parse.Statement {
	stmt := &NowStmt{
		Location: p.Current(),
	}

	// Timezone
	tz := args.Match(parse.TokenString)
	if tz == nil {
		errors.ThrowSyntaxError(args.Current().ErrorToken(), "now expect a timezone as first argument")
	}
	stmt.TZ = tz.Val

	// Offset
	if sign := args.Match(parse.TokenAdd, parse.TokenSub); sign != nil {
		offset := args.Match(parse.TokenString)
		if offset == nil {
			errors.ThrowSyntaxError(args.Current().ErrorToken(), "expected an time offset")
		}
		timeOffset, err := parseTimeOffset(offset.Val, sign.Val == "+")
		if err != nil {
			errors.ThrowSyntaxError(args.Current().ErrorToken(), "unable to parse time offset '%s': %s", offset.Val, err)

		}
		stmt.Offset = timeOffset
	}

	// Format
	if args.Match(parse.TokenComma) != nil {
		format := args.Match(parse.TokenString)
		if format == nil {
			errors.ThrowSyntaxError(args.Current().ErrorToken(), "expected a format string")
		}
		stmt.Format = format.Val
	}

	if !args.End() {
		errors.ThrowSyntaxError(args.Current().ErrorToken(), "malformed now-tag args")
	}

	return stmt
}

func parseTimeOffset(offset string, add bool) (*Offset, error) {
	pairs := strings.Split(offset, ",")
	specs := map[string]int{}
	for _, pair := range pairs {
		splitted := strings.Split(pair, "=")
		if len(splitted) != 2 {
			return nil, fmt.Errorf("expected a key=value pair, got '%s'", pair)
		}
		unit := strings.TrimSpace(splitted[0])
		value, err := strconv.Atoi(strings.TrimSpace(splitted[1]))
		if err != nil {
			return nil, err
		}
		specs[unit] = value
	}
	to := &Offset{}
	for unit, value := range specs {
		if !add {
			value = -value
		}
		switch strings.ToLower(unit) {
		case "year", "years":
			to.Years = value
		case "month", "months":
			to.Months = value
		case "day", "days":
			to.Days = value
		case "hour", "hours":
			to.Hours = value
		case "minute", "minutes":
			to.Minutes = value
		case "second", "seconds":
			to.Seconds = value
		default:
			return nil, fmt.Errorf("unknown unit '%s", unit)
		}
	}
	return to, nil
}

func init() {
	Statements.MustRegister("now", nowParser)
}
