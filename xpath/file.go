package xpath

type File string

// NewFile .
func NewFile(file string) *File {
	var f = File(file)
	return &f
}

// String .
func (f *File) String() string {
	return string(*f)
}

// Handle do something
func (f *File) Handle(fn func(f *File) error) error {
	return fn(f)
}
