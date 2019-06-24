package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
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
var LABELED_DATA_DIR = ""
var WRITES_CHAN = make(chan []byte, 255)
var WAIT_GROUP sync.WaitGroup
var encounteredProjectIds = make(map[int]bool)
var encounteredAuthorIds = make(map[int]bool)

type ManifestProjectLabel struct {
	AuthorId  int    `json:AuthorId`
	Author    string `json:"Author"`
	Project   string `json:"Project"`
	ProjectId int    `json:"ProjectId"`
	FileRef   string `json:"FileRef"`
}

type ManifestJson struct {
	TotalFiles    int                    `json:"TotalFiles"`
	TotalAuthors  int                    `json:"TotalAuthors"`
	TotalProjects int                    `json:"TotalProjects"`
	Projects      []ManifestProjectLabel `json:"AuthorProjects"`
}

func (m *ManifestJson) AddFile() {
	m.TotalFiles += 1
}

func (m *ManifestJson) AddAuthor(authorId int) {
	if encounteredAuthorIds[authorId] == true {
		// Already encountered, don't increment TotalAuthors
	} else {
		m.TotalAuthors += 1
		encounteredAuthorIds[authorId] = true
	}

}

func (m *ManifestJson) AddProject(projectId int) {
	if encounteredProjectIds[projectId] == true {
		// Already encountered, don't increment TotalAuthors
	} else {
		m.TotalProjects += 1
		encounteredProjectIds[projectId] = true
	}
}

func copyFile(sourceFile string, destinationFile string) {
	input, err := ioutil.ReadFile(sourceFile)
	if err != nil {
		log.Printf("\t\tError reading source abap file: %v\n", err)
		return
	}

	err = ioutil.WriteFile(destinationFile, input, 0755)
	if err != nil {
		log.Printf("\t\tError creating abap file in destination dir %v\n", destinationFile)
		log.Print(err)
		return
	}
	fmt.Printf("\tCopying abap file to data directory successful! %v", destinationFile)
}

func getRawData(repoDir string, label *ManifestProjectLabel) error {
	// Copy any abap files to the raw data directory
	err := filepath.Walk(repoDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Printf("prevent panic by handling failure accessing a path %q: %v\n", path, err)
			return err
		}

		if strings.HasSuffix(path, ".abap") {
			sourceFile, _ := filepath.Abs(path)
			log.Printf("\tFound abap data! %v\n", sourceFile)
			destinationFile, err := filepath.Abs(fmt.Sprintf("%v/%v", LABELED_DATA_DIR, filepath.Base(path)))

			if err != nil {
				log.Fatalf("Unable to find absolute path of destination file: %v,\n\tTrace: %v", destinationFile, err)
			}

			absPathDestinationFile, _ := filepath.Abs(destinationFile)
			if _, err := os.Stat(absPathDestinationFile); os.IsNotExist(err) {
				// If the destination file does not exist, copy it to the destination dir
				copyFile(sourceFile, destinationFile)
				if err != nil {
					return err
				}
			} else {
				log.Printf("Skipping copy of existing abap file %v", destinationFile)
			}

			// Annotate manifest
			/*
				AuthorId  int    `json:AuthorId`
				Author    string `json:"Author"`
				Project   string `json:"Project"`
				ProjectId int    `json:"ProjectId"`
				FileRef   string `json:"FileRef"`
			*/

			label.FileRef = destinationFile
			b, err := json.Marshal(label)
			if err != nil {
				log.Fatalf("ERROR - marshalling project data label failed: %v, %v", label, err)
			}

			log.Printf("\t\tWriting to manifest channel: %v\n", label)
			writeToManifestChan(b)

		}
		return nil
	})

	return err
}

func writeToManifestChan(data []byte) {
	// Write label to manifest file
	var b bytes.Buffer
	b.Write(data)
	b.Write([]byte("\n"))
	WRITES_CHAN <- b.Bytes()
}

