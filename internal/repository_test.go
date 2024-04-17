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

	t.Run("got status", func(t *testing.T) {
		tmp:= t.TempDir()
		repo, err := internal.FindOrCreateRepo(tmp)
		if err != nil {
			t.Error(err)
		} 
		internal.Status(repo)
		CreateFilesTesting(tmp, []string{"src"}, []TestingFile{
			{Name: "readme.md", RelativePath: "src/readme.md", Data: []byte("some-readme")},
			{Name: "readme.md", RelativePath: "src/cache.md", Data: []byte("some-cache")},
			{Name: "readme.md", RelativePath: "src/base64.md", Data: []byte("some-base64")},
		})
		internal.Status(repo)
	})

}
