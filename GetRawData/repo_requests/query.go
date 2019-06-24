package repo_requests

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sort"
	"strings"
)

type Owner struct {
	Login string `json:"login"`
	Id    int    `json:"id"`
}

type Repo struct {
	Id       int     `json:"id"`
	Name     string  `json:"name"`
	FullName string  `json:"full_name"`
	Owner    Owner   `json:"owner"`
	CloneUrl string  `json:"clone_url"`
	Score    float64 `json:"score"`
	Language string  `json:"language"`
}

func (r *Repo) String() string {
	return fmt.Sprintf("Id: %v\n\tName:%v\n\tScore:%v\n\tLanguage:%v\n\n", r.Id, r.FullName, r.Score, r.Language)
}

type ByRepoId []Repo

func (brip ByRepoId) Len() int           { return len(brip) }
func (brip ByRepoId) Swap(i, j int)      { brip[i], brip[j] = brip[j], brip[i] }
func (brip ByRepoId) Less(i, j int) bool { return brip[i].Id < brip[j].Id }

type GitHubRepoResponse struct {
	Total int    `json:"total_count"`
	Items []Repo `json:"items"`
}

func (g *GitHubRepoResponse) FilterOnlyAbapRepos() {
	filter := make(map[int]struct{}, len(g.Items))

	sort.Sort(ByRepoId(g.Items))

	var repos []Repo
	tmp := repos[:0]

	// First, filter for abap repos
	for _, repo := range g.Items {
		if strings.ToLower(repo.Language) == "abap" {
			filter[repo.Id] = struct{}{}
		}

		//If the repo Id exists in the filter, add it because it's abap
		if _, ok := filter[repo.Id]; ok {
			tmp = append(tmp, repo)
		}
	}

	g.Items = tmp

}

func getRepoUrls(keyword string, repoSizeKb int, target *GitHubRepoResponse) *GitHubRepoResponse {
	client := &http.Client{}

	req, err := http.NewRequest("GET", "https://api.github.com/search/repositories", nil)

	if err != nil {
		log.Fatal(err)
	}

	req.Header.Set("Accept", "application/vnd.github.mercy-preview+json")

	q := req.URL.Query()

	params := fmt.Sprintf("%v+language:ABAP+size:>=%d+is:public", keyword, repoSizeKb)

	q.Add("q", params)

	req.URL.RawQuery = q.Encode()

	fmt.Printf("Querying for repos: %v\n", req.URL)
	resp, respErr := client.Do(req)

	if respErr != nil {
		log.Fatal(respErr)
	}

	defer resp.Body.Close()

	json.NewDecoder(resp.Body).Decode(&target)

	return target

}

func uniqueAbapRepos(r []Repo) []Repo {
	encountered := map[int]bool{}
	var uniqueRepos []Repo

	for _, repo := range r {
		if encountered[repo.Id] == true {
			// Don't add if already found
		} else {
			uniqueRepos = append(uniqueRepos, repo)
			encountered[repo.Id] = true
		}
	}

	return uniqueRepos
}

func GetAbapRepos(kws []string, gtkb int) []Repo {
	/*
		Args:
			kws		[]string		Provide a slice of keywords to search by for repositories
			gtkb	int				Provide the minimum size of repositories to search for in kb
	*/

	//Example keywords to search by
	//kws := []string{"*", "abap", "sap", "transaction"}

	var totalAbapRepos []Repo

	for _, keyword := range kws {
		target := new(GitHubRepoResponse)

		getRepoUrls(keyword, gtkb, target)
		target.FilterOnlyAbapRepos()

		totalAbapRepos = append(totalAbapRepos, target.Items...)

	}

	totalAbapRepos = uniqueAbapRepos(totalAbapRepos)

	lastIndex := len(totalAbapRepos) - 1

	fmt.Printf("Found %v repos\n", len(totalAbapRepos))
	for _, idx := range []int{0, lastIndex} {
		fmt.Printf("Top Repo: \n%v", totalAbapRepos[idx].String())
	}

	return totalAbapRepos
}
