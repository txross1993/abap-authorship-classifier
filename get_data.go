package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	dotenv "github.com/joho/godotenv"
	git "gopkg.in/src-d/go-git.v4"
)

var MANIFEST_FILE = ""
var WRITES_CHAN = make(chan []byte, 255)
var WAIT_GROUP sync.WaitGroup

type Owner struct {
	User string `json:"login"`
}

type Repo struct {
	Id          int    `json:"id"`
	Owner       Owner  `json:"owner"`
	CloneUrl    string `json:"clone_url"`
	ProjectName string `json:"name"`
}

type GitHubRepoResponse struct {
	Total int    `json:"total_count"`
	Items []Repo `json:"items"`
}

type SortCondition int

const (
	best_match   SortCondition = 0
	comments     SortCondition = 1
	interactions SortCondition = 2
)

func (c SortCondition) GetParams() string {
	if c < best_match || c > interactions {
		return ""
	}

	conditions := [...]string{
		"",
		"comments",
		"interactions",
	}

	return conditions[c]
}

type ManifestLabel struct {
	author   string `json:"author"`
	project  string `json:"project"`
	file_ref string `json:"file_ref"`
}

func get_repo_urls(repo_size_kb int, target GitHubRepoResponse, sc SortCondition) GitHubRepoResponse {
	client := &http.Client{}

	req, err := http.NewRequest("GET", "https://api.github.com/search/repositories", nil)

	if err != nil {
		log.Fatal(err)
	}

	req.Header.Add("Accept", "application/vnd.github.mercy-preview+json")

	q := req.URL.Query()

	params := []string{
		"language:abap",
		fmt.Sprintf("size:>=%d", repo_size_kb),
		"is:public",
	}

	for _, qry_param := range params {
		q.Add("q", qry_param)
	}

	if sc != 0 {
		q.Add("sort", sc.GetParams())
	}

	req.URL.RawQuery = q.Encode()

	fmt.Printf("Querying for repos: %v", req.URL)
	resp, resp_err := client.Do(req)

	if resp_err != nil {
		log.Fatal(resp_err)
	}

	defer resp.Body.Close()

	json.NewDecoder(resp.Body).Decode(&target)

	return target

}

func GetRepos(in ...GitHubRepoResponse) []Repo {
	encountered := map[int]bool{}
	var all_repos []Repo
	for _, ghr := range in { // for all github responses
		for _, repo := range ghr.Items {
			if encountered[repo.Id] == true {
				// Don't add if already found
			} else {
				all_repos = append(all_repos, repo)
				encountered[repo.Id] = true
			}
		}
	}

	return all_repos
}

func copy_file(sourceFile string, destinationFile string) {
	input, err := ioutil.ReadFile(sourceFile)
	if err != nil {
		fmt.Println(err)
		return
	}

	err = ioutil.WriteFile(destinationFile, input, 0755)
	if err != nil {
		fmt.Println("Error creating", destinationFile)
		fmt.Println(err)
		return
	}
}

func get_raw_data(repo_dir string, label *ManifestLabel) error {
	// Copy any abap files to the raw data directory
	err := filepath.Walk(repo_dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Printf("prevent panic by handling failure accessing a path %q: %v\n", path, err)
			return err
		}

		if strings.HasSuffix(path, ".abap") {
			sourceFile, _ := filepath.Abs(path)
			destinationFile, _ := filepath.Abs(fmt.Sprintf("%v/%v", os.Getenv("LABELED_DATA_DIR"), filepath.Base(path)))
			copy_file(sourceFile, destinationFile)

			// Annotate manifest
			label.file_ref = destinationFile
			data, _ := json.Marshal(label)
			write_to_manifest_chan(data)

		}
		return nil
	})

	return err
}

func write_to_manifest_chan(data []byte) {
	// Write label to manifest file
	var b bytes.Buffer
	b.Write([]byte(data))
	b.Write([]byte("\n"))
	WRITES_CHAN <- b.Bytes()
}

func clone_repo(url string, destination string) error {
	_, err := git.PlainClone(destination, false, &git.CloneOptions{
		URL: url,
	})

	fmt.Printf("Cloned repository %v", url)

	if err != nil {
		return err
	}

	return nil
}

func clone_and_cp(repo_url string, dest string, label *ManifestLabel) error {
	defer WAIT_GROUP.Done()
	err := clone_repo(repo_url, dest)
	if err != nil {
		return err
	}
	err = get_raw_data(dest, label)
	return err
}

func CloneAllAndCopy(in []Repo, dest_dir string) {
	for _, repo := range in {
		label := &ManifestLabel{
			author:   repo.Owner.User,
			project:  repo.ProjectName,
			file_ref: "",
		}
		dest := dest_dir + "/" + repo.ProjectName
		fmt.Printf("Initiating goroutine for repo %v", repo.ProjectName)
		go clone_and_cp(repo.CloneUrl, dest, label)
	}
	return
}

func main() {
	// Load env vars
	dotenv.Load()
	REPO_DIR := os.Getenv("REPO_DIR")
	MANIFEST_FILE, _ := filepath.Abs(os.Getenv("MANIFEST_FILE"))
	REPO_SIZE_KB, _ := strconv.Atoi(os.Getenv("REPO_SIZE_KB"))
	/*
		Query results only return 30 repos at a time,
		so query 3 times with different sort arguments to get 90 repositories
	*/

	// Best Match
	sort_by_best_matches_json := GitHubRepoResponse{}
	sbbm := get_repo_urls(REPO_SIZE_KB, sort_by_best_matches_json, best_match)

	// Num Interactions
	sort_by_interactions := GitHubRepoResponse{}
	sbi := get_repo_urls(REPO_SIZE_KB, sort_by_interactions, interactions)

	// Comments
	sort_by_comments_json := GitHubRepoResponse{}
	sbc := get_repo_urls(REPO_SIZE_KB, sort_by_comments_json, comments)

	all_repos := GetRepos(sbbm, sbi, sbc)
	WAIT_GROUP.Add(len(all_repos))
	fmt.Println("Got all repositories . . . Beginning cloning and labeling")

	/*
		Clone all repositories into data dir
		Traverse repositories, find .abap files, and move them to LABELED_DATA_DIR
		Annotate manifest with file name, github author, and project
	*/

	// Clone

	CloneAllAndCopy(all_repos, REPO_DIR)

	// Write channel to manifest file
	f, _ := os.Create(MANIFEST_FILE)
	bw := bufio.NewWriter(f)

	defer f.Close()
	defer close(WRITES_CHAN)

	go func() {
		for {
			data := <-WRITES_CHAN
			_, err := bw.Write(data)

			if err != nil {
				fmt.Println(err)
			}
			if bw.Available() < 32 {
				bw.Flush()
			}
		}

	}()

	WAIT_GROUP.Wait()

}
