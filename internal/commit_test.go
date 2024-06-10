package internal_test

import (
	"testing"

	internal "github.com/danielrrv/got/internal"
)

func TestCommit(t *testing.T) {
	t.Run("createCommit", func(t *testing.T) {
		projectTemporalFolder := t.TempDir()

		repo, err := internal.FindOrCreateRepo(projectTemporalFolder)
		if err != nil {
			t.Errorf("Expected to create the repo, %v", err.Error())
		}
		CreateFilesTesting(projectTemporalFolder, []string{"src"}, []TestingFile{
			{Name: "readme.md", RelativePath: "src/readme.md", Data: []byte("some-readme")},
			{Name: "cache.rs", RelativePath: "src/cache.rs", Data: []byte("some-cache")},
			{Name: "base64.c", RelativePath: "src/base64.c", Data: []byte("some-base64")},
		})
		// Previous commit must have a tree.
		// We have to deserialize the tree and validate the new blobs are different from the tree already.
		//To compare 2 blob in the tree the have to have the same absolute path.
		repo.Index.AddOrModifyEntries(repo, []string{"src/readme.md", "src/cache.rs", "src/base64.c"})
		repo.Index.Persist(repo)
		m := internal.CreateTreeFromFiles(repo, []string{"src/readme.md", "src/cache.rs", "src/base64.c"})		
		tree:=internal.FromMapToTree(repo, m, "src")

		tree.TraverseTree(func(ti internal.TreeItem) {
				//	Here we have to go index and capture the cache of the stage area.
				blob, err := internal.BlobFromUserPath(repo, ti.Path)
				if err != nil {
					panic(err)
				}
				internal.WriteObject(repo, blob, internal.BlobHeaderName)
			},
			func(ti internal.TreeItem) {
				internal.WriteObject(repo, ti, internal.TreeHeaderName)
			},
		)
		commit := internal.CreateCommit(repo, &tree, "some-message", "")
		hash, err := internal.WriteObject(repo, *commit, internal.CommitHeaderName)
		if err != nil {
			panic(err)
		}
		deserializeCommit := internal.ReadCommit(repo, hash)
		if deserializeCommit.Tree != tree.Hash {
			t.Errorf("Expected to tree hashes be equal")
		}
	})
}
