package xpath

import (
	"io/fs"
	"path/filepath"
)

const defaultRootMaxDepth = -1

type Root string

// NewRoot .
func NewRoot(root string) *Root {
	var r = Root(root)
	return &r
}

// Walk .
func (r *Root) Walk(fn func(path string, info fs.FileInfo) error) error {
	return filepath.Walk(r.String(), func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// do something
		if err := fn(path, info); err != nil {
			return err
		}
		return nil
	})
}

// String .
func (r *Root) String() string {
	return string(*r)
}

// Dirs .
func (r *Root) Dirs(opts ...rootOption) ([]*Root, error) {
	// var options = loadRootOptions(opts...)

	var roots = make([]*Root, 0)
	return roots, r.Walk(func(path string, info fs.FileInfo) error {
		if info.IsDir() {
			roots = append(roots, NewRoot(path))
		}
		return nil
	})
}

// Files .
func (r *Root) Files(opts ...rootOption) ([]*File, error) {
	// var options = loadRootOptions(opts...)

	var files = make([]*File, 0)
	return files, r.Walk(func(path string, info fs.FileInfo) error {
		if !info.IsDir() {
			files = append(files, NewFile(path))
		}
		return nil
	})
}

// rootOptions .
type rootOptions struct {
	maxDepth int8
}

type rootOption func(o *rootOptions)

// loadRootOptions .
func loadRootOptions(opts ...rootOption) *rootOptions {
	options := &rootOptions{
		maxDepth: defaultRootMaxDepth,
	}

	for _, o := range opts {
		o(options)
	}

	return options
}

// MaxDepth .
func MaxDepth(n int8) rootOption {
	return func(o *rootOptions) {
		o.maxDepth = n
	}
}
