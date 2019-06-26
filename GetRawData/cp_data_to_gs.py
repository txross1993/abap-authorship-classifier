#!/usr/bin/python3

###########################################################################
# Copy and update manifest.json to manifest_gs.json to reflect gs locations
###########################################################################

import json, sys

dest_prefix = ""

def load_manifest(input_json):
    with open(input_json, "rb") as f:
        data = f.read()

    return json.loads(data.decode('utf-8'))

def replacePrefix(json_data):
    global dest_prefix
    for repo in json_data["AuthorProjects"]:
        parts = repo["FileRef"].split('/')
        base_file_name = parts[-1]
        repo["FileRef"] = dest_prefix+("/{}".format(base_file_name))

    return json_data

def write_manifest_gs(replaced_json_data):
    with open(sys.argv[3], "w") as f:
        f.write(json.dumps(replaced_json_data))

def main():

    if len(sys.argv) < 4:
        print("Usage:\n")
        print("\tUse this utility to edit the manifest.json generated by GetRawData/get_data.go FileRefs to a new destination, retaining the base file names\n")
        print("\t\tNOTE: this utility expects to operate on local filesystem. \n")
        print("\tcp_data_to_gs.py <destination_file_prefix> <input_manifest_filepath> <output_manifest_filepath>\n")
        print("\tdestination_file_prefix: \"gs://mybucket/mydestinationfolder\" for example \n")
        print("\tinput_manifest_filepath: \"/path/to/my/manifest\"\n")
        print("\tinput_manifest_filepath: \"/path/to/my/output_manifest\"\n")
        sys.exit(1)

    print("Args: {}\t{}\t{}".format(sys.argv[1], sys.argv[2], sys.argv[3]))

    global dest_prefix
    dest_prefix = sys.argv[1]
    data = load_manifest(sys.argv[2])
    data_replaced = replacePrefix(data)
    write_manifest_gs(data_replaced)

if __name__=="__main__":
    main()