import requests, json, os, shutil
from dotenv import load_dotenv, find_dotenv
from git import Repo

# Load environment vars
load_dotenv(find_dotenv())
REPO_DIR = os.getenv("REPO_DIR")
LABELED_DATA_DIR = os.getenv("LABELED_DATA_DIR")
MANIFEST_FILE = os.getenv("MANIFEST_FILE")

# Check Github for public repositories of language ABAP that is at least 100 KB in size
REPO_SIZE_GT = 100 # Size in kilobytes of repository
URL = "https://api.github.com/search/repositories"
HEADERS = {"Accept": "application/vnd.github.mercy-preview+json"}
QUERY_PARAMS = ["language:abap",
                "is:public",
                "size:>={}".format(REPO_SIZE_GT)] # Github api query search terms
PARAMS = {"q": QUERY_PARAMS}

def clone_repos() -> list:
    repo_list = []

    r = requests.get(URL, headers=HEADERS, params=PARAMS)
    data = r.json()
    print("Total number of repositories: {}".format(data['total_count']))

    # Clone all repos into director named <github_author>_<github_project>
    for repo in data["items"]:
        # dir name
        repo_dir = "{}/{}_{}".format(REPO_DIR,repo['owner']['login'], repo['name'])

        # append to list
        repo_list.append(repo_dir)

        # clone   
        _ = Repo.clone_from(repo['clone_url'], repo_dir)

    return repo_list

def get_author_project(repo str):
    parts = repo.split("/")
    author_project = parts[-1].split("_")
    author = author_project[0]
    project = author_project[1]
    return author, project

def generate_manifest() -> dict:
    manifest = {
        "total_count": int,
        "repos": []
    }

    return manifest

def extract_files_from_repos(repo_dirs list) -> dict:
    """
    Args
    repo_dirs       list    Provide the list of git repositories

    Returns 
    manifest  dict   json for manifest
    """
    manifest = generate_manifest()

    manifest["total_count"] = len(repo_dirs)

    pattern = "^{}/.+_[^/].+$".format(REPO_DIR)
    repo_dir_pattern = re.compile(pattern)

    for repo in repo_dirs:

        author, project = get_author_project(repo)

        for root, dirs, fnames in os.walk(os.path.abspath(repo)):
            # Find .abap files in repo
            for fname in fnames:
                if fname.split(".")[-1] == "abap":
                    # cp file to raw data directory
                    orig_fname = os.path.join( os.path.abspath(root), fname )
                    new_fname = os.path.join( os.path.abspath(os.getenv( LABELED_DATA_DIR )), fname)
                    shutil.copy(orig_fname, new_fname)

                    # Write line for datapoint 'author,project,file_path,\n'
                    manifest["repos"].append({"author": author,
                                            "project": project,
                                            "file_ref": new_fname})
    
    return manifest



def main():
    repo_list = clone_repos()
    manifest = extract_files_from_repos(repo_list)

    with open(os.path.abspath(MANIFEST_FILE), "w", encoding="utf-8") as f:
        json.dump(manifest, f, ensure_ascii=False, indent=2)

if __name__=="__main__":
    main()