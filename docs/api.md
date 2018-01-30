# API Documentation

Documentation for the Flotilla API. 


## Task Endpoints

Endpoints for creating task definitions, executing task definitions, viewing run history, and retrieving run logs.

#### List Task Definitions

Lists all configured task definitions

* **URL**
  `/api/v1/task`
  
* **Method**
  `GET`
  
* **URL Params**
  
  **Optional**
  
  * `limit=[integer]`
  
      Limit results to this number. Default is 1024.
  
  * `offset=[integer]`
  
      Page offset. Default is 0.
      
  * `image=[string]`
  
      Limit results to those with the specified image. Strict match.
      
  * `group_name=[string]`
  
      Limit results to those with the specified group. Strict match.
   
  * `alias=[string]`
     
      Limit results to those with the specified alias. Substring match.
      
  * `tag=[string]`
  
      Limit results to those that are tagged with the specified tag. Strict match.
  
* **Success Response**

  * **Code:** 200 <br />
    **Content:**
    
    ```
    { 
      group_name : "testing",
      definitions: [
      {
        "alias": "new_alias"
        "command": "echo 'eat me'",
        "cpu": 0,
        "definition_id": "testing-05140d0f-2f9b-4f61-9778-349b73fbcc70",
        "group_name": "testing",
        "image": "library/ubuntu:latest",
        "memory": 1024
      }]
    }
    ```

#### Create Task Definition

Create a new task definition. These task definitions can then be listed, updated, executed, or deleted.

* **URL**
`/api/v1/task`

* **Method**
`POST`

* **URL Params**
* **Data Params**

  **Required**
  
  `image=[string]`
  
 Docker image name to use. `[?registry_uri]:[?registry_port]/[image_name]:[image_tag]`
  
 `command=[string]`
 
 Shell command to run inside the specified docker image. It can contain newlines.
 
 `group_name=[string]`
 
 Logical group this task definition belongs to. This could, for example, be an application namespace or a functional group of users.
 
 **Optional**
 
 ```
 env=[
   {name:[string], value:[string]}
 ]
 ```
 
 Environment variables to be set within every task executed with this definition.
 
 `memory=[int]`
 
 Memory, specified in megabytes, to allocate to tasks executed with this definition.
 
 ```
 tags=[
      [string]
 ]
 ```
 
 List of tags to use for this definition. Helpful for search and filtering.
 
 `alias=[string]`
 
 An alias to give this definition.
 
 
* **Success Response**

   * **Code:** 200 <br />
    **Content:**
    
    ```
    {
      "alias": "hello_world",
      "command": "echo 'hi'",
      "definition_id": "testing-5c042dbe-5895-465e-bb53-c53c07dcd5ea",
      "group_name": "testing",
      "image": "library/ubuntu:latest",
      "memory": 512
    }
    ```

#### Get Task Definition

Gets a specific task definition, by definition_id.

* **URL**
`/api/v1/task/[definition_id]`

* **Method**
`GET`

* **URL Params**
* **Data Params**
* **Success Response**

  * **Code:** 200 <br />
    **Content:**
    
    ```
    {
      "alias": "hello_world",
      "command": "echo 'hi'",
      "definition_id": "testing-5c042dbe-5895-465e-bb53-c53c07dcd5ea",
      "environment": [{
        "name": "GIT_SSH",
        "value": "/stitchfix/git/git-ssh.sh"
      }],
      "group_name": "testing",
      "image": "library/ubuntu:latest",
      "memory": 512,
      "tags": []
      "user": "nobody"
    }
    ```
    
#### Update Task Definition

Update a task definition. Currently, all fields that can be set at task definition creation time, with the exception of `group_name` can also be updated. This includes the `image`, `alias`, `command`, `tags`, `env`, and `memory`.

* **URL**
`/api/v1/task/[definition_id]`

* **Method**
`PUT`

* **URL Params**
* **Data Params**
* **Success Response**
  * **Code:** 200 <br />
    **Content:**
    
    ```
    {
      "alias": "hello_world-updated",
      "command": "echo 'hi'",
      "definition_id": "testing-5c042dbe-5895-465e-bb53-c53c07dcd5ea",
      "group_name": "testing",
      "image": "library/ubuntu:latest",
      "memory": 512,
      "tags": ["different","tags"]
      "user": "nobody"
    }
    ```

#### Delete Task Definition

Deregisters (and eventually deletes) a task definition. Definitions are deregistered, rather than being immediately deleted, since there may still be running tasks associated with the definition.

* **URL**
`/api/v1/task/[definition_id]`

* **Method**
`DELETE`

