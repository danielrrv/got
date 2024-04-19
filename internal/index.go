package internal

import (
	"bytes"
	"crypto/sha1"
	"encoding/binary"
	"encoding/hex"
	"fmt"

	// "fmt"
	"os"
	"path/filepath"
	"slices"
	"time"
)

// 1. A file is added repository `got add src/cache.rs`, or `got add src/lib.rs src/mod.rs src/readme.md`
// 	 1.1. The files are added to the repository index describing their location, file's content.
//	 1.2. A file, which is already being tracked is modified. The file is found in the index and its modified time and file's content updated.
// 2. The changes are ready to be commited.
//	1.1. A tree is created mapping the real path locations and content of the files.
// 3. A commit will be made with `got commit -m "first commit"`
//  1.1. A commit object is created [see CreateCommit@commit.go],
//	1.2 Each file is persisted on disk each file content compressed
//  1.3. The tree is persisted on disk.
//  1.4. The commit object is persisted on disk

// The index is the same worktree. Only if the tree/blob has changed, then the hash changes. making reference to new tree or blob.
const (
	blockSize = 4 // bytes
)

var (
	// Index signature
	Signature = Byte4{'D', 'I', 'R', 'C'}
	// Index version
	IndexVersion = Byte4{'1', '1', '1', '2'}
)

type Byte4 [blockSize]byte

// uint32 base type
type Bit32 uint32

// representation of 12 bits block.
type Bit12 uint16

// Convert uint32 into 4 bytes of uint8 ByteOrder BigIndian.
func (b Bit32) Bytes() []byte {
	
	rt := make([]byte, 4)
	rt[0] = byte(b >> 24 & 0xFF)
	rt[1] = byte(b >> 16 & 0xFF)
	rt[2] = byte(b >> 8 & 0xFF)
	rt[3] = byte(b & 0xFF)
	return rt
}

// Convert uint16 into 2 bytes of uint8 ByteOrder BigIndian of 13 bits only considered.
func (b Bit12) Bytes() []byte {
	fmt.Println(b)
	rt := make([]byte, 2)
	rt[0] = byte(b >> 8 & 0x0F)
	rt[1] = byte(b & 0xFF)
	return rt
}

// Cast 2 bytes into uint16 type.
func Bit12FromBytes(v []byte) Bit12 {
	if len(v) != 2 {
		panic("bit12 does not have 2 bytes")
	}
	v[0] = byte(v[0] & 0xFF)
	v[1] = byte(v[1])

	return (Bit12)(binary.BigEndian.Uint16(v))
}

// cast 4 bytes into uin32 type
func Bit32FromBytes(v []byte) Bit32 {
	return (Bit32)(binary.BigEndian.Uint32(v))
}

type IndexEntry struct {
	// the last time a file's metadata changed
	Ctime_s Bit32
	// the ctime nanosecond fractions
	Mtime_s Bit32
	// This is the on-disk size from stat(2), truncated to 32-bit.
	FileSize Bit32
	Hash     string //sha1
	PathName string
}

type CacheEntry struct {
	PathName              string
	Hash                  string
	CompressedFileContent []byte
}

type Index struct {
	// Signature of the index.
	Signature Byte4
	// Version of the index.
	Version Byte4
	// Number of entries
	Size Bit32
	// Entries of the index.
	Entries []IndexEntry
	Cache   []CacheEntry
}

func NewIndex() *Index {
	return &Index{
		Signature: Signature,
		Version:   IndexVersion,
		Size:      Bit32(0),
		Entries:   nil,
		Cache:     nil,
	}
}

func Hex2bytes(s string) []byte{
	b, err := hex.DecodeString(s)
	if err !=  nil {
		panic(err)
	} 
	return b
}
func Bytes2hex(d []byte) string{
	return hex.EncodeToString(d)
}

// Convert index non-zero pointer into bytes.
func (i *Index) SerializeIndex() []byte {
	packet := AllocatePacket(0)
	i.Size = Bit32(len(i.Entries))
	//[signature| version | size of entries | entries...[ctime(uint32)|mtime(uint32)|filesize(uint32)|hash|nameLength(uint32)| pathName] ]
	packet.Set(i.Signature[:], i.Version[:], i.Size.Bytes())
	for _, entry := range i.Entries {
		nameLength := Bit12(len(entry.PathName))
		packet.Set(entry.Ctime_s.Bytes(), entry.Mtime_s.Bytes())
		packet.Set(entry.FileSize.Bytes(), Hex2bytes(entry.Hash), nameLength.Bytes(), []byte(entry.PathName))
		packet.Set([]byte{0x00})
	}
	for _, cacheEntry := range i.Cache {
		fileSizeCompressed := Bit12(len(cacheEntry.CompressedFileContent))
		internalPacket := AllocatePacket(0)
		// Construct the cache packet for the entry.
		internalPacket.Set([]byte{0x13})
		//TODO: deserialize this.
		internalPacket.Set([]byte(cacheEntry.PathName), []byte{0x20}, Hex2bytes(cacheEntry.Hash), fileSizeCompressed.Bytes(), cacheEntry.CompressedFileContent)
		internalPacket.Set([]byte{0x00})
		packet.Set(internalPacket.buff)
	}
	return packet.buff
}

