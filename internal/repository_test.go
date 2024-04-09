package internal_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	internal "github.com/danielrrv/got/internal"
)

var (
	projectRoot = "src"
	) 


func TestRepository(t *testing.T) {
	
	t.Run("find folder recursively up", func(t *testing.T) {
		tmp:= t.TempDir()
		err := os.Mkdir(filepath.Join(tmp, projectRoot), 0744)
		if err != nil {
			panic(err)
		}
		path, _ := internal.FindRecursivelyFolder(filepath.Join(tmp, projectRoot), ".got", 2)
		if path != ""{
			t.Errorf("Expected not to find any .got folder")
		}

		err = os.Mkdir(filepath.Join(tmp, projectRoot, ".got"), 0744)
		if err != nil {
			panic(err)
		}
		childFolder :=filepath.Join(tmp, projectRoot, "other-folder", "child-folder") 
		err = os.MkdirAll(childFolder, 0744)
		if err != nil {
			panic(err)
		}

		path, _ = internal.FindRecursivelyFolder(childFolder, ".got", 3)
		if path == ""{
			t.Errorf("Expected to find .got folder")
		}
		
		fmt.Println(path)
	})
}
