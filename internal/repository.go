package internal

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	// "slices"
)

const (
	gotRepositoryDirRefs      = "refs"
	gotRepositoryDirRefsTags  = "tags"
	gotRepositoryDirRefsHeads = "heads"
	gotRepositoryDirObjects   = "objects"
	gotRootRepositoryDir      = ".got"
	version                   = "v1.0.0"
)

var (
	ErrorPathInvalid            = errors.New("path is invalid")
	ErrorPathDoesNotExist       = errors.New("path does not exist")
	ErrorRepositoryDoesNotExist = errors.New("repository does not exist")
	ErrorLoadConfig             = errors.New("unable to load the configuration")
	ErrorOpeningFile            = errors.New("unable to find the file in the repo")
)
var (
	VersionRegex = regexp.MustCompile("^version")
)

type GotConfig struct{}

// https://github.com/atom/git-utils/blob/master/src/repository.h
type GotRepository struct {
	// The root folder of the repository.
	GotTree string
	// The repository configurations.
	GotConfig GotConfig
	// .got folder name.
	GotDir string
	// Temporary database.
	Index *Index
}

// Create a file inside of the repo dir(.got)
func CreateOrUpdateRepoFile(repo *GotRepository, filename string, data []byte) error {
	path := filepath.Join(repo.GotDir, filename)
	_, err := os.Stat(path)

	// The error is different from `file doesn't exist` then return.
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	// The file either doesn't exist or user want to write in any case on it.
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0644)
	if err == nil {
		defer file.Close()
		if _, err := file.Write(data); err != nil {
			return err
		}
		return nil
	}
	return err
}

// Set the repo configuration after setup. Future usage.
// func (gr *GotRepository) SetConfig(config map[string]interface{}) {
// 	gr.GotConfig = config.(GotConfig)
// }

func FindRecursivelyFolder(path string, folder string, until int) (string, error) {
	if until == 0 {
		return "", errors.New("recursive search exhausted")
	}
	if pathExist(filepath.Join(path, folder), true) {
		return path, nil
	}
	return FindRecursivelyFolder(filepath.Dir(path), folder, until-1)
}

// This must guarantee that folders exists otherwise it must create them from scratch
//
//   - Also, this is the place where the index is refreshed from disk. It doesn't exist yet, it
//     it must be created. If exists already, just refresh.
//
//   - To prove the repo existance, we can check the index file has data inside.
//     Find the repo if exist in the path or create new one.
func FindOrCreateRepo(path string) (*GotRepository, error) {
	if !pathExist(path, true) {

		return nil, ErrorPathDoesNotExist
	}
	fmt.Println(path)
	// Default folder location in case not .got folder found.
	gotDir := filepath.Join(path, gotRootRepositoryDir)
	treeDir := path
	dirPath, err := FindRecursivelyFolder(path, gotRootRepositoryDir, 4)
	// No error found, then dirPath is the rootDir where .got lives.
	if err == nil {
		treeDir = dirPath
		gotDir = filepath.Join(dirPath, gotRootRepositoryDir)
	}
	// Possible repo representation. No folder created yet.
	repo := &GotRepository{
		GotTree:   treeDir,
		GotDir:    gotDir,
		GotConfig: GotConfig{},
		Index:     NewIndex(),
	}
	//The folder path/.got exist and it has a version file that will determine this is already created repo.
	// The existance of index doesn't guarantee that the others important folders are created.
	if pathExist(filepath.Join(repo.GotDir, "index"), false) && pathExist(filepath.Join(repo.GotDir, gotRepositoryDirObjects), false) {
		content, err := os.ReadFile(filepath.Join(repo.GotDir, "index"))
		if err != nil {
			panic(err)
		}
		if len(content) > 0 {
			repo.Index.DeserializeIndex(content)
			return repo, nil
		} else {
			goto IndexFromScratch
		}
		// Otherwise returns default repo with index nil.
	} else {
		//The path/.got doesn't exist. Let's create it.
		if dir := gotDir; !pathExist(dir, true) {
			os.Mkdir(dir, fs.ModePerm|0755)
		}
		//Create the version's file and write in it.
		if err := CreateOrUpdateRepoFile(repo, "version", []byte(fmt.Sprintf("version: %s", version))); err != nil {
			panic(err)
		}

		//Create the version's file and write in it.
		if err := CreateOrUpdateRepoFile(repo, "HEAD", []byte(fmt.Sprintf("ref: refs/heads/main"))); err != nil {
			panic(err)
		}

		//Create path/refs
		if dir := filepath.Join(gotDir, gotRepositoryDirRefs); !pathExist(dir, true) {
			os.Mkdir(dir, fs.ModePerm|0755)
		}
		//Create path/objects
		if dir := filepath.Join(gotDir, gotRepositoryDirObjects); !pathExist(dir, true) {
			os.Mkdir(dir, fs.ModePerm|0755)
		}
		//Create path/refs/heads
		if dir := filepath.Join(gotDir, gotRepositoryDirRefs, gotRepositoryDirRefsHeads); !pathExist(dir, true) {
			os.Mkdir(dir, fs.ModePerm|0755)
		}
		//Create path/refs/tags
		if dir := filepath.Join(gotDir, gotRepositoryDirRefs, gotRepositoryDirRefsTags); !pathExist(dir, true) {
			os.Mkdir(dir, fs.ModePerm|0755)
		}
		goto IndexFromScratch
	}
	//Index doesn't exist at all.
IndexFromScratch:
	indexData := repo.Index.SerializeIndex()
	//Create the version's file and write in it.
	if err := CreateOrUpdateRepoFile(repo, "index", indexData); err != nil {
		panic(err)
	}

	return repo, nil
}

