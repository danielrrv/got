package internal

// "bytes"
// "reflect"
// "strings"

type Commit struct {
	Author      string `object:"author"`
	Committer   string `object:"committer"`
	Tree        string `object:"tree"`
	Date        string `object:"date"`
	Description string `object:"description"`
	Parent      string `object:"parent"`
}

func (c *Commit) Serialize() ([]byte, error) {
	return Serialize(c)
}
func (c *Commit) Deserialize(d []byte) (error) {
	return Deserialize(c, d)
}
