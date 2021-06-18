# Single Middleware Frontend
A middleware that can transform, modify and forward requests based on pre-defined configuration. It allows
you to modify your client request such as add, modify, or delete key-value, transform your request into JSON or XML 
and send the response back to the client. Execute Request configuration as a 
serial or parallel request


## Prerequisite
1. Create a project folder inside configures folder. Take a look inside 
   **configures.testing/test-1**.
2. Inside a project folder  create file **configure.json,  base.json, 
   serial/parallel.json**. **configure.json** can be renamed to anything.
3. Create src/.env file **(see src/.env.example file)**.
4. Specify the path to the configures directory in .env **(see src/.env.example file)** .

<br> <br>
See **configures.testing directory structure** for examples.


## File Base.json
```
{
  "project_max_circular" : 10,
  "circular_response" : {
    "status_code" : 508,
    "adds" : {
      "body" : {
        "message" : "Circular request",
        "error" : "circular request detected"
      }
    }
  }
}
```

### project_max_circular
Specify maximal number of circular request. This can be happen if 
configure-1 refer to configure-2 but configure-2 refer to configure-1.
### circular_response
Specify response to be returned if circular request has reached 
project_max_circular.s

## File Router.json
In order to make request to this middleware, take a look inside **configures.testing/router.json**.
 - **path** : Client to middleware endpoint.,
 - **project_directory** : Project directory path relative to **router.json**,
 - **type** : Make a serial or parallel request. The default value is serial 
   if value is empty. Middleware will read related files (serial/parallel.json)
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

## File Serial.json
Execute request sequentially. This json structure accept list of configures.
for example:
``` 
{
  "configures": [
    {
      "file_name": "test-1_configure-0.json",
      "alias": "$configure_test-1",
      "failure_response": {
        "status_code": 500,
        "transform": "ToJson",
        "adds": {
          "body": {
            "error_message": "Request Logic error or there is something error"
          }
        }
      },
      "c_logics": []
    }
  ]
}
```
### file_name
Specify configure file name
### alias
Specify alias for configure file name. Must contain prefix $configure
### failure_response
Specify response if all logic fail
### c_logics
Accept list of json logic structure, based on https://jsonlogic.com/operations.html for example :
```
    {
          "rule": {
            "==": [
              "$configure_test-3_2--$request--$query[movie_id]",
              550
            ]
          },
          data : null,
          "response":null,
          "next_success : null,
          "next_failure": null,
          "failure_response : null
        }
```
#### rule
Json logic operator/rule.

#### data
Json logic data.

#### response
Return response if logic is true.

#### failure_response
return response if logic is false.

#### next_success
specify configure alias to be processed if logic is true. Response must be null.

#### next_failure
specify configure alias to be processed if logic is false. Response must be
false.

#### Note
<li>If cLogics is empty, return last configure
executed response. </li>
<li> next_success or next_failure in file parallel.json value only accept 
serial.json or parallel.json, cannot refer to other configure</li>

## File Parallel.json
Execute request simultaneously. See example **configures.testing/test-6.3** .
### failure_response
response to be returned when all logic is fail.
### configures
List of configuration file.
#### file_name
configure file name.
#### alias
configure alias. Must contain prefix $configure.

#### c_logics
check logic after all parallel request processed.

## File Configure.json

### Request to destination endpoint
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

If a request/response don't have any wrapper for the body and you want to convert it to xml, 
it automatically wrap your request with **<doc>** such that **<doc>**(body) 
**</doc>** ,  because this middleware use package [clbanning/mxj](https://github.com/clbanning/mxj), <br>

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
    },
    c_logics : [],
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

#### 4.4 c_logics in request
Check logic before sending request. If logic is false, you can specify 
failure_response to return response or next_failure execute next config.
In serial request, if failure_response and response not specified, middleware 
will return 
serial failure_response for current processed configure that has been 
specified in serial.json. In Parallel request, middleware will return 
failure_response that has been specified in parallel.json.
Middleware will not send any request. If logic is true, you can specify response
to return response or next_success to execute next config.
If response and next_success not specified, middleware will send request.
```
    "c_logics": [
      {
        "rule": {
          "and": [
            {
              "==": [
                "$configure_test-4_0--$request--$query[movie_id]",
                550
              ]
            },
            {
              "==": [
                "$configure_test-4_0--$request--$query[movie_id2]",
                384018
              ]
            }
          ]
        },
        "data" : null,
        "response" :  null,
        "next_success" : "$configure_test-4_1",
        "next_failure" : null,
        "failure_response" : null
      }
    ]
```

#### 5 Modify response from each request.
You can modify response from each request but only for **body** and **header**, the rules are the same like request modification.
<br>
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
1. If you use serial route, you can only pick the value from previous configure, you can't pick the value from configure2.json
the middleware execute the request sequentially. For example, you **can** pick a value for **configure1.json** from
**configure0.json** request or response,  you **can't** pick a value from **configure3.json**. This is because
**configure0.json** is the first index in configures directory, following by **configure1.json**, and last **configure3.json**.

**Remember that the order of configure-n.json file in configures directory is really important for serial route**

2. If you  use parallel route, you can't pick the value between each configure because the middleware
execute the request simultaneously. For example, you **can't** pick a value for **configure1.json** from **configure0.json** request or response.


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
















