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

var BaseRepoConfig = GotConfig{
	Bare: false,
	User: UserConfig{
		Name:  "Arnulfo Telaentierra",
		Email: "dejemonosdevainas@email.com",
	},
	Branch: "master",
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
		if err := CreateOrUpdateRepoFile(repo, "HEAD", []byte("ref: refs/heads/main")); err != nil {
			panic(err)
		}
		//Create the config's file and write in it.
		if err := CreateOrUpdateRepoFile(repo, "config", []byte(BaseRepoConfig.toBytes())); err != nil {
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

type DirEntry struct {
	RelativePath string
}

// List recursively the files in the worktree.
func listWorkTree(rootDir string) []string {
	entries := make([]string, 0)
	dirs, err := os.ReadDir(rootDir)
	if err != nil {
		panic(err)
	}
	// Implementation to discard .got folder.
	dirs = slices.DeleteFunc(dirs, func(e fs.DirEntry) bool {
		return e.Name() == ".got" && e.IsDir()
	})
	for _, dir := range dirs {
		if dir.IsDir() {
			entries = append(entries, listWorkTree(filepath.Join(rootDir, dir.Name()))...)
		} else {
			entries = append(entries, filepath.Join(rootDir, dir.Name()))
		}
	}
	return entries
}

func relativize(repo *GotRepository, path string) string {
	rel, err := filepath.Rel(repo.GotTree, path)
	if err != nil {
		panic(err)
	}
	return rel
}

func relativizeMultiPaths(repo *GotRepository, paths []string) []string {
	rels := make([]string, 0)
	for _, path := range paths {
		rels = append(rels, relativize(repo, path))
	}
	return rels
}

// Candiate to pointer Got Reposu
func (repo *GotRepository) Status() {
	//read all the files in the worktree.
	worktree := listWorkTree(repo.GotTree)
	//Container for tracked files(already either in index or in DB)
	trackedFiles := make([]string, 0)
	// Container for untracked files.
	untrackedFiles := make([]string, 0)

	for _, node := range worktree {
		// - what it does: Find out whether the node/file is already in index.
		// - Here we don't know if the blob is index only or it has been persisted in DB already.
		if idxAtIndex := slices.IndexFunc(repo.Index.Entries, func(entry IndexEntry) bool {
			return relativize(repo, node) == entry.PathName
		}); idxAtIndex >= 0 {
			trackedFiles = append(trackedFiles, repo.Index.Entries[idxAtIndex].PathName)
		} else {
			untrackedFiles = append(untrackedFiles, node)
		}
	}

	var headCommit *Commit
	// Get the HEAD reference to compare with its tree.
	ref := repo.GetHEADReference()

	if ref.Invalid {
		fmt.Println("HEAD reference is invalid. Whether there no commit or the file is incorrect.")
		headCommit = nil
	} else {
		headCommit = ReadCommit(repo, ref.Reference)
	}
	// What it means: There is already a worktree and there are commit already on the HEAD.
	if headCommit != nil {
		fmt.Println("There already a tree. So there might be a blon from tree in the tree.")
		// Based on the HEAD commit, obtain the tree associated.
		rawTree, err := ReadObject(repo, TreeHeaderName, headCommit.Tree)
		if err != nil {
			panic(err)
		}
		var dummy TreeItem
		tree := dummy.Deserialize(rawTree)
		// What it does: recursively make all tree blob flatten into an array of strings.
		pathsInTree := tree.FlatItems()

		for _, trackFile := range trackedFiles {
			// - Find out if the tracked file is persisted on the tree of the HEAD.
			// - if is persisted compare its hash with latest file state hash of the user.
			//   -  if the hashes are different, files are different and user has modified the file.
			//     - if blob is in cache means that the user has already added the file to stage area.
			//     - if cache hash is different from latest file state hash of the user,then user has modified the file since the
			//        the last time the files has been added to stage area.
			if idxPersisted := slices.IndexFunc(pathsInTree, func(ti TreeItem) bool {
				return relativize(repo, ti.Path) == trackFile
			}); idxPersisted >= 0 {
				// Generate an in-memory object from latest user tracked file.
				hash, err := CreatePossibleObjectFromData(repo, Blob{Path: filepath.Join(repo.GotTree, trackFile)}, BlobHeaderName)
				if err != nil {
					panic(err)
				}
				// Validation #1: tree blob different from user blob. File changed.
				if hash != pathsInTree[idxPersisted].Hash {
					// Validation #2: stage area blob different from user blob. File has changed.
					if idxAtCache := slices.IndexFunc(repo.Index.Cache, func(cache CacheEntry) bool {
						return trackFile == cache.PathName
					}); idxAtCache >= 0 {
						if repo.Index.Cache[idxAtCache].Hash != hash {
							// File has changed from previous state of the stage area.
							fmt.Println("cache file different from user's recently changed file lines.")
						} else {
							// Cache and user file are the same. File recently added but not modified.
						}
					} else {
						//File no added after being modified
					}

				} else {
					// Tracked file is not modified.
				}
			} else {
				//File tracked for first time because is not in tree.
			}
		}
	} else {
		// THis is the first commit. Only matter validation of the cache vs user files.
	}
	fmt.Println(relativizeMultiPaths(repo, untrackedFiles), trackedFiles)
}

func (repo *GotRepository) GetConfiguration() GotConfig {
	content, err := os.ReadFile(filepath.Join(repo.GotDir, "config"))
	if err != nil {
		panic(err)
	}
	var config GotConfig
	if err = Unmarshal(content, &config); err != nil {
		panic(err)
	}
	return config
}
