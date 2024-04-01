package internal_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	internal "github.com/danielrrv/got/internal"
)

const (
	ROOT_DIR = "/home/daniel/got"
)

func TestSerialize(t *testing.T) {
	t.Run("Serialize/Serialize", func(t *testing.T) {
		commit := internal.Commit{
			Author:      "Daniel",
			Committer:   "Daniel Rodirguez",
			Tree:        "3456787654334567",
			Description: "Some beuatiful day",
			Date:        "25-05-2023",
			Parent:      "34567876543",
		}

		bt, err := commit.Serialize()
		if err != nil {
			t.Fatalf("Expected serialized to be bytes, %v", err.Error())
		}
		commit2 := new(internal.Commit)
		err = commit2.Deserialize(bt)
		if err != nil {
			t.Fatalf("Expected deserialized back to Commit %v", err.Error())
		}
		if commit2.Author != commit.Author {
			t.Fatalf("Expected commit2 to be equal to commit %s!=%s", commit.Author, commit2.Author)
		}
	})
	t.Run("Write a commit object", func(t *testing.T) {
		commit := internal.Commit{
			Author:      "Daniel",
			Committer:   "Daniel Rodirguez",
			Tree:        "3456787654334567",
			Description: "Some beuatiful day",
			Date:        "25-05-2023",
			Parent:      "34567876543",
		}
		repo, err := internal.FindOrCreateRepo(ROOT_DIR)
		if err != nil {
			t.Errorf("No repo found.")
		}
		hash, err := internal.WriteObject(repo, &commit, "commit")
		if err != nil {
			t.Errorf("no object written, %v", err.Error())
		}
		err = internal.RemoveObjectFrom(repo, hash)
		if err != nil {
			t.Errorf("unable to remove the created object.")
		}
		err = os.Remove(filepath.Join(repo.GotDir, "objects", hash[:2]))
		if err != nil {
			t.Errorf("unable to remove the created object.")
		}
	})
	t.Run("Read a commit object", func(t *testing.T) {
		// t.FailNow()
		commit := internal.Commit{
			Author:      "Daniel",
			Committer:   "Daniel Rodirguez",
			Tree:        "3456787654334567",
			Description: "Some beuatiful day",
			Date:        "25-05-2023",
			Parent:      "34567876543",
		}
		repo, err := internal.FindOrCreateRepo("/home/daniel/got")
		if err != nil {
			t.Errorf("No repo found.")
		}
		hash, err := internal.WriteObject(repo, &commit, "commit")
		if err != nil {
			t.Errorf("no object written, %v", err.Error())
		}
		commit2:= new(internal.Commit)
		err = internal.ReadObject(repo, commit2, "commit", hash)
		if err != nil {
			t.Errorf("%v", err.Error())
		}
		if commit2.Date !=commit.Date{
			t.Errorf("Expected to commit2.Date equals to commit.Date")
		}
		err = internal.RemoveObjectFrom(repo, hash)
		if err != nil {
			t.Errorf("unable to remove the created object.")
		}
		err = os.Remove(filepath.Join(repo.GotDir, "objects", hash[:2]))
		if err != nil {
			t.Errorf("unable to remove the created object.")
		}
	})

	t.Run("compress", func(t *testing.T) {
		bb := []byte("Hello worldddddddddddd")
		var cc bytes.Buffer
		internal.Compress(bb, &cc)
		if len(cc.Bytes()) == len(bb) {
			t.Errorf("Expected to coppu the bytes")
		}
	})
	t.Run("decompress", func(t *testing.T) {
		bb := []byte("Hello worldddddddddddd")
		var cc bytes.Buffer
		internal.Compress(bb, &cc)
		bbb := make([]byte, 0, cc.Len())
		bbb = append(bbb, cc.Bytes()...)
		var ccc bytes.Buffer
		internal.Decompress(bbb, &ccc)
		if len(cc.Bytes()) == len(bb) {
			t.Errorf("Expected to coppu the bytes")
		}
	})

}
