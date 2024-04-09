package internal

import (
	"bytes"
	"cmp"
	"crypto/sha1"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

type Mode []byte

var (
	BlobMode Mode = []byte{0x31, 0x30, 0x30, 0x36, 0x34, 0x34} //100644
	TreeMode Mode = []byte{0x30, 0x34, 0x30, 0x30, 0x30, 0x30} //040000

	ErrorCorruptedData = errors.New("invalid object persistance. Temporal hash isn't final hash")
)

func (m Mode) String() string {
	switch (string)(m) {
	case string(BlobMode):
		return `blob`
	case string(TreeMode):
		return `tree`
	default:
		panic("No conversion type.")
	}
}

type TreeItem struct {
	Mode Mode
	// User file location
	Path string
	// Blob hash
	Hash     string
	Children []TreeItem
}

type Tree struct {
	Hash   string
	Buffer []TreeItem
	Size   int
}

type OFS struct {
	path string
	mode Mode
}

func (o OFS) Serialize() []byte {
	content, err := os.ReadFile(o.path)
	if err != nil {
		panic(err)
	}
	return content
}

func (o *OFS) String() {
	fmt.Printf("%s,%s", o.path, string(o.mode))
}

// Find index of OFS element based on path-key.
func indexOf(offs []OFS, key string) int {
	for i, of := range offs {
		if of.path == key {
			return i
		}
	}
	return -1
}

// Traverse the tree graph.
func TraverseTree(repo *GotRepository, t TreeItem, visitBlob func(TreeItem), visitTree func(TreeItem)) {
	visitTree(t)
	for _, item := range t.Children {
		if bytes.Equal(item.Mode, BlobMode) {
			visitBlob(item)
		}
		if bytes.Equal(item.Mode, TreeMode) {
			visitTree(item)
			for _, child := range item.Children {
				TraverseTree(repo, child, visitBlob, visitTree)
			}
		}
	}
}

// Convert map of OFS into TreeItem graph. Intermediate converter.
func FromMapToTree(repo *GotRepository, m map[string][]OFS, parent string) TreeItem {
	items := m[parent]
	re := make([]TreeItem, 0)
	for _, item := range items {
		if bytes.Equal(item.mode, BlobMode) {
			hash, err := CreatePossibleObjectFromData(repo, item, blobHeaderName)
			if err != nil {
				panic(err)
			}
			re = append(re, TreeItem{
				Path:     item.path,
				Hash:     hash,
				Mode:     BlobMode,
				Children: nil,
			})
		}
		// If item is tree, them keep traversing.
		if bytes.Equal(item.mode, TreeMode) {
			re = append(re, FromMapToTree(repo, m, item.path))
		}
	}

	t := TreeItem{
		Path:     parent,
		Mode:     TreeMode,
		Hash:     "",
		Children: re,
	}
	//Create hash of the tree.
	hash, err := CreatePossibleObjectFromData(repo, t, treeHeaderName)
	if err != nil {
		panic(err)
	}
	t.Hash = hash
	return t
}

// Create map if OFS from array of files.
func CreateTreeFromFiles(repo *GotRepository, files []string) map[string][]OFS {
	m := make(map[string][]OFS)
	for _, wholePath := range files {
		// SPlit by file system separator. Not MS.Window tested.
		dirs := strings.Split(wholePath, string(filepath.Separator))
		for i := len(dirs) - 1; i > 0; i-- {
			if isFile(filepath.Join(repo.GotTree, filepath.Join(dirs[0:i]...), dirs[i])) {
				if indexOf(m[dirs[i-1]], dirs[i]) == -1 {
					m[dirs[i-1]] = append(m[dirs[i-1]], OFS{path: dirs[i], mode: BlobMode})
				}
			} else {
				if indexOf(m[dirs[i-1]], dirs[i]) == -1 {
					m[dirs[i-1]] = append(m[dirs[i-1]], OFS{path: dirs[i], mode: TreeMode})
				}
			}
		}
	}
	return m
}

// Determine whether or not the path is file.
func isFile(path string) bool {
	fi, err := os.Stat(path)
	if err != nil {
		return false
	}
	if fi.IsDir() {
		return false
	}
	return true
}


// Convert TreeItem into []bytes. No recursive.
func (t TreeItem) Serialize() []byte {
	bb := make([]byte, 0)
	slices.SortFunc(t.Children, func(a, b TreeItem) int {
		return cmp.Compare(a.Path, b.Path)
	})
	for _, buf := range t.Children {
		bb = append(bb, buf.Mode...)
		bb = append(bb, byte(0x20))
		bb = append(bb, []byte(buf.Path)...)
		bb = append(bb, byte(0x00))
		bb = append(bb, []byte(buf.Hash)...)
	}
	return bb
}

// Deserialize raw bytes to TreeItem struct.
func (t TreeItem) Deserialize(d []byte) TreeItem {
	for len(d) > 0 {
		modeSep := bytes.Index(d, []byte{0x20})
		pathTerm := bytes.Index(d, []byte{0x00})

		if bytes.Equal(d[0:modeSep], BlobMode) {
			t.Children = append(t.Children, TreeItem{
				Mode:     d[0:modeSep],
				Path:     string(d[modeSep+1 : pathTerm]),
				Hash:     string(d[pathTerm+1 : sha1.Size]),
				Children: nil,
			})
			d = d[pathTerm+sha1.Size:]
		}
		if bytes.Equal(d[0:modeSep], TreeMode) {
			repo, err := FindOrCreateRepo(string(d[modeSep+1 : pathTerm]))
			if err != nil {
				panic(err)
			}
			content, err := ReadObject(repo, treeHeaderName, string(d[pathTerm+1:sha1.Size]))
			if err != nil {
				panic(err)
			}
			t.Children = append(t.Children, t.Deserialize(content))
		}
	}
	return t
}
