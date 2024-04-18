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

type OFS struct {
	path string
	mode Mode
}

// Location implements GotObject.
func (o OFS) Location() string {
	return o.path
}

// Read the user file.
func (o OFS) Serialize() []byte {
	content, err := os.ReadFile(o.path)
	if err != nil {
		panic(err)
	}
	return content
}

func (t TreeItem) Location() string {
	return t.Path
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
func (t *TreeItem) TraverseTree(visitBlob func(TreeItem), visitTree func(TreeItem)) {
	if bytes.Equal(t.Mode, TreeMode) {
		visitTree(*t)
	}
	for _, item := range t.Children {
		if bytes.Equal(item.Mode, BlobMode) {
			visitBlob(item)
		}
		if bytes.Equal(item.Mode, TreeMode) {
			visitTree(item)
			for _, child := range item.Children {
				child.TraverseTree(visitBlob, visitTree)
			}
		}
	}
}

// Flatten the tree to linear structure of blobs.
func (t *TreeItem) FlatItems() []TreeItem {
	ret := make([]TreeItem, 0)
	t.TraverseTree(func(ti TreeItem) {
		ret = append(ret, ti)
	}, func(ti TreeItem) {})
	return ret
}

// Convert map of OFS into TreeItem graph. Intermediate converter.
func FromMapToTree(repo *GotRepository, m map[string][]OFS, parent string) TreeItem {
	items := m[parent]
	re := make([]TreeItem, 0)
	for _, item := range items {
		// Branch #1: The item is blob. Just create the in-memory object and append.
		if bytes.Equal(item.mode, BlobMode) {
			//possible hash of a OFS blob must be equal to the actual blob.
			hash, err := CreatePossibleObjectFromData(repo, item, BlobHeaderName)
			if err != nil {
				panic(err)
			}
			re = append(re, TreeItem{
				Path:     item.path,
				Hash:     hash,
				Mode:     BlobMode,
				Children: nil,
			})
			continue
		}
		// Branch #2: the item is tree. Keep drill down recursively the graph.
		if bytes.Equal(item.mode, TreeMode) {
			re = append(re, FromMapToTree(repo, m, item.path))
		}
	}
	// Based parent tree. 
	t := TreeItem{
		Path:     parent,
		Mode:     TreeMode,
		Hash:     "",
		Children: re,
	}
	//Create in-memory object of the tree.
	hash, err := CreatePossibleObjectFromData(repo, t, TreeHeaderName)
	if err != nil {
		panic(err)
	}
	t.Hash = hash
	return t
}

// Create map of OFS from array of files.
func CreateTreeFromFiles(repo *GotRepository, files []string) map[string][]OFS {
	m := make(map[string][]OFS)
	for _, wholePath := range files {
		// Split by file system separator. Not MS.Window tested.
		dirs := strings.Split(wholePath, string(filepath.Separator))
		for i := len(dirs) - 1; i > 0; i-- {
			if ok, err := isFile(filepath.Join(repo.GotTree, filepath.Join(dirs[0:i]...), dirs[i])); ok {
				if err != nil {
					panic(err)
				}
				if indexOf(m[dirs[i-1]], dirs[i]) == -1 {
					m[dirs[i-1]] = append(m[dirs[i-1]], OFS{path: filepath.Join(repo.GotTree, wholePath), mode: BlobMode})
				}
			} else {
				if indexOf(m[dirs[i-1]], dirs[i]) == -1 {
					m[dirs[i-1]] = append(m[dirs[i-1]], OFS{path: filepath.Join(repo.GotTree, filepath.Join(dirs[0:i]...)), mode: TreeMode})
				}
			}
		}
	}
	return m
}

// Determine whether or not the path is file.
func isFile(path string) (bool, error) {
	fi, err := os.Stat(path)
	if err != nil {
		return false, err
	}
	if fi.IsDir() {
		return false, nil
	}
	return true, nil
}

// Convert TreeItem into []bytes. No recursive.
func (t TreeItem) Serialize() []byte {
	bb := make([]byte, 0)
	// What it does: Sort the path so that the hash of the tree with the same item but different order give the same hash.
	slices.SortFunc(t.Children, func(a, b TreeItem) int {
		return cmp.Compare(a.Path, b.Path)
	})
	// [mode of 4 bytes]|[space with 0x20]|[path no limit]|[terminator 0x00]|[sha1 of 20 bytes]
	for _, buf := range t.Children {
		bb = append(bb, buf.Mode...)
		bb = append(bb, byte(0x20))
		bb = append(bb, []byte(buf.Path)...) //Must be full path from got tree.
		bb = append(bb, byte(0x00))
		bb = append(bb, []byte(Hex2bytes(buf.Hash))...)
	}
	return bb
}

// Deserialize raw bytes to TreeItem struct.
func (t TreeItem) Deserialize(d []byte) TreeItem {

	for len(d) > 0 {
		// The Mode separator 0x20.
		modeSep := bytes.Index(d, []byte{0x20})
		// The path terminator 0x00
		pathTerm := bytes.Index(d, []byte{0x00})
		//Implementation to deserialize a blob.
		if bytes.Equal(d[0:modeSep], BlobMode) {
			t.Children = append(t.Children, TreeItem{
				//Mode[0, 0x20]
				Mode: d[0:modeSep],
				//Path[0x20 + 1, 0x00]
				Path: string(d[modeSep+1 : pathTerm]),
				//Hash[0x00, 0x00  + sha1.Size(20 bytes)]
				Hash:     string(Bytes2hex(d[pathTerm+1 : sha1.Size+pathTerm+1])),
				Children: nil,
			})
			// Discard consumed bytes pathTermIndex + 20bytes(sha1) + 1(The skipped 0x00)
			d = d[pathTerm+sha1.Size+1:]
			//Implementation to skip going through tree when already consumed a treeItem.
			continue
		}
		// Implementation to deserialize a tree.
		if bytes.Equal(d[0:modeSep], TreeMode) {
			// Find the a possible repository.
			//TODO: Find a way to get the real repository.
			repo, err := FindOrCreateRepo(string(d[modeSep+1 : pathTerm]))
			if err != nil {
				panic(err)
			}
			// Read the tree from DB based on the hash of it.
			content, err := ReadObject(repo, TreeHeaderName, string(Bytes2hex(d[pathTerm+1:sha1.Size+pathTerm+1])))
			if err != nil {
				panic(err)
			}
			// Append the children of this if exist.
			if len(content) > 0 {
				t.Children = append(t.Children, t.Deserialize(content))
			}
			// Discard consumed bytes pathTermIndex + 20bytes(sha1) + 1(The skipped 0x00)
			d = d[pathTerm+1+sha1.Size:]
		}
	}
	return t
}
