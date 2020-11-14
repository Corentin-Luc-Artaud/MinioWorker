# MinioWorker

MinioWorker is a simple program which allow you to trigger automated actions when an object is created in a MinIO bucket

# Configuration
````json
{
  "endpoint": "localhost:9000",     //the address of the MinIO server
  "accessKeyID": "minioadmin",
  "secretAccessKey": "minioadmin",
  "useSSL": false,                  // not implemented yet
  "maxConcurentWorker": 5,
  "maxConcurentUploadFiles": 10,
  "pipelineDirectory": "pipelines", //relative to the working directory
  "resultDestination": "compute"    //directory to put the produced files
}
````

All the files in the pipeline directory should be a pipeline definition

## Scripts

a script is a programs that expect to read on its standard input at least two lines  
- the directory it can use
- the file to process

it should write on its standard output the relative path (relative to the context directory) of all the files the process produces 

## Pipelines
```json
{
  "name": "hash",
  "bucketSource": "",
  "destinationPrefix": "hashes",
  "scripts": [
    {"script":"hash/hash.go", "runner":"go", "options": "run"}
  ],
  "include": "",
  "exclude": "\\.md5$"
}
```
This pipeline will compute the md5 hash for all the files that the name do not end with `.md5` in all the buckets. The produced files given the configuration above will be under the folder `/compute/hashes` in the corresponding bucket.

When a pipeline is triggered a temporary directory is created as the context directory for the instance.  

All the scripts should be under the pipelinedirectory.