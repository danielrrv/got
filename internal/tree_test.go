package internal_test

import (
	// "bytes"
	// "fmt"
	// "os"
	// "path/filepath"
	"fmt"
	// "os"
	"path/filepath"
	"testing"

	internal "github.com/danielrrv/got/internal"
)

func TestTree(t *testing.T) {
	t.Run("CreateTreeFromFiles", func(t *testing.T) {
		tmp := t.TempDir()
		repo, err := internal.FindOrCreateRepo(tmp)
		if err != nil {
			t.Errorf("No repo found.")
		}
		files := []string{filepath.Join(tmp, "test/a/co.txt"), filepath.Join(tmp, "test/a/c/cx.txt"), filepath.Join(tmp, "test/a/b/mx.txt"), filepath.Join(tmp, "test/a/b/jx.txt")}
		internal.CreateTreeFromFiles(repo, files)
		// fmt.Printf("%v\n", m)
	})

	t.Run("Serialize/deserialize tree", func(t *testing.T) {
		tmp := t.TempDir()
		
		// err := os.MkdirAll(tmp, 0744)
		// if err != nil {panic(err)}

		repo, err := internal.FindOrCreateRepo(tmp)
		if err != nil {
			t.Errorf("No repo found.")
		}
		CreateFilesTesting(tmp, []string{"src","src/a", "src/a/c", "src/a/b"}, []TestingFile{
			{Name: "co.txt", RelativePath: "src/a/co.txt", Data: []byte("some-readme")},
			{Name: "cx.txt", RelativePath: "src/a/c/cx.txt", Data: []byte("some-cache")},
			{Name: "mx.txt", RelativePath: "src/a/b/mx.txt", Data: []byte("some-base64")},
			{Name: "jx.txt", RelativePath: "src/a/b/jx.txt", Data: []byte("some-other")},
		})
		

		m := internal.CreateTreeFromFiles(repo, []string{"src/a/co.txt", "src/a/c/cx.txt", "src/a/b/mx.txt", "src/a/b/jx.txt"})
		tree := internal.FromMapToTree(repo, m, "src")
		tree.TraverseTree(
			func(ti internal.TreeItem) {
				//	Here we have to go index and capture the cache of the stage area.
				blob, err := internal.BlobFromUserPath(repo, ti.Path)
				if err != nil {
					panic(err)
				}
				internal.WriteObject(repo, blob, internal.BlobHeaderName)
				// fmt.Println(ti)
			},
			func(ti internal.TreeItem) {
				internal.WriteObject(repo, ti, internal.TreeHeaderName)
			},
		)

		rawTreeContent, err := internal.ReadObject(repo, internal.TreeHeaderName, tree.Hash)
		if err != nil {
			panic(err)
		}
		var dummy internal.TreeItem
		deserializeTeee:= dummy.Deserialize(rawTreeContent) 

		fmt.Println(deserializeTeee.Hash, tree.Hash)

	})
}
