#!/usr/bin/python3

###########################################################################
# Copy and update manifest.json to manifest_gs.json to reflect gs locations
###########################################################################

import json, sys

dest_prefix = "gs://emergingtech/abap_classifier/data"

def load_manifest(input_json):
    with open(input_json, "rb") as f:
        data = f.read()

    return json.loads(string(data))

def replacePrefix(json_data):
    for repo in json_data["AuthorProjects"]:
        parts = repo["FileRef"].split('/')
        base_file_name = parts[-1]
        repo["FileRef"] = dest_prefix+("/{}".format(base_file_name))

    return json_data

def write_manifest_gs(replaced_json_data):
    with open(sys.argv[2], "w") as f:
        f.write(json.dumps(replaced_json_data))

def main():
    data = load_manifest(sys.argv[1])
    data_replaced = replacePrefix(data)
    write_manifest_gs(data_replaced)

if __name__=="__main__":
    main()