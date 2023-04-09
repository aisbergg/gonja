// Package log provides a simple logging facility for debugging. Using the build
// tags above, you can enable/disable logging for each of the components. For
// example, if you want to enable logging for the parser, you can do:
//
//	go build -tags debug_parse.
//
// The disabled versions of the functions are optimized away by
// the compiler, so there is no runtime overhead.
//
// It is based on the https://github.com/lestrrat-go/pdebug package.
package log