// Util function to determine whether the file/dir exists.
func pathExist(path string, mustBeDir bool) bool {
	fi, err := os.Stat(path)
	if errors.Is(err, os.ErrNotExist) {
		return false
	}
	if err != nil {
		panic(err)
	}
	if mustBeDir {
		return !errors.Is(err, os.ErrNotExist) && fi.IsDir()
	}
	return true
}



// List recursively the files in the worktree.
func listWorkTree(rootDir string) []fs.DirEntry {
	entries := make([]fs.DirEntry, 0)
	dirs, err := os.ReadDir(rootDir)
	// Implementation to discard .got folder.
	dirs = slices.DeleteFunc(dirs, func(e fs.DirEntry) bool {
		return e.Name() == ".got" && e.IsDir()
	})
	if err != nil {
		panic(err)
	}

	for _, dir := range dirs {
		if dir.IsDir() {
			entries = append(entries, listWorkTree(filepath.Join(rootDir, dir.Name()))...)
		} else {
			entries = append(entries, dir)
		}
	}
	return entries
}

func Status(repo *GotRepository) {
	worktree := listWorkTree(repo.GotTree)

	trackedFiles := make([]string, 0)
	untrackedFiles := make([]string, 0)

	for _, node := range worktree {
		idx := slices.IndexFunc(repo.Index.Entries, func(entry IndexEntry) bool {
			return node.Name() == entry.PathName
		})
		if idx >= 0 {
			trackedFiles = append(trackedFiles, repo.Index.Entries[idx].PathName)
		} else {
			untrackedFiles = append(untrackedFiles, node.Name())
		}
	}
	if len(trackedFiles) == 0 {
		fmt.Println("Nothing tracked yet.")
	}
	if len(trackedFiles) == 0 && len(untrackedFiles) == 0 {
		fmt.Println("Nothing in the worktree.")
	}
	fmt.Println(untrackedFiles)
	// refData, err := os.ReadFile(filepath.Join(repo.GotDir, "HEAD"))
	// if err != nil {
	// 	panic(err)
	// }
	// ref := ReferenceFromHEAD(repo, refData)
	// if ref.Invalid {
	// 	fmt.Println("HEAD reference is invalid. Whether there no commit or the file is incorrect.")
	// 	//nothing to compare with.
	// } else {
	// 	var headCommit string
	// 	if ref.IsDirect {
	// 		headCommit = ref.Reference
	// 	}
	// 	commit := ReadCommit(repo, headCommit)
	// 	fmt.Println(commit)
	// }

	// Given the tracked files,
	// 1. Compare; head treee
	// for _, tracked := range trackedFiles{
	// headCommit := ReadObject()
	// slices.IndexFunc(repo.Index.Cache, func(c CacheEntry) bool {
	// 	if tracked == c.PathName{

	// 	}
	// })
	// }

	//23) blob from tree vs blob from cache vs actual blob.
	// 1. List all files
	// 2. It'll create 2 groups;1) All tracked vs all Non-tracked.
	// 2.1. The tracked files may have 3 versions;
	//  - The blob already persisted in DB from HEAD ref.
	//  - The cached blob from previous git add where cached is different from persisted.
	//  - The user file changed recently (because, cached != file)
	// Once git added. Cache file needs to be updated.
}
