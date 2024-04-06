package internal_test

import (
	"bytes"
	"fmt"
	"testing"

	internal "github.com/danielrrv/got/internal"
)

func TestConvertBit32ToByte(t *testing.T) {
	t.Run("Convert bi32 to byte", func(t *testing.T) {
		n := 3456787
		v := internal.Bit32(n)
		if !bytes.Equal(v.Bytes(), []byte{0x00, 0x34, 0xBF, 0x13}) {
			t.Errorf("Expected to be equal.")
		}
		if internal.Bit32FromBytes(v.Bytes()) != v {
			t.Errorf("Expected to be equal.")
		}
	})
	t.Run("Serialize/Deserialize index", func(t *testing.T) {
		the_index := internal.Index{
			Signature: internal.Signature,
			Version:   internal.IndexVersion,
			Size:      3,
			Entries: []internal.IndexEntry{
				{Ctime_s: 1712343393,
					Mtime_s:  1712343627,
					FileSize: 23434334,
					Hash:     []byte{0x96, 0xd2, 0xfa, 0x69, 0x73, 0xa9, 0xd6, 0x5f, 0x9a, 0x9f, 0xa1, 0x95, 0xca, 0x9b, 0x07, 0x7e, 0x62, 0xad, 0x79, 0x84},
					PathName: "applicaiton.js",
				},
				{Ctime_s: 1712343393,
					Mtime_s:  1712343627,
					FileSize: 23434334,
					Hash:     []byte{0x96, 0xd2, 0xfa, 0x69, 0x73, 0xa9, 0xd6, 0x5f, 0x9a, 0x9f, 0xa1, 0x95, 0xca, 0x9b, 0x07, 0x7e, 0x62, 0xad, 0x79, 0x84},
					PathName: "file.txt",
				},
				{Ctime_s: 1712343393,
					Mtime_s:  1712343627,
					FileSize: 23434334,
					Hash:     []byte{0x96, 0xd2, 0xfa, 0x69, 0x73, 0xa9, 0xd6, 0x5f, 0x9a, 0x9f, 0xa1, 0x95, 0xca, 0x9b, 0x07, 0x7e, 0x62, 0xad, 0x79, 0x84},
					PathName: "readme.md",
				},
			},
		}
		data := the_index.SerializeIndex()
		fmt.Printf("%v\n", data)
		otherIndex := internal.DeserializeIndex(data)
		// fmt.Println(otherIndex.)
		for _,entry := range otherIndex.Entries{
			fmt.Printf("hash=%v\tpath=%s\n",entry.Hash, entry.PathName)
		}
		
	})
}
