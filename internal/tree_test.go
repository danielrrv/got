package internal_test

import (
	// "bytes"
	// "fmt"
	// "os"
	// "path/filepath"
	"fmt"
	"os"
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
		tmp := "/home/daniel/got/tests"
		
		err := os.MkdirAll(tmp, 0744)
		if err != nil {panic(err)}

		repo, err := internal.FindOrCreateRepo(tmp)
		if err != nil {
			t.Errorf("No repo found.")
		}

		files := []string{filepath.Join(tmp, "src/a/co.txt"), filepath.Join(tmp, "src/a/c/cx.txt"), filepath.Join(tmp, "src/a/b/mx.txt"), filepath.Join(tmp, "src/a/b/jx.txt")}

		err = os.MkdirAll(filepath.Join(tmp, "src"), 0744)
		if err != nil {panic(err)}
		err = os.MkdirAll(filepath.Join(tmp, "src/a/c"), 0744)
		if err != nil {panic(err)}
		err = os.MkdirAll(filepath.Join(tmp, "src/a/b"), 0744)
		if err != nil {panic(err)}


		for _, file := range files {
			fd, err := os.OpenFile(file, os.O_WRONLY|os.O_CREATE, 0755)
			if err != nil {
				panic(err)
			}

			if _, err := fd.Write([]byte(file)); err != nil {
				panic(err)
			}
			fd.Close()
		}

		m := internal.CreateTreeFromFiles(repo, files)
		fmt.Println(m)
		t.FailNow()
		tree := internal.FromMapToTree(repo, m, "src")
		tree.TraverseTree(repo,
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
