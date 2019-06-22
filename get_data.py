import requests, json, os, shutil

from concurent.futures import ThreadPoolExecutor
from dotenv import load_dotenv, find_dotenv
from git import Repo

# Load environment vars
load_dotenv(find_dotenv())
print("Loading environment variables")
REPO_DIR = os.getenv("REPO_DIR")
print("Directory for found repositories: %s" % REPO_DIR)
LABELED_DATA_DIR = os.getenv("LABELED_DATA_DIR")
print("Directory for labeled manifest and abap files: %s" % LABELED_DATA_DIR)
MANIFEST_FILE = os.getenv("MANIFEST_FILE")
print("File path for manifest: %s" % MANIFEST_FILE)

# Check Github for public repositories of language ABAP that is at least 100 KB in size
REPO_SIZE_GT = os.getenv("REPO_SIZE_KB") # Size in kilobytes of repository
URL = "https://api.github.com/search/repositories"
HEADERS = {"Accept": "application/vnd.github.mercy-preview+json"}
QUERY_PARAMS = ["language:abap",
                "is:public",
                "size:>={}".format(REPO_SIZE_GT)] # Github api query search terms
PARAMS = {"q": QUERY_PARAMS}

def clone_repo(url, repo_dir):
    _ = Repo.clone_from(url, repo_dir)
    return

def clone_repos() -> list:
    repo_url_list = []
    repo_dir_list = []

    r = requests.get(URL, headers=HEADERS, params=PARAMS)
    print("Running query: {}".format(r.url))
    data = r.json()
    print("Total number of repositories: {}".format(data['total_count']))

    # Clone all repos into director named <github_author>_<github_project>
    for repo in data["items"]:
        # dir name
        repo_dir = "{}/{}_{}".format(REPO_DIR, repo['owner']['login'], repo['name'])

        # append to list
        print("Adding repository to list %s" % repo_dir)
        repo_url_list.append(repo['clone_url'])
        repo_dir_list.append(repo_dir)

        # clone
    with ThreadPoolExecutor(max_workers=4) as executor:
        _ = executor.map(clone_repo, repo_url_list, repo_dir_list)

    return repo_dir_list

def get_author_project(repo):
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

def extract_files_from_repos(manifest, repo_dirs) -> dict:
    """
    Args
    repo_dirs       list    Provide the list of git repositories

    Returns
    manifest  dict   json for manifest
    """

    manifest["total_count"] += len(repo_dirs)

    for repo in repo_dirs:

        author, project = get_author_project(repo)

        for root, _, fnames in os.walk(os.path.abspath(repo)):
            # Find .abap files in repo
            for fname in fnames:
                if fname.split(".")[-1] == "abap":
                    print("Found abap source code: %s" % fname)
                    # cp file to raw data directory
                    orig_fname = os.path.join(os.path.abspath(root), fname)
                    print("Original File Path %s" % orig_fname)
                    new_fname = os.path.join(os.path.abspath(os.getenv(LABELED_DATA_DIR)), fname)
                    print("Target file path %s" % new_fname)

                    print("Copying file {} to data directory {}".format(fname, LABELED_DATA_DIR))
                    shutil.copy(orig_fname, new_fname)

                    # Write line for datapoint 'author,project,file_path,\n'
                    manifest["repos"].append({"author": author, "project": project, "file_ref": new_fname})
    
    return manifest

def main():
    repo_list = clone_repos()
    print("Got repo list, cloning and extracting")
    manifest = generate_manifest()
    manifest = extract_files_from_repos(manifest, repo_list[0:5])

    with open(os.path.abspath(MANIFEST_FILE), "w", encoding="utf-8") as f:
        json.dump(manifest, f, ensure_ascii=False, indent=2)

if __name__ == "__main__":
    main()