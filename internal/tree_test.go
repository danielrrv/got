package internal_test

import (
	// "bytes"
	// "fmt"
	// "os"
	// "path/filepath"
	"fmt"
	"testing"

	internal "github.com/danielrrv/got/internal"
)

func TestTree(t *testing.T) {
	t.Run("CreateTreeFromFiles", func(t *testing.T) {
		repo, err := internal.FindOrCreateRepo("/home/daniel/got")
		if err != nil {
			t.Errorf("No repo found.")
		}
		files := []string{"test/a/co.txt", "test/a/c/cx.txt", "test/a/b/mx.txt", "test/a/b/jx.txt"}
		m := internal.CreateTreeFromFiles(repo, files)
		fmt.Printf("%v\n",m)
	})
	t.Run("FromMapToTree", func(t *testing.T) {
		repo, err := internal.FindOrCreateRepo("/home/daniel/got")
		if err != nil {
			t.Errorf("No repo found.")
		}
		files := []string{"test/a/co.txt", "test/a/c/cx.txt", "test/a/b/mx.txt", "test/a/b/jx.txt"}
		m := internal.CreateTreeFromFiles(repo, files)
		tree :=internal.FromMapToTree(repo,m, "test")
		internal.TraverseTree(repo, tree)
	})
}
