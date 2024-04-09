package internal

import (
	// "bytes"/
	"bytes"
	"crypto/sha1"
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	// "os"
	// "slices"
)

// import "fmt"

const (
	blockSize = 4 // bytes
)

var (
	Signature    = Byte4{'D', 'I', 'R', 'C'}
	IndexVersion = Byte4{'1', '0', '0'}
)

type Byte4 [blockSize]byte
type Bit32 uint32
type Bit12 uint16


func (b Bit32) Bytes() []byte {
	rt := make([]byte, 4)
	rt[0] = byte(b >> 24 & 0xFF)
	rt[1] = byte(b >> 16 & 0xFF)
	rt[2] = byte(b >> 8 & 0xFF)
	rt[3] = byte(b & 0xFF)
	return rt
}

func (b Bit12) Bytes() []byte {
	rt := make([]byte, 2)
	rt[0] = byte(b >> 8 & 0x0F)
	rt[1] = byte(b & 0xFF)
	return rt
}

func Bit12FromBytes(v []byte) Bit12 {
	if len(v) != 2 {
		panic("bit12 does not have 2 bytes")
	}
	v[0] = byte(v[0] & 0xFF)
	v[1] = byte(v[1])
	return (Bit12)(binary.BigEndian.Uint16(v))
}

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
	Hash     []byte //sha1
	PathName string
}

type Index struct {
	Signature Byte4
	Version   Byte4
	Size      Bit32
	Entries   []IndexEntry
}

var TheIndex *Index

func init() {

}

func (i *Index) SerializeIndex() []byte {
	packet := AllocatePacket(1024 * 100)
	packet.Set(i.Signature[:], i.Version[:], i.Size.Bytes())
	for _, entry := range i.Entries {
		nameLength := Bit12(len(entry.PathName))
		packet.Set(entry.Ctime_s.Bytes(), entry.Mtime_s.Bytes())
		packet.Set(entry.FileSize.Bytes(), entry.Hash, nameLength.Bytes(), []byte(entry.PathName))
		packet.Set([]byte{0x00})
	}
	return packet.buff
}

func DeserializeIndex(data []byte) *Index {

	if !bytes.Equal(data[0:blockSize], Signature[:]) {
		panic("Invalid index.")
	}
	if !bytes.Equal(data[blockSize:blockSize*2], IndexVersion[:]) {
		panic("Invalid index.")
	}
	index := new(Index)
	index.Signature = Signature
	index.Version = IndexVersion
	sizeOfEntry := Bit32FromBytes(data[blockSize*2 : blockSize*3])
	data = data[blockSize*3:]
	index.Size = sizeOfEntry
	// os.Exit(1)
	entries := make([]IndexEntry, 0)
	for sizeOfEntry > 0 {
		//Times
		Ctime_s := Bit32FromBytes(data[:blockSize*1])
		Mtime_s := Bit32FromBytes(data[blockSize*1 : blockSize*2])
		//File
		FileSize := Bit32FromBytes(data[blockSize*2 : blockSize*3])
		// sizeOfEntry -= blockSize * 6
		Hash := data[blockSize*3 : (blockSize*3)+sha1.Size]

		//filename length.
		nameLength := Bit12FromBytes(data[(blockSize*3)+sha1.Size : (blockSize*3)+sha1.Size+2])
		//34 bytes
	
		PathName := string(data[(blockSize*3)+sha1.Size+2 : (blockSize*3)+sha1.Size+2+nameLength])
		//34 + namelength
		entries = append(entries, IndexEntry{
			Ctime_s:  Ctime_s,
			Mtime_s:  Mtime_s,
			FileSize: FileSize,
			Hash:     Hash,
			PathName: PathName,
		})
		sizeOfEntry = sizeOfEntry - 1
		fmt.Println(PathName)
		data = data[(blockSize*3+sha1.Size+2+nameLength)+1:]
		fmt.Println(data)
	}
	if len(entries) != int(index.Size) {
		panic("corruption in deserialize index")
	}
	index.Entries = append(index.Entries, entries...)
	// copy(index.Entries, entries)
	return index
}

func (i *Index) parseIndex(data []byte) {
	i.Signature = Signature
	i.Version = IndexVersion
}

func (i *Index) refresh(repo *GotRepository) {
	indexContent, err := os.ReadFile(filepath.Join(repo.GotDir, "index"))
	if err != nil {
		panic(err)
	}
	i.parseIndex(indexContent)

}
