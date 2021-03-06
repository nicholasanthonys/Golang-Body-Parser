# Golang Body Parser
A middleware that can transform, modify and forward requests based on pre-defined configuration. It allows
you to modify your client request such as add, modify, or delete key-value, transform your request into JSON or XML 
and send the response back to the client.


## Prerequisite
1. Create a project folder inside configures folder. Take a look inside **configures.example**.
2. Inside a project folder  create file **configure-n.json and response.json**. **configure-n.json** can be renamed to configure0.json, or configure1.json as long as your file have 'configure', **NOTE : double dash (--)** for in configure files name are prohibited since this symbol will be used. 
3. Create src/.env file **(see src/.env.example file)**.
4. Specify the path to the configures directory in .env **(see src/.env.example file)** .

<br> <br>
See **configures.example directory structure** for examples.


## Middleware Route/Endpoint
### Client to middleware endpoint
In order to make request to this middleware, take a look inside **configures.example/router.json**.
 - **path** : Client to middleware endpoint.,
 - **project_directory** : Project directory path relative to **router.json**,
 - **type** : Make a serial or parallel request. The default value is serial if value is empty.
 - **method** : Client to middleware request method. <br>Available values are 
   - GET
   - POST
   - PUT
   - DELETE
   
#### Path parameter
If you want to use path parameter from Client to middleware, you can use **/:key**. 

For example, if middleware make a request with POST method to middleware at http://localhost:8000/smsotp/generate/25 for project smsotp, and then  make a request
from middleware to destination as a serial request :
```
[
  {
    "path" : "/smsotp/generate/:smsId",
    "project_directory" : "smsotp",
    "type" : "serial",
    "method" : "POST"
  }
]
```


### Middleware to destination endpoint
## Request
 **For each configure file in your project directory, it represents a request**.
#### 1. Specify the target URL for request
In order to forward your request, the middleware need to know the destination url and destination path.
1. Specify **destinationUrl** in configure file
2. Specify **destinationPath** in configure file
3. The final endpoint is : **destinationURL + destinationPath**

#### 2. Specify the request method
You must specify your request method for key **method** in configure file. The following values are
available : **(UPPERCASE)** :
- GET
- POST
- PUT
- DELETE

### 3. Transform your request
You can transform  your request to JSON or XML. By default, the middleware will transform your request to JSON format.
You can specify your request format for key **transform** in configure file. The following values are available :
- **ToJson** to transform your request to JSON format.
- **ToXml** to transform your request to XML format

If a request/response don't have any wrapper for the body and you want to convert it to xml, because this middleware use package [clbanning/mxj](https://github.com/clbanning/mxj),
it automatically wrap your request with **<doc>** such that **<doc>**(body) **</doc>**. <br>

For example if you send an empty request to the middleware, and your configuration for request like below and you want to transform it to XML : 
``` 
  "request": {
    "destinationPath": "/",
    "destinationUrl": "http://localhost:3001",
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

The middleware will add **name** and **last_name** to the body, but because the request don't have any wrapper element, your request will be :
```
<doc>
    <last_name>anthony</last_name>
    <name>nicholas</name>
</doc>
```

If you want to pick the value from this configure, for example **name**, you have to mention the **doc** like **$body[doc][name]** .


### 4. Request modification
Middleware will do **addition**, **modification**, **deletion** sequentially. Request modification can be performed for **header, body, and query**.

#### 4.1 adds
You can add key-value to your request, by specify **key** and **value**. You can also add key-value to a nested
object by using ".(dot)", or directly add a nested object. If the object is not exist, the middleware will create it for you. In this example, we will add key **id** ,
**name** and **favourite_cars** property to a nested object **user** to a request body using dot syntax, add nested object address.
```
"adds": {
      "header": {
      },
      "body": {
          "id" : "123",
          "user.name" : "nicholas",
          "user.favorite_cars": [
              "honda",
              "fiat",
              "toyota",
              "ferrari"
          ],
          "address" : {
              "city" : "bandung",
              "province" : "west java"
          }
      },
      "query": {}
    },
```

[comment]: <> (Note : **Header and query key-value cannot be a nested object**)

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
will be used by the middleware to do response modification. You can only do response modification to header and body.
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
1. If you use serial route, you can only pick the value from previous configure, you can't pick the value from configure2.json
the middleware execute the request sequentially. For example, you **can** pick a value for **configure1.json** from
**configure0.json** request or response,  you **can't** pick a value from **configure3.json**. This is because
**configure0.json** is the first index in configures directory, following by **configure1.json**, and last **configure3.json**.

**Remember that the order of configure-n.json file in configures directory is really important for serial route**

2. If you  use parallel route, you can't pick the value between each configure because the middleware
execute the request simultaneously. For example, you **can't** pick a value for **configure1.json** from **configure0.json** request or response.

3. In file **response.json**, you are safe to pick value from configure-n.json request or response. This is because the response is the last step to be sent to the client, so
the middleware will wait until every request execution is finished.

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
















