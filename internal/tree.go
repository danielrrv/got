package internal

import (
	"bytes"
	"cmp"
	"crypto/sha1"
	"errors"
	"os"
	"path/filepath"
	"slices"
)

type Mode []byte

var (
	BLOB_MODE Mode = []byte{0x31, 0x30, 0x30, 0x36, 0x34, 0x34} //100644
	TREE_MODE Mode = []byte{0x30, 0x34, 0x30, 0x30, 0x30, 0x30} //040000
)

type TreeItem struct {
	mode Mode
	path string
	hash string
}

type Tree struct {
	Buffer []TreeItem
	Size   int
}

func readTree(hash string) {

}

func parseTree(data []byte) (*Tree, error) {
	size := len(data)
	items := make([]TreeItem, 0)
	if size == 0 {
		return nil, errors.New("nothing to parse")
	}
	for size > 0 {
		mode := data[:len(BLOB_MODE)]
		if bytes.Compare(mode, TREE_MODE) == 0 {
			// Read the tree
			// Parse the tree and adds the treeItems to the items.
		}
		pathSize := data[6]
		path := data[7:int(pathSize)]
		hash := data[7+int(pathSize)+1:]
		items = append(items, TreeItem{
			mode: mode,
			path: string(path),
			hash: string(hash),
		})
		size -= len(mode) + 1 + int(pathSize) + sha1.Size
		data = data[len(mode)+1+int(pathSize)+sha1.Size+1:]
	}
	return &Tree{
		Buffer: items,
		Size:   len(items),
	}, nil
}

func (t *Tree) ToByteBuffer() []byte {
	bb := make([]byte, 0)
	slices.SortFunc(t.Buffer, func(a, b TreeItem) int {
		return cmp.Compare(a.path, b.path)
	})
	for _, buf := range t.Buffer {
		bb = append(bb, buf.mode...)
		bb = append(bb, byte(len(buf.path)))
		bb = append(bb, []byte(buf.path)...)
		bb = append(bb, []byte(buf.hash)...)
	}
	return bb
}

func (t *Tree) writeTree(repo *GotRepository, hash string) error {
	var bb bytes.Buffer
	TryCreateFolderIn(filepath.Join(repo.GotDir, gotRepositoryDirObjects), string(hash[:2]))
	file, err := os.OpenFile(filepath.Join(repo.GotDir, gotRepositoryDirObjects, string(hash[:2]), string(hash[2:])), os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer file.Close()
	Compress(t.ToByteBuffer(), &bb)
	file.Write(bb.Bytes())
	return nil
}

func isTree(repo *GotRepository, hash string) bool {
	var t Tree
	err := ReadObject(repo, &t, "tree", hash)
	if err != nil {
		panic(err)
	}
}
