package loaders

import (
	"bufio"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/aisbergg/gonja/pkg/gonja/errors"
	"github.com/aisbergg/gonja/pkg/gonja/exec"
	"golang.org/x/text/encoding"
)

// FilesystemLoader represents a local filesystem loader for templates.
//
// Search paths can be relative or absolute. Relative paths are relative to the
// current working directory. For security reasons, the loader prevents access
// to files outside of the search paths. This means that path traversal using
// `../` is not possible.
//
// If you want to improve performance, you can use the [CachedLoader] to cache
// the loaded templates. Simply wrap the FilesystemLoader with the CachedLoader
// like this:
//
//	loader := loaders.NewCachedLoader(loaders.MustNewFileSystemLoader("templates"))
type FilesystemLoader struct {
	searchPaths    []string
	encoding       encoding.Encoding
	followingLinks bool
}

// MustNewFileSystemLoader creates a new FilesystemLoader. It panics if an error
// occurs.
func MustNewFileSystemLoader(searchPaths ...string) *FilesystemLoader {
	fs, err := NewFileSystemLoader(searchPaths...)
	if err != nil {
		panic(err)
	}
	return fs
}

// NewFileSystemLoader creates a new [FilesystemLoader].
func NewFileSystemLoader(searchPaths ...string) (*FilesystemLoader, error) {
	return NewFileSystemLoaderWithOptions(nil, false, searchPaths...)
}

// MustNewFileSystemLoaderWithOptions creates a new FilesystemLoader with the
// given options. It panics if an error occurs.
func MustNewFileSystemLoaderWithOptions(
	encoding encoding.Encoding,
	followLinks bool,
	searchPaths ...string,
) *FilesystemLoader {
	fs, err := NewFileSystemLoaderWithOptions(encoding, followLinks, searchPaths...)
	if err != nil {
		panic(err)
	}
	return fs
}

// NewFileSystemLoaderWithOptions creates a new [FilesystemLoader] with the
// given options. Mind that the files are searched in the order of the given
// search paths.
//
//   - encoding: The encoding of the template files. If nil, the default
//     encoding is used.
//   - followLinks: If true, symlinks are followed.
//   - searchPaths: The paths to search for templates. The paths can be relative
//     or absolute. Relative paths are relative to the current working directory.
func NewFileSystemLoaderWithOptions(
	encoding encoding.Encoding,
	followLinks bool,
	searchPaths ...string,
) (loader *FilesystemLoader, err error) {
	if len(searchPaths) == 0 {
		return nil, errors.NewTemplateLoadError("", "no search paths given")
	}

	// make search paths absolute
	for i, path := range searchPaths {
		searchPaths[i], err = filepath.Abs(path)
		if err != nil {
			return nil, errors.NewTemplateLoadError(path, "failed to make the given path '%s' absolute: %s", path, err)
		}
	}

	// remove duplicates and add to loader
	cleaned := make([]string, 0, len(searchPaths))
	seen := map[string]struct{}{}
	for _, path := range searchPaths {
		if _, ok := seen[path]; ok {
			continue
		}
		seen[path] = struct{}{}
		cleaned = append(cleaned, path)
	}

	return &FilesystemLoader{
		searchPaths:    cleaned,
		encoding:       encoding,
		followingLinks: followLinks,
	}, nil
}

// Load returns a template by name.
func (fs *FilesystemLoader) Load(name string, cfg *exec.EvalConfig) (*exec.Template, error) {
	// load file
	reader, err := fs.loadFile(name)
	if err != nil {
		return nil, err
	}
	defer reader.Close()
	buf, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, errors.NewTemplateLoadError(name, "error loading template: %s", err)
	}

	// parse template
	tpl, err := exec.NewTemplate(name, string(buf), cfg)
	if err != nil {
		return nil, err
	}

	return tpl, nil
}

// loadFile goes through the search paths and returns the contents of the first
// file that is found.
func (fs *FilesystemLoader) loadFile(path string) (io.ReadCloser, error) {
	// clean path to prevent directory traversal
	cleanedPath := filepath.Join("/", path)

	for _, searchPath := range fs.searchPaths {
		absPath := filepath.Join(searchPath, cleanedPath)
		// check if path exists
		stat, err := os.Stat(absPath)
		if os.IsNotExist(err) {
			continue
		}
		// check if path is a file
		if !stat.Mode().IsRegular() {
			continue
		}
		// check if path is a symlink
		if !fs.followingLinks && stat.Mode()&os.ModeSymlink != 0 {
			continue
		}

		// open file
		file, err := os.Open(absPath)
		if err != nil {
			return nil, errors.NewTemplateLoadError(cleanedPath, "failed to open file '%s': %s", path, err)
		}
		reader := bufio.NewReader(file)
		if fs.encoding != nil {
			// use configured encoding
			decoder := fs.encoding.NewDecoder()
			reader = bufio.NewReader(decoder.Reader(reader))
		}

		// return a ReadCloser that wraps the reader and the file
		return struct {
			io.Reader
			io.Closer
		}{reader, file}, nil
	}

	return nil, errors.NewTemplateNotFoundError(path)
}
