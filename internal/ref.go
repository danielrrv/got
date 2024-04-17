package internal

import (
	"crypto/sha1"
	"fmt"
	"os"
	"path/filepath"
	"regexp"


)

var (
	//Maximun 16 characters for branch names. No validation so far.
	refRegex = regexp.MustCompile(`(^ref: )(refs/heads/[a-zA-Z-]{1,16}[/]?[a-zA-Z-]{1,16})`)
)

type Ref struct {
	Invalid   bool
	IsDirect  bool
	Reference string
}

func (r *Ref) WriteRef(repo *GotRepository) error {
	// What it does: When the reference is indirect, this reference points to a branch on the ref/heads/ folder.
	if !r.IsDirect {
		r.Reference = fmt.Sprintf("ref: refs/heads/%s", r.Reference)
	}
	return CreateOrUpdateRepoFile(repo, "HEAD", []byte(fmt.Sprintf(r.Reference)))
}

func ReferenceFromHEAD(repo *GotRepository, referenceData []byte) *Ref {
	// Implementation to determine the ref is indirect. So validate the existance of it. Otherwise is first commit.
	if matchGroup := refRegex.FindAllStringSubmatch(string(referenceData), -1); matchGroup != nil {
		if len(matchGroup[0]) >= 3 {
			refPath := matchGroup[0][2]
			content, err := os.ReadFile(filepath.Join(repo.GotDir, refPath))
			if err != nil {
				//Invalidate beucase reading the refs/heads/{ref-branch} failed.
				return &Ref{
					Invalid:   true,
					IsDirect:  false,
					Reference: string(referenceData),
				}
			}
			//Find the refs/heads/{ref-branch} has content.
			return ReferenceFromHEAD(repo, content)
		}
	} else {
		if len(referenceData) == sha1.Size*2 {
			_, err := os.Stat(filepath.Join(repo.GotDir, "objects", string(referenceData[:2]), string(referenceData[2:])))
			if err != nil {
				// - Invalidate beucase reading the file failed.
				// - The reference is the hash.
				return &Ref{
					Invalid:   true,
					IsDirect:  true,
					Reference: string(referenceData),
				}
			}
			// - the object(commit) exists and the reference is a hash.
			return &Ref{
				Invalid:   false,
				IsDirect:  true,
				Reference: string(referenceData),
			}
		}else{
			return &Ref{
				Invalid: true,
				IsDirect: true,
				Reference: string(referenceData),
			}
		}
	}
	return nil
}
