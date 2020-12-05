# Golang Body Parser
A server that can transform, modify and forward requests based on pre-defined configuration. It allows
you to modify your client request such as add, modify, or delete key-value, transform your request into JSON or XML 
and send the response back to the client.

##Prerequisite
1. You need to have a directory that contains **configure-n.json and response.json**. **configure-n.json** can be renamed to configure0.json, or configure1.json as long as your file have 'configure', **NOTE : double dash (--)** for in configure files name are prohibited since this symbol will be used. 
2. Create src/.env file **(see src/.env.example file)**.
3. Specify the path to the configures directory in .env **(see src/.env.example file)** .
<br> <br>
See **configures.example directory structure** for examples.


## Server Route/Endpoint
In order to make request to this server, you have to specify the **path** value in configure-n.json (your configure file.json). By default, the server routes are :  
- **/serial** for each GET, POST, PUT, DELETE
- **/parallel** for each GET, POST, PUT, DELETE

If you specify the path in configure file, then the path will be :
- **/serial/your-path** for each GET, POST, PUT, DELETE
- **/parallel/your-path**  for each GET, POST, PUT, DELETE
##### Path parameter
If you want to use parameter in your path, you can use **/:value**. For example, if you want the server route is **/user/1**, then you can specify
**/user/:id** in configure path.

##Request
 **For each configure file in your configures directory, it represents a request**.
#### 1. Specify the target URL for request
In order to forward your request, the server need to know the destination url and destination path.
1. Specify **destinationUrl** in configure file
2. Specify **destinationPath** in configure file

The final url is **destinationUrl/destinationPath**. This is the url where the server will do a client request.

#### 2. Specify the request method
You must specify your request method for key **methodUsed** in configure file. The following values are
available : **(UPPERCASE)** :
- GET
- POST
- PUT
- DELETE

### 3. Transform your request
You can transform  your request to JSON or XML. By default, the server will transform your request to JSON format.
You can specify your request format for key **transform** in configure file. The following values are available :
- **ToJson** to transform your request to JSON format.
- **ToXml** to transform your request to XML format

If a request/response don't have any wrapper for the body and you want to convert it to xml, because this server use package [clbanning/mxj](https://github.com/clbanning/mxj),
it automatically wrap your request with **<doc>** such that **<doc>**(body) **</doc>**. <br>

For example if you send an empty request to the server, and your configuration for request like below and you want to transform it to XML : 
``` 
  "request": {
    "destinationPath": "/",
    "destinationUrl": "http://localhost:3001",
    "methodUsed": "PUT",
    "transform": "ToJson",
    "logBeforeModify" : "",
    "logAfterModify" : "",
    "adds": {
      "header": {
      },
      "body": {
        "name" : "nicholas",
        "last_name": "anthony"


      },
      "query": {
      }
    },
    "modifies": {
      "header": {

      },
      "body": {

      },
      "query": {
      }
    },
    "deletes": {
      "header": [
      ],
      "body": [
      ],
      "query": [
      ]
    }
  },
```

The server will add **name** and **last_name** to the body, but because the request don't have any wrapper element, your request will be :
```
<doc>
    <last_name>anthony</last_name>
    <name>nicholas</name>
</doc>
```

If you want to pick the value from this configure, for example **name**, you have to mention the **doc** like **$body[doc][name]** .


### 4. Request modification
To perform addition, modification, deletion, **methodUsed** value in configure must exist in key array **methods**. 
For example, if **methods** contain **POST** and your **methodUsed** value is **PUT**, request modification will not be performed.
After you specify **methods** value in configure, server will do **addition**, **modification**, **deletion** sequentially. 
Request modification can be performed for **header, body, and query**.

#### 4.1 adds
You can add key-value to your request, by specify **key** and **value**. You can also add key-value to a nested
object by using ".(dot)". If the object is not exist, the server will create it for you. In this example, we will add key **id** , and also a 
**name** property to a nested object **user** to a request body.
```
"adds": {
      "header": {
      },
      "body": {
          "id" : "123",
          "user.name" : "nicholas"
      },
      "query": {}
    },
```

Note : **Header and query key-value cannot be a nested object**

