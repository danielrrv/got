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
		tmp := t.TempDir()
		err := os.Mkdir(filepath.Join(tmp, projectRoot), 0744)
		if err != nil {
			panic(err)
		}
		path, _ := internal.FindRecursivelyFolder(filepath.Join(tmp, projectRoot), ".got", 2)
		if path != "" {
			t.Errorf("Expected not to find any .got folder")
		}

		err = os.Mkdir(filepath.Join(tmp, projectRoot, ".got"), 0744)
		if err != nil {
			panic(err)
		}
		childFolder := filepath.Join(tmp, projectRoot, "other-folder", "child-folder")
		err = os.MkdirAll(childFolder, 0744)
		if err != nil {
			panic(err)
		}

		path, _ = internal.FindRecursivelyFolder(childFolder, ".got", 3)
		if path == "" {
			t.Errorf("Expected to find .got folder")
		}

		fmt.Println(path)
	})

	t.Run("got status", func(t *testing.T) {
		tmp := t.TempDir()
		repo, err := internal.FindOrCreateRepo(tmp)
		if err != nil {
			t.Error(err)
		}
		repo.Status()
		CreateFilesTesting(tmp, []string{"src"}, []TestingFile{
			{Name: "readme.md", RelativePath: "src/readme.md", Data: []byte("some-readme")},
			{Name: "cache.rs", RelativePath: "src/cache.rs", Data: []byte("some-cache")},
			{Name: "base64.c", RelativePath: "src/base64.c", Data: []byte("some-base64")},
		})
		repo.Status()
		repo.Index.AddOrModifyEntries(repo, []string{"src/readme.md"})
		m := internal.CreateTreeFromFiles(repo, []string{"src/readme.md"})
		tree := internal.FromMapToTree(repo, m, "src")
		tree.TraverseTree(func(ti internal.TreeItem) {
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
		commit := internal.CreateCommit(repo, &tree, "message", "")
		hash, err := internal.WriteObject(repo, *commit, "commit")
		if err != nil {
			panic(err)
		} 
		fmt.Println("hash of the commit:",hash)
		// // Implementation to clear the cache after committing changes in DB.
		repo.Index.Cache = nil
		//Persist on disk/
		repo.Index.Persist(repo)
		//Save the reference HEAD.
		ref:= internal.Ref{
			Invalid: false,
			IsDirect: true,
			Reference: hash,
		} 
		ref.WriteRef(repo)

		//Find out the new status
		repo.Status()

		//Modify the file. Cache is clear because commit was made recently.
		CreateFilesTesting(tmp, []string{"src"}, []TestingFile{
			{Name: "readme.md", RelativePath: "src/readme.md", Data: []byte("let's change the content of this file with some modifications")},
		})
		//Find pit tje new status
		repo.Status()

		// Now add the file and see status
		repo.Index.AddOrModifyEntries(repo, []string{"src/readme.md"})
		//Find pit tje new status
		repo.Status()
		//Modify some lines
		CreateFilesTesting(tmp, []string{"src"}, []TestingFile{
			{Name: "readme.md", RelativePath: "src/readme.md", Data: []byte("let's change the content of this file with some modifications\n Let's add a new line and see")},
		})
		//Find pit the new status
		repo.Status()
		
	})

}