func cloneRepo(url string, destination string) error {
	_, err := git.PlainClone(destination, false, &git.CloneOptions{
		URL: url,
	})

	log.Printf("Cloned repository %v\n", url)

	if err != nil {
		return err
	}

	return nil
}

func cloneAndCopy(repo_url string, dest string, label *ManifestProjectLabel) {
	defer WAIT_GROUP.Done()
	if _, err := os.Stat(dest); os.IsNotExist(err) {
		// If the repo directory does not exist, clone it
		err := cloneRepo(repo_url, dest)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		log.Println("Skipping existing repository: %v\n", repo_url)
	}
	err := getRawData(dest, label)
	if err != nil {
		log.Fatal("ERROR getting raw abap data files: %v", err)
	}
}

func CloneAllAndCopy(in []rq.Repo, destDir string) {
	/* ManifestProjectLabel
	AuthorId  int    `json:AuthorId`
	Author    string `json:"Author"`
	Project   string `json:"Project"`
	ProjectId int    `json:"ProjectId"`
	FileRef   string `json:"FileRef"`
	*/
	for _, repo := range in {
		label := &ManifestProjectLabel{
			AuthorId:  repo.Owner.Id,
			Author:    repo.Owner.Login,
			Project:   repo.Name,
			ProjectId: repo.Id,
			FileRef:   "",
		}
		dest := destDir + "/" + label.Project
		log.Printf("Initiating goroutine for repo %v", label.Project)
		go cloneAndCopy(repo.CloneUrl, dest, label)
	}
	return
}

func main() {
	// Load env vars
	dotenv.Load()
	REPO_DIR, _ := filepath.Abs(os.Getenv("REPO_DIR"))
	MANIFEST_FILE, _ := filepath.Abs(os.Getenv("MANIFEST_FILE"))
	REPO_SIZE_KB := os.Getenv("REPO_SIZE_KB")
	LABELED_DATA_DIR = os.Getenv("LABELED_DATA_DIR")
	if value, exists := os.LookupEnv("REPO_SIZE_KB"); exists {
		log.Printf("Found env var REPO_SIZE_KB: %v", value)
	}

	log.Printf("Loaded environment variables:\n\tREPO_DIR: %v,\n\tMANIFEST_FILE: %v,\n\tREPO_SIZE_KB: %v,\n\tLABELED_DATA_DIR: %v", REPO_DIR, MANIFEST_FILE, REPO_SIZE_KB, LABELED_DATA_DIR)

	REPO_SIZE_KB_INT, err := strconv.Atoi(REPO_SIZE_KB)
	if err != nil {
		REPO_SIZE_KB_INT = 3000
	}

	log.Printf("Converted REPO_SIZE_KB_INT: %v", REPO_SIZE_KB_INT)

	// Search for repositories
	keyword_search := []string{"*", "abap", "sap", "program"}
	allRepos := rq.GetAbapRepos(keyword_search, REPO_SIZE_KB_INT)

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

	defer f.Close()
	defer close(WRITES_CHAN)

	MANIFEST_JSON := ManifestJson{
		TotalAuthors: 0,
		TotalFiles:   0,
		Projects:     []ManifestProjectLabel{},
	}

	go func() {
		for {
			data := <-WRITES_CHAN
			label := ManifestProjectLabel{}
			log.Printf("Retrieved data from channel %v\n", data)
			err := json.Unmarshal(data, &label)
			if err != nil {
				log.Fatalf("ERROR - Unable to unmarshal json from channel %v", err)
			}

			// Increment Manifest Json Counters
			MANIFEST_JSON.AddAuthor(label.AuthorId)
			MANIFEST_JSON.AddProject(label.ProjectId)
			MANIFEST_JSON.AddFile()
			MANIFEST_JSON.Projects = append(MANIFEST_JSON.Projects, label)

		}

	}()

	WAIT_GROUP.Wait()

	//Write Json To File
	jsonData, err := json.MarshalIndent(MANIFEST_JSON, "", " ")
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Writing to manifest file %v", MANIFEST_FILE)
	ioutil.WriteFile(MANIFEST_FILE, jsonData, 0755)
}
