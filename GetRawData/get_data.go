package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	rq "github.com/txross1993/abap-authorship-classifier/GetRawData/repo_requests"

	dotenv "github.com/joho/godotenv"
	git "gopkg.in/src-d/go-git.v4"
)

var MANIFEST_FILE = ""
var WRITES_CHAN = make(chan []byte, 255)
var WAIT_GROUP sync.WaitGroup

type ManifestLabel struct {
	author  string `json:"author"`
	project string `json:"project"`
	fileRef string `json:"file_ref"`
}

func copyFile(sourceFile string, destinationFile string) {
	input, err := ioutil.ReadFile(sourceFile)
	if err != nil {
		fmt.Printf("\t\tError reading source abap file: %v\n", err)
		return
	}

	err = ioutil.WriteFile(destinationFile, input, 0755)
	if err != nil {
		fmt.Printf("\t\tError creating abap file in destination dir %v\n", destinationFile)
		fmt.Println(err)
		return
	}
	fmt.Printf("\tCopying abap file to data directory successful! %v", destinationFile)
}

func getRawData(repoDir string, label *ManifestLabel) error {
	// Copy any abap files to the raw data directory
	err := filepath.Walk(repoDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Printf("prevent panic by handling failure accessing a path %q: %v\n", path, err)
			return err
		}

		if strings.HasSuffix(path, ".abap") {
			sourceFile, _ := filepath.Abs(path)
			fmt.Printf("\tFound abap data! %v\n", sourceFile)
			destinationFile, _ := filepath.Abs(fmt.Sprintf("%v/%v", os.Getenv("LABELED_DATA_DIR"), filepath.Base(path)))

			absPathDestinationFile, _ := filepath.Abs(destinationFile)
			if _, err := os.Stat(absPathDestinationFile); os.IsNotExist(err) {
				// If the destination file does not exist, copy it to the destination dir
				copyFile(sourceFile, destinationFile)
				if err != nil {
					return err
				}
			} else {
				fmt.Println("Skipping copy of existing abap file %v \n", destinationFile)
			}

			// Annotate manifest
			label.fileRef = destinationFile
			data, _ := json.Marshal(label)
			fmt.Printf("\t\tWriting to manifest channel: %v\n", label)
			writeToManifestChan(data)

		}
		return nil
	})

	return err
}

func writeToManifestChan(data []byte) {
	// Write label to manifest file
	var b bytes.Buffer
	b.Write([]byte(data))
	b.Write([]byte("\n"))
	WRITES_CHAN <- b.Bytes()
}

func cloneRepo(url string, destination string) error {
	_, err := git.PlainClone(destination, false, &git.CloneOptions{
		URL: url,
	})

	fmt.Printf("Cloned repository %v\n", url)

	if err != nil {
		return err
	}

	return nil
}

func cloneAndCopy(repo_url string, dest string, label *ManifestLabel) error {
	defer WAIT_GROUP.Done()
	if _, err := os.Stat(dest); os.IsNotExist(err) {
		// If the repo directory does not exist, clone it
		err := cloneRepo(repo_url, dest)
		if err != nil {
			return err
		}
	} else {
		fmt.Println("Skipping existing repository: %v\n", repo_url)
	}
	err := getRawData(dest, label)
	return err
}

func CloneAllAndCopy(in []rq.Repo, destDir string) {
	for _, repo := range in {
		label := &ManifestLabel{
			author:  repo.Author(),
			project: repo.Name,
			fileRef: "",
		}
		dest := destDir + "/" + label.project
		fmt.Printf("Initiating goroutine for repo %v", label.project)
		go cloneAndCopy(repo.CloneUrl, dest, label)
	}
	return
}

func main() {
	// Load env vars
	dotenv.Load()
	REPO_DIR := os.Getenv("REPO_DIR")
	MANIFEST_FILE, _ := filepath.Abs(os.Getenv("MANIFEST_FILE"))
	REPO_SIZE_KB, _ := strconv.Atoi(os.Getenv("REPO_SIZE_KB"))

	// Search for repositories
	keyword_search := []string{"*", "abap", "sap", "program"}
	allRepos := rq.GetAbapRepos(keyword_search, REPO_SIZE_KB)

	WAIT_GROUP.Add(len(allRepos))
	fmt.Println("Got all repositories . . . Beginning cloning and labeling")

	/*
		Clone all repositories into data dir
		Traverse repositories, find .abap files, and move them to LABELED_DATA_DIR
		Annotate manifest with file name, github author, and project
	*/

	// Clone

	CloneAllAndCopy(allRepos, REPO_DIR)

	// Write channel to manifest file
	f, _ := os.Create(MANIFEST_FILE)
	bw := bufio.NewWriter(f)

	defer f.Close()
	defer close(WRITES_CHAN)

	go func() {
		for {
			data := <-WRITES_CHAN
			fmt.Printf("Retrieved data from channel %v\n", data)
			_, err := bw.Write(data)

			if err != nil {
				fmt.Println(err)
			}
			bw.Flush()
		}

	}()

	WAIT_GROUP.Wait()

}
