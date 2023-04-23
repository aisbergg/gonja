package exec

import (
	"fmt"
	"sort"
	"strings"
)

// KVPair represents a key/value pair.
type KVPair struct {
	Key   string
	Value Value
}

// VarArgs represents pythonic variadic args/kwargs.
type VarArgs struct {
	Args         []Value
	Kwargs       []KVPair
	ValueFactory *ValueFactory
}

// NewVarArgs creates a new VarArgs.
func NewVarArgs(valueFactory *ValueFactory) *VarArgs {
	return &VarArgs{
		Args:   []Value{},
		Kwargs: make([]KVPair, 0),
	}
}

// String returns a string representation of the variables arguments.
func (va *VarArgs) String() string {
	args := []string{}
	for _, arg := range va.Args {
		args = append(args, arg.String())
	}
	for _, kv := range va.Kwargs {
		args = append(args, fmt.Sprintf("%s=%s", kv.Key, kv.Value.String()))
	}
	return strings.Join(args, ", ")
}

// First returns the first argument or nil AsValue.
func (va *VarArgs) First() Value {
	if len(va.Args) > 0 {
		return va.Args[0]
	}
	return NewNilValue()
}

// HasKwarg returns true if the keyword argument exists.
func (va *VarArgs) HasKwarg(key string) bool {
	for _, kv := range va.Kwargs {
		if kv.Key == key {
			return true
		}
	}
	return false
}

// GetKwarg gets a keyword argument. It panics if the keyword argument does not
// exist. Make sure to define all the expected keyword arguments before calling
// this.
func (va *VarArgs) GetKwarg(key string) Value {
	for _, kv := range va.Kwargs {
		if kv.Key == key {
			return kv.Value
		}
	}
	panic(fmt.Errorf("[BUG] keyword argument %s does not exist", key))
}

// SetKwarg sets a keyword argument
func (va *VarArgs) SetKwarg(key string, value Value) {
	for i, kv := range va.Kwargs {
		if kv.Key == key {
			va.Kwargs[i].Value = value
			return
		}
	}
	va.Kwargs = append(va.Kwargs, KVPair{Key: key, Value: value})
}

// setDefaultKwarg sets a keyword argument if it does not exist.
func (va *VarArgs) setDefaultKwarg(key string, value any) {
	for _, kv := range va.Kwargs {
		if kv.Key == key {
			return
		}
	}
	va.Kwargs = append(va.Kwargs, KVPair{Key: key, Value: va.ValueFactory.NewValue(value, false)})
}

// Kwarg represents a keyword argument.
type Kwarg struct {
	Name    string
	Default any
}

// Expect validates VarArgs against an expected signature
func (va *VarArgs) Expect(args int, kwargs []*Kwarg) *ReducedVarArgs {
	rva := &ReducedVarArgs{VarArgs: va}
	reduced := &VarArgs{
		Args:   va.Args,
		Kwargs: make([]KVPair, 0),
	}

	if args == 0 && len(kwargs) == 0 && (len(va.Args) > 0 || len(va.Kwargs) > 0) {
		rva.err = fmt.Errorf("expected no arguments, got %d", len(va.Args)+len(va.Kwargs))
		return rva
	}

	// set args
	reduceIdx := -1
	unexpectedArgs := []string{}
	if len(va.Args) < args {
		// Priority on missing arguments
		if args > 1 {
			rva.err = fmt.Errorf("expected %d arguments, got %d", args, len(va.Args))
		} else {
			rva.err = fmt.Errorf("expected an argument, got %d", len(va.Args))
		}
		return rva
	} else if len(va.Args) > args {
		reduced.Args = va.Args[:args]
		for idx, arg := range va.Args[args:] {
			if len(kwargs) > idx {
				reduced.Kwargs = append(reduced.Kwargs, KVPair{Key: kwargs[idx].Name, Value: arg})
				reduceIdx = idx + 1
			} else {
				unexpectedArgs = append(unexpectedArgs, arg.String())
			}
		}
	}

	// set kwargs
	unexpectedKwArgs := []string{}
outerLoop:
	for _, inKwarg := range va.Kwargs {
		for defIdx, defKwarg := range kwargs {
			if inKwarg.Key == defKwarg.Name {
				if reduceIdx < 0 || defIdx >= reduceIdx {
					reduced.Kwargs = append(reduced.Kwargs, inKwarg)
					continue outerLoop
				} else {
					rva.err = fmt.Errorf("got multiple values for argument '%s'", inKwarg.Key)
					return rva
				}
			}
		}
		kv := strings.Join([]string{inKwarg.Key, inKwarg.Value.String()}, "=")
		unexpectedKwArgs = append(unexpectedKwArgs, kv)
	}

	if len(unexpectedArgs) > 0 {
		if len(unexpectedArgs) == 1 {
			rva.err = fmt.Errorf("unexpected argument '%s'", unexpectedArgs[0])
		} else {
			rva.err = fmt.Errorf("unexpected arguments '%s'", strings.Join(unexpectedArgs, ", "))
		}
	} else if len(unexpectedKwArgs) > 0 {
		sort.Strings(unexpectedKwArgs)
		if len(unexpectedKwArgs) == 1 {
			rva.err = fmt.Errorf("unexpected keyword argument '%s'", unexpectedKwArgs[0])
		} else {
			rva.err = fmt.Errorf("unexpected keyword arguments '%s'", strings.Join(unexpectedKwArgs, ", "))
		}
	}
	if rva.err != nil {
		return rva
	}

	// fill defaults
	for _, kwarg := range kwargs {
		reduced.setDefaultKwarg(kwarg.Name, kwarg.Default)
	}
	rva.VarArgs = reduced
	return rva
}

// ExpectArgs ensures VarArgs receive only arguments
func (va *VarArgs) ExpectArgs(args int) *ReducedVarArgs {
	return va.Expect(args, []*Kwarg{})
}

// ExpectNothing ensures VarArgs does not receive any argument
func (va *VarArgs) ExpectNothing() *ReducedVarArgs {
	return va.ExpectArgs(0)
}

// ExpectKwArgs allow to specify optionally expected KwArgs
func (va *VarArgs) ExpectKwArgs(kwargs []*Kwarg) *ReducedVarArgs {
	return va.Expect(0, kwargs)
}

// ReducedVarArgs represents pythonic variadic args/kwargs
// but values are reduced (ie. kwargs given as args are accessible by name)
type ReducedVarArgs struct {
	*VarArgs
	err error
}

// IsError returns true if there was an error on Expect call
func (rva *ReducedVarArgs) IsError() bool {
	return rva.err != nil
}

func (rva *ReducedVarArgs) Error() string {
	if rva.IsError() {
		return rva.err.Error()
	}
	return ""
}
