package internal_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/danielrrv/got/internal"
)

func TestReference(t *testing.T) {
	t.Run("ReferenceFromHEAD when is a ref.", func(t *testing.T) {
		tmp := t.TempDir()
		repo, err := internal.FindOrCreateRepo(tmp)
		if err != nil {
			panic(err)
		}
		content, err := os.ReadFile(filepath.Join(repo.GotDir, "HEAD"))
		
		if err != nil {
			t.Errorf("Expected to read the file, %v", err.Error())
		}
		ref := internal.ReferenceFromHEAD(repo, content)
		fmt.Println(ref)
	})
	t.Run("ReferenceFromHEAD when is direct hash ref", func(t *testing.T) {
		tmp := t.TempDir()
		repo, err := internal.FindOrCreateRepo(tmp)
		ref := internal.Ref{
			IsDirect:  true,
			Reference: "ffffaffffaffffaffffaffffaffffaffffaffffa",
			Invalid:   true,
		}
		ref.WriteRef(repo)
		if err != nil {
			panic(err)
		}
		content, err := os.ReadFile(filepath.Join(repo.GotDir, "HEAD"))
		
		if err != nil {
			t.Errorf("Expected to read the file, %v", err.Error())
		}
		ref2 :=internal.ReferenceFromHEAD(repo, content)
		if !ref2.Invalid {
			t.Errorf("Expected to be invalid")
		}
		possibleRefPath := filepath.Join(repo.GotDir, "objects", "ff")
		err = os.MkdirAll(possibleRefPath, 0755)
		if err != nil {
			panic(err)
		}
		err = internal.CreateOrUpdateRepoFile(repo, filepath.Join("objects", "ff", "ffaffffaffffaffffaffffaffffaffffaffffa"), []byte("some object data"))
		if err != nil {
			t.Errorf("Expected to create referered object, %v", err.Error())
		}
		content, err = os.ReadFile(filepath.Join(repo.GotDir, "HEAD"))
		
		if err != nil {
			t.Errorf("Expected to read the file, %v", err.Error())
		}
		ref3 :=internal.ReferenceFromHEAD(repo, content)
		if ref3.Invalid {
			t.Errorf("Expected to be valid")
		}
	})
	t.Run("ReferenceFromHEAD when is indirect hash ref", func(t *testing.T) {
		tmp := t.TempDir()
		refString := "ref: refs/heads/the-branch"
		repo, err := internal.FindOrCreateRepo(tmp)
		ref := internal.Ref{
			IsDirect:  true,
			Reference: refString,
			Invalid:   true,
		}
		ref.WriteRef(repo)

		if err != nil {
			panic(err)
		}

		content, err := os.ReadFile(filepath.Join(repo.GotDir, "HEAD"))
		
		if err != nil {
			t.Errorf("Expected to read the file, %v", err.Error())
		}

		ref2 := internal.ReferenceFromHEAD(repo, content)
		if !ref2.Invalid {
			t.Errorf("Expected to be invalid")
		}
		if ref2.Invalid != true && ref2.IsDirect != false && ref2.Reference != refString{
			t.Errorf("Expected reference valid but indirect.")
		}
		
		possibleRefPath := filepath.Join(repo.GotDir, "refs", "heads")
		
		err = os.MkdirAll(possibleRefPath, 0755)
		if err != nil {
			panic(err)
		}
		someHash := "some-hash"
		err = internal.CreateOrUpdateRepoFile(repo, filepath.Join("refs", "heads", "the-branch"), []byte(someHash))
		if err != nil {
			t.Errorf("Expected to create referered object, %v", err.Error())
		}
		content, err = os.ReadFile(filepath.Join(repo.GotDir, "HEAD"))
		
		if err != nil {
			t.Errorf("Expected to read the file, %v", err.Error())
		}
		ref3 := internal.ReferenceFromHEAD(repo, content)

		fmt.Println(ref3)
		if ref3.Reference != someHash {
			t.Errorf("Expected to have read the reference of the reference.")
		}
	})

}
