package internal_test

import (
	"testing"

	internal "github.com/danielrrv/got/internal"
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

}