####4.2 modify
You can modify your request by specify key-value pair in **modify** . In order to modify a certain key from your request, key must already **existed*** in your request.
The following examples shows that we want to modify **id** and modify **name** from nested object **user**.
```
    "modifies": {
      "header": {
      },
      "body": {
        "id" : 456,
        "user.name" : "anthony"
      },
      "query": {}
    },
```

Note : **key-value for request header and query cannot be a nested object**
#### 4.3 deletes
To delete keys from your request, you can specify it in **deletes**. The following example shows that we want to delete **name** property from nested object **user**, and we want to delete **id**.
```
    "deletes": {
      "header": [],
      "body": [
           "id",
           "user.name"
      ],
      "query": []
    }
```

Note : You can also delete an entire object, for example, replace **user.name** with **user**, you will delete object user from your request.

#### 4.4 Modify response from each request.
You can modify response from each request but only for **body** and **header**, the rules are the same like request modification.
<br>
Note : **key-value for response header cannot be a nested object**

### 5. Response Modification to client
To do a response modification that will be sent to the client, you may want to take a look at response.json file in configures directory. This file
will be used by the server to do response modification. You can only do response modification to header and body.
The rules are the same like request modification, but in response.json we have **configureBased** key.

####5.1 Return response from specific request
Because configure file represent each request, we can point which configure response we want to return. For example, if
we want to return response from request configure0.json, and we want to add additional key **id** to response header and **name** to nested object **user** to response body,
we can write like this

```
{
  "configureBased": "$configure0.json",
  "response": {
    "transform": "ToJson",
    "adds": {
      "header": {
        "id" : "123"
      },
      "body": {
        "user.name" : "nicholas"
      }
    },
    "modifies": {
      "header" : [],
      "body": {}
    },
    "deletes": {
      "header" : [],
      "body": []
    }
  }
}
```
In this example, we take the response from request configure0.json as a base response by specifying the value **configureBased**, and we add additional key-value. If you don't want
to use a certain request  as a base response, you can leave **configureBased** empty.

Note : **key-value for response header cannot be a nested object**

### 6. Get a value between each configure
To pick a value between each configure, you must consider from **which configure** you want to pick the value, whether your value located in :
- **request header, body, query, or path parameter**
- **response body or header**

For example, if our configure1.json want to pick **name** from nested object **user** which located in response body in configure0.json, 
we can write :
```
$configure0.json--$request--$body[user][name]
```
If you want to pick a value from **array**, for example if the request from configure0.json has **cars** 
in the nested object **user** and the value is **["toyota","honda","hyundai"]**, then we can pick toyota like the following code :
```
$configure0.json--$request--$body[user][cars][0]
```
Notice that each section is separated by **double dash (--)**. That's why you can't use double dash for configure file name.
#### Consideration to pick a value between each configure
1. If you use serial route, you can only pick the value from previous configure, because
a certain value from configure0.json, however, you can't pick the value from configure2.json
the server execute the request sequentially. For example, you **can** pick a value for **configure1.json** from
**configure0.json** request or response, however, you **can't** pick a value from **configure3.json**. This is because
**configure0.json** is the first index in configures directory, following by **configure1.json**, and last **configure3.json**.

**Remember that the order of configure-n.json file in configures directory is really important for serial route**

2. If you  use parallel route, you can't pick the value between each configure because the server
execute the request simultaneously. For example, you **can't** pick a value for **configure1.json** from **configure0.json** request or response.

3. In file **response.json**, you are safe to pick value from configure-n.json request or response. This is because the response is the last step to be sent to the client, so
the server will wait until every request execution is finished.

### 7. Logging
You can log **header, body, query, path parameter** from each configure file by specifying
value for keys : 
-  **logBeforeModify** : This will log a value before change/modify request/response. 
-  **logAfterModify** : This will log a value after change/modify request/response.

Following values are available:
- **$body**
- **$header**
- **$query**
- **$path**

<br> 
Or if you want to log specific value, you can specify it like this: $body[user][name].

Example :
``` 
"logBeforeModify" : "$body", // this will log the whole body before doing any change.
"logAfterModify" : "$body[user][name]" // this will log name from object user after doing change.
```
















