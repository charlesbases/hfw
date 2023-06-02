package xpath

import (
	"fmt"
	"testing"
)

func Test(t *testing.T) {
	var root = `E:\Programs\src\github.com\charlesbases\hfw\`

	{
		roots, err := NewRoot(root).Dirs()
		if err != nil {
			panic(err)
		}
		fmt.Println("- - - - root - - - -")
		for _, x := range roots {
			fmt.Println(x.String())
		}
	}

	{
		files, err := NewRoot(root).Files()
		if err != nil {
			panic(err)
		}
		fmt.Println("- - - - file - - - -")
		for _, x := range files {
			fmt.Println(x.String())
		}
	}
}
