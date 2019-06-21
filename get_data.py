import requests, json
from git import Repo

REPO_SIZE_GT = 100 # Size in kilobytes of repository
URL = "https://api.github.com/search/repositories"
HEADERS = {"Accept": "application/vnd.github.mercy-preview+json"}
QUERY_PARAMS = ["language:abap",
                "is:public",
                "size:>={}".format(REPO_SIZE_GT)] # Github api query search terms

PARAMS = {"q": QUERY_PARAMS}
DATA_DIR = "/home/theadora_ross/cnn-abap-features/data/raw"

def main():
    # curl -H "Accept: application/vnd.github.mercy-preview+json" https://api.github.com/search/repositories?q=language:abap&is:public&size:>=100

    
    
    r = requests.get(URL, headers=HEADERS, params=PARAMS)

    print(r.url)

    data = r.json()

    print("Total number of repositories: {}".format(data['total_count']))

    for repo in data["items"]:
        #print(repo['clone_url'])
        Repo.clone_from(repo['clone_url'], "{}/{}_{}".format(DATA_DIR,repo['owner']['login'], repo['name']))

if __name__=="__main__":
    main()