// Convert bytes into Index pointer.
func (index *Index) DeserializeIndex(data []byte) {
	if !bytes.Equal(data[0:blockSize], Signature[:]) {
		panic("Invalid index.")
	}
	if !bytes.Equal(data[blockSize:blockSize*2], IndexVersion[:]) {
		panic("Invalid index.")
	}
	index.Signature = Signature
	index.Version = IndexVersion
	sizeOfEntry := Bit32FromBytes(data[blockSize*2 : blockSize*3])
	data = data[blockSize*3:]
	entries := make([]IndexEntry, 0)
	for sizeOfEntry > 0 {
		//Times
		Ctime_s := Bit32FromBytes(data[:blockSize*1])
		Mtime_s := Bit32FromBytes(data[blockSize*1 : blockSize*2])
		//File
		FileSize := Bit32FromBytes(data[blockSize*2 : blockSize*3])
		
		Hash := data[blockSize*3 : (blockSize*3)+sha1.Size]

		//filename length.
	
		nameLength := Bit12FromBytes(data[(blockSize*3)+sha1.Size : (blockSize*3)+sha1.Size+2])
		//34 bytes
		PathName := string(data[(blockSize*3)+sha1.Size+2 : (blockSize*3)+sha1.Size + 2 + nameLength])
		//34 + namelength
		entries = append(entries, IndexEntry{
			Ctime_s:  Ctime_s,
			Mtime_s:  Mtime_s,
			FileSize: FileSize,
			Hash:     Bytes2hex(Hash),
			PathName: PathName,
		})
		sizeOfEntry = sizeOfEntry - 1
		data = data[(blockSize*3+sha1.Size+2+nameLength)+1:]
	}
	cache := make([]CacheEntry, 0)
	//there caches entries.

	if len(data) > 0 {
		// for len(data) > 0 {
		fmt.Println("with cache", len(data), data)
		// 	data = data[1:]
		// }

	} else {

		fmt.Println("without cache", string(data))
	}

	// if len(entries) != int(index.Size) {
	// 	panic("corruption in deserialize index")
	// }
	if len(entries) > 0 {
		index.Entries = slices.Clone(entries)
		index.Size = Bit32(len(entries))
	}
	if len(cache) > 0 {
		index.Cache = slices.Clone(cache)
	}
}

// Read from disk the latest state of the index.
func (i *Index) Refresh(repo *GotRepository) {
	indexContent, err := os.ReadFile(filepath.Join(repo.GotDir, "index"))
	if err != nil {
		panic(err)
	}
	i.DeserializeIndex(indexContent)
}

func (i *Index) Persist(repo *GotRepository) error {
	return CreateOrUpdateRepoFile(repo, "index", i.SerializeIndex())
}

// Add or modify entries in the index.
func (index *Index) AddOrModifyEntries(repo *GotRepository, filePaths []string) {
	// index.Refresh(repo)
	// TODO: empty folder are ignored.
	// TODO: Support add/modify trees, because a file inside of a existing tree[traverseTree] means modify that treeItem and append the blob to that tree.
	for _, fileP := range filePaths {
		possibleBlob, err := BlobFromUserPath(repo, fileP)
		if err != nil {
			panic(err)
		}
		//Index in the db.
		idx := slices.IndexFunc(index.Entries, func(entry IndexEntry) bool {

			return entry.PathName == fileP
		})
		
		//Index in the db.
		cachedIdx := slices.IndexFunc(index.Cache, func(entry CacheEntry) bool {
			return entry.PathName == fileP
		})
	
		// Update only if path exists already and the hash are different.
		if idx >= 0 {
			if possibleBlob.Hash == index.Entries[idx].Hash{
				fmt.Println("Nothing to add. File are the same.")
				return
			}
			// entry := index.Entries[idx]
			index.Entries[idx].Mtime_s = Bit32(time.Now().Unix())
			index.Entries[idx].FileSize = Bit32(len(possibleBlob.FileContent))
			index.Entries[idx].Hash = possibleBlob.Hash
			// No needed but for correctness.
			index.Entries[idx].PathName = fileP
			// Implementation to cache the file and compress its content.
			// After commit the cache will be cleared and only entries(The tracked) files will be preserve.
			var compressedFileContent bytes.Buffer
			Compress(possibleBlob.FileContent, &compressedFileContent)
			// Entry already in cached.
			if cachedIdx >= 0 {
				cacheEntry := index.Cache[cachedIdx]
				cacheEntry.PathName = fileP
				cacheEntry.Hash = possibleBlob.Hash
			} else {
				// Add untracked/modified file to the cache.
				index.Cache = append(index.Cache, CacheEntry{
					PathName:              fileP,
					Hash:                 	possibleBlob.Hash,
					CompressedFileContent: compressedFileContent.Bytes(),
				})
			}
		} else {
			index.Entries = append(index.Entries, IndexEntry{
				Ctime_s:  Bit32(time.Now().Unix()),
				Mtime_s:  Bit32(time.Now().Unix()),
				FileSize: Bit32(len(possibleBlob.FileContent)),
				Hash:     possibleBlob.Hash,
				PathName: fileP,
			})
			// Add untracked/modified file to the cache.
			var compressedFileContent bytes.Buffer
			Compress(possibleBlob.FileContent, &compressedFileContent)
			index.Cache = append(index.Cache, CacheEntry{
				PathName:              fileP,
				Hash:                  possibleBlob.Hash,
				CompressedFileContent: compressedFileContent.Bytes(),
			})
		}

	}
}
