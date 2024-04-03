package internal

import (
	"bytes"
	"cmp"
	"crypto/sha1"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	// "os"
	// "path/filepath"
	"slices"
)

type Mode []byte

var (
	BLOB_MODE Mode = []byte{0x31, 0x30, 0x30, 0x36, 0x34, 0x34} //100644
	TREE_MODE Mode = []byte{0x30, 0x34, 0x30, 0x30, 0x30, 0x30} //040000
)

func (m *Mode) String() string {
	switch m {
	case &BLOB_MODE:
		return "blob"
	case &TREE_MODE:
		return "tree"
	default:
		panic("No conversion type")
	}
}

type TreeItem struct {
	Mode Mode
	// User file location
	Path string
	// Blob hash
	Hash     string
	Children []*TreeItem
}

type Tree struct {
	Hash   string
	Buffer []TreeItem
	Size   int
}

// func treeFromSubFolders(repo *GotRepository, path string) *Tree {
// 	// /repo/a/b/c.txt, /repo/a/b/mx.txt /repo/a/c/mt.txt

// 	// if isDir(path){
// 	// 	tree := Tree{
// 	// 		Buffer: []TreeItem{{mode: TREE_MODE, path:path, hash: }},
// 	// 	}
// 	// }

// // 	if isFile(path) {
// // 		content, err := os.ReadFile(filepath.Join(repo.GotTree, filepath.Join(dirs...)))
// // 		if err != nil {
// // 			panic(err)
// // 		}
// // 		hash, err := CreatePossibleObjectFromData(repo, content, BLOB_HEADER_NAME)
// // 		if err != nil {
// // 			panic(err)
// // 		}
// // 		return &Tree{
// // 			Buffer: []TreeItem{{mode: BLOB_MODE, path: filepath.Base(path), hash: hash}},
// // 			Size:   1,
// // 		}
// // 	}
// // }

type OFS struct {
	path string
	mode Mode
}

func (o *OFS) String() {
	fmt.Printf("%s,%s", o.path, string(o.mode))
}

func indexOf(offs []OFS, key string) int {
	for i, of := range offs {
		if of.path == key {
			return i
		}
	}
	return -1
}

func TraverseTree(repo *GotRepository, t *TreeItem) {
	fmt.Printf("%s\t%s\t%s\n", t.Mode, t.Path, t.Hash)
	for _, item := range t.Children {
		if bytes.Equal(item.Mode, BLOB_MODE) {
			fmt.Printf("%s\t%s\t%s\n", item.Mode, item.Path, item.Hash)
		}
		if bytes.Equal(item.Mode, TREE_MODE) {
			fmt.Printf("%s\t%s\t%s\n", item.Mode, item.Path, item.Hash)
			for _, child := range item.Children {
				TraverseTree(repo, child)
			}
		}
	}
}

func FromMapToTree(repo *GotRepository, m map[string][]OFS, parent string) *TreeItem {
	items := m[parent]
	re := make([]*TreeItem, 0)
	for _, item := range items {
		if bytes.Equal(item.mode, BLOB_MODE) {
			bb := []byte{0x34, 0x34, 0x34}
			hash, err := CreatePossibleObjectFromData(repo, bb, blobHeaderName)
			if err != nil {
				panic(err)
			}
			re = append(re, &TreeItem{
				Path:     item.path,
				Hash:     hash,
				Mode:     BLOB_MODE,
				Children: nil,
			})
		}
		if bytes.Equal(item.mode, TREE_MODE) {
			re = append(re, FromMapToTree(repo, m, item.path))
		}
	}
	
	t := TreeItem{
		Path:     parent,
		Mode:     TREE_MODE,
		Hash:     "",
		Children: re,
	}

	hash, err := CreatePossibleObjectFromData(repo, t.ToByteBuffer(), treeHeaderName)
	if err != nil {
		panic(err)
	}
	t.Hash = hash
	return &t
}

func CreateTreeFromFiles(repo *GotRepository, files []string) map[string][]OFS {
	m := make(map[string][]OFS)
	for _, wholePath := range files {
		dirs := strings.Split(wholePath, string(filepath.Separator))
		for i := len(dirs) - 1; i > 0; i-- {
			if isFile(filepath.Join(repo.GotTree, filepath.Join(dirs[0:i]...), dirs[i])) {
				if indexOf(m[dirs[i-1]], dirs[i]) == -1 {
					m[dirs[i-1]] = append(m[dirs[i-1]], OFS{path: dirs[i], mode: BLOB_MODE})
				}
			} else {
				if indexOf(m[dirs[i-1]], dirs[i]) == -1 {
					m[dirs[i-1]] = append(m[dirs[i-1]], OFS{path: dirs[i], mode: TREE_MODE})
				}
			}
		}
	}
	return m
}

func isFile(path string) bool {
	fi, err := os.Stat(path)
	// fmt.Println(path,fi.IsDir())
	if err != nil {
		return false
	}
	if fi.IsDir() {
		return false
	}
	return true
}

func ParseTree(repo *GotRepository, data []byte) ([]TreeItem, error) {
	size := len(data)
	items := make([]TreeItem, 0)
	if size == 0 {
		return nil, errors.New("nothing to parse")
	}
	for size > 0 {
		mode := data[:len(BLOB_MODE)]
		pathSize := data[len(BLOB_MODE)]
		path := data[len(BLOB_MODE)+1 : int(pathSize)]
		hash := data[len(BLOB_MODE)+1+int(pathSize)+1:]
		if bytes.Equal(mode, TREE_MODE) {
			obj, err := ReadObject(repo, treeHeaderName, string(hash))
			if err != nil {
				fmt.Printf("unable to read tree, %v", err.Error())
			} else {
				treeItems, err := ParseTree(repo, obj)
				if err != nil {
					items = append(items, treeItems...)
				}
			}
		}
		items = append(items, TreeItem{
			Mode: mode,
			Path: string(path),
			Hash: string(hash),
		})
		size -= len(mode) + 1 + int(pathSize) + sha1.Size
		data = data[len(mode)+1+int(pathSize)+sha1.Size+1:]
	}
	return items, nil
}

func (t *Tree) ToByteBuffer() []byte {
	bb := make([]byte, 0)
	slices.SortFunc(t.Buffer, func(a, b TreeItem) int {
		return cmp.Compare(a.Path, b.Path)
	})
	for _, buf := range t.Buffer {
		bb = append(bb, buf.Mode...)
		bb = append(bb, byte(len(buf.Path)))
		bb = append(bb, []byte(buf.Path)...)
		bb = append(bb, []byte(buf.Hash)...)
	}
	return bb
}

func (t *TreeItem) ToByteBuffer() []byte {
	bb := make([]byte, 0)
	slices.SortFunc(t.Children, func(a, b *TreeItem) int {
		return cmp.Compare(a.Path, b.Path)
	})
	for _, buf := range t.Children {
		bb = append(bb, buf.Mode...)
		bb = append(bb, byte(len(buf.Path)))
		bb = append(bb, []byte(buf.Path)...)
		bb = append(bb, []byte(buf.Hash)...)
	}
	return bb
}

func (t *Tree) Persist(repo *GotRepository, hash string) (string, error) {
	return WriteObject(repo, t.ToByteBuffer(), treeHeaderName)
}
