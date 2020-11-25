# A server to do addition, modification, and deletion request based on json file.

### To run this project, simply go build main.go --configures=./configures
 --configures is an argument to read directory which contain configure0.json to configure-n.json and response.json  relative to main.go.
 The ordering in configure-n.json is important for server to process each configures serially if you use serial endpoint.

## Adds, Default, Modify Request/Response
Server will do the addition, modification, and deletion sequentially. To add a key-value pair, you can specify it
in configure-n.json in the adds object. 
###1. Addition
To add a key-value to a non-nested object, you can write usual key-value pair to the adds object in configure-n.json.
To add a key-value pair to a nested object, you can separate it by dot(.). For example if you want to add a key "name" and value "nicholas", you can
write user.name=nicholas.
###2. Modification
To modify a key-value pair, the key must be exist in request.
###3. Deletion
To delete a key-value pair, you can add a key to deletion array in configure-n.json

## Get request value between each configures
If you want to pick request between each configure, you need to specify which configure to be picked.
For example if you want to pick request from configure0.json
