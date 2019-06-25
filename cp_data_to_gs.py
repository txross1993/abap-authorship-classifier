#!/usr/bin/python3

###########################################################################
# Copy and update manifest.json to manifest_gs.json to reflect gs locations
###########################################################################

import json

dest_prefix = "gs://emergingtech/abap_classifier/data"

def load_manifest(input_json):
    with open(input_json, "rb") as f:
        data = f.read()

    return json.loads(data)

def replacePrefix(json_data):
    for repo in json_data["AuthorProjects"]:
        parts = repo["FileRef"].split('/')
        base_file_name = parts[-1]
        repo["FileRef"] = dest_prefix+("/{}".format(base_file_name))

    return json_data

def write_manifest_gs(replaced_json_data):
    with open("manifest_gs.json", "w") as f:
        f.write(json.dumps(replaced_json_data))

def main():
    data = load_manifest("manifest.json")
    data_replaced = replacePrefix(data)
    write_manifest_gs(data_replaced)

if __name__=="__main__":
    main()