* **URL Params**
* **Data Params**
* **Success Response**
  * **Code:** 200 <br />
    **Content:**
    
    ```
    {
      "definition_id": "testing-5c042dbe-5895-465e-bb53-c53c07dcd5ea",
      "deleted": true
    }
    ```

#### Execute Task

Instantiates the specified task definition as a running task on a cluster.

* **URL**
`/api/v1/task/[definition_id]/execute`

* **Method**
`PUT`

* **URL Params**
* **Data Params**
  
  **Optional**
  
  * `cluster=[string]`
  
     Cluster name of cluster to launch task on. Default is `default`.
      


  * `env=[list]`
  
      ```
      env=[
        {'name':[string], 'value':[string]}
      ]
      ```
      
      Specify runtime environment overrides.
      
  
* **Success Response**
  * **Code:** 200 <br />
    **Content:**
    
    ```
    {
      "alias": "hello_world-updated",
      "command": "echo 'hi'",
      "definition_id": "testing-5c042dbe-5895-465e-bb53-c53c07dcd5ea",
      "group_name": "testing",
      "image": "library/ubuntu:latest",
      "memory": 512,
      "tags": ["different","tags"]
    }
    ```


#### Get All Task Executions

Lists all executions of the specified task definition.

* **URL**
`/api/v1/task/[definition_id]/history`

* **Method**
`GET`

* **URL Params**

  **Optional**
    
  * `limit=[integer]`
  
      Limit results to this number. Default is 1024.
  
  * `offset=[integer]`
  
      Page offset. Default is 0.
      
  * `status=[string]`
  
      Restrict results to those that have the specified status. One of `QUEUED`, `PENDING`, `RUNNING`, or `STOPPED`. Case sensitive.
       
  * `cluster_name=[string]`

      Return only those results that executed on the specified cluster.
  
* **Data Params**
* **Success Response**
  * **Code:** 200 <br />
    **Content:**
    
    ```
    {
      "definition_id": "testing-2a36919b-3ddd-49dd-bbe3-52301f440132",
      "history": [
        {
          "alias": "a1",
          "cluster": "flotilla-thingamabob",
          "command": " echo $RUNTIME_PARAM",
          "definition_id": "testing-2a36919b-3ddd-49dd-bbe3-52301f440132",
          "exit_code": 0,
          "finished_at": "2016-08-10 13:21:12",
          "group_name": "testing",
          "image": "library/ubuntu:latest",
          "memory": 512,
          "run_id": "0074a511-e635-4755-8661-eab7f6f30a67",
          "started_at": "2016-08-10 13:21:09",
          "status": "STOPPED",
          "tags": [
            "applesauce"
          ]
        }
      ]
      "offset": 0
      "total": 2
    }
    ```
    
#### Get Task Execution

Returns the specified task execution by `run_id`.

* **URL**
`/api/v1/task/[definition_id]/history/[run_id]`

* **Method**
`GET`

* **URL Params**
* **Data Params**
* **Success Response**
  * **Code:** 200 <br />
    **Content:**
    
    ```
    {
      "status": "STOPPED",
      "exit_code": 0,
      "started_at": "2016-08-08 08:13:37-0700",
      "finished_at": "2016-08-08 08:13:37-0700",
      "run_id": "f82eda60-6155-46de-b758-927a2e18a97a",
      "definition_id": "testing-5c042dbe-5895-465e-bb53-c53c07dcd5ea",        
      "alias": "hello_world-updated",
      "command": "echo 'hi'",
      "group_name": "testing",
      "image": "library/ubuntu:latest",
      "memory": 512,
      "tags": ["different","tags"]
    }
    ```

#### Stop Task Execution

Stops a running task. No-op if the task is no longer running.

* **URL**
`/api/v1/task/[definition_id]/history/[run_id]`

* **Method**
`DELETE`

* **URL Params**
* **Data Params**
* **Success Response**
  * **Code:** 200 <br />
    **Content:**
    
    ```
    {
      "definition_id": "testing-5c042dbe-5895-465e-bb53-c53c07dcd5ea",
      "stopped": true
    }
    ```
    
#### Get Task Execution Logs

Returns latest logs for a specified task `run_id`

* **URL**
`/api/v1/[run_id]/logs`

* **Method**
`GET`

* **URL Params**

  **Optional**
  
  `last_seen=[string]`
  
  Last seen token (from a previous call to this endpoint). Allows only fetching logs that come after the specified `last_seen` token.
  
* **Data Params**
* **Success Response**
  * **Code:** 200 <br />
    **Content:**
    
    ```
    {
      "last_seen": "f/32797019482293329766487489112145098696112137148578922496",
      "log": "hi"
    }
    ```
