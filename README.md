# MongoDB Collection to Couchbase Importer

Import MongoDB collection data that had been exported via the MongoDB CLI tool, [mongoexport](https://docs.mongodb.com/manual/reference/program/mongoexport/).  This project is not production ready and should be used with care.

## Building the Project

This project was written with the Go programming language.  To build the project, Golang must be installed on your machine.  After adding this project to your **$GOPATH**, execute the following from within the project directory:

```
go get -d -v
go build
```

You should be left with a **collectionimport** binary.

## Exporting Collection Data from MongoDB

Per the MongoDB documentation, collections can be exported via the following command:

```
mongoexport --db DATABASE_NAME_HERE --collection COLLECTION_NAME_HERE --out OUTPUT_FILE.json
```

The output JSON file will consist of one document per line and might look similar to the following:

```
{"_id":{"$oid":"58a23bbc7627c561d77d7a81"},"name":"Comcast","tenants":[{"$oid":"58a23bc27627c561d77d7a82"},{"$oid":"58a244c89dfa1f66a092776e"}]}
{"_id":{"$oid":"58a244bc9dfa1f66a092776d"},"name":"Google"}
```

It is normal that every referred document in a MongoDB collection be wrapped in an `$oid` property.  As is, this JSON file can be imported into Couchbase with the official import tools.  However, the data is a bit messy.

## Importing Collection Data to Couchbase

The goal here is import the MongoDB collection in a very clean fashion.  To run the `collectionimport` tool, execute the following:

```
./collectionimport \
    --input-file FILE_NAME.json \
    --collection COLLECTION_NAME \
    --couchbase-host localhost \
    --couchbase-bucket default \
    --workers 20
```

If using the previous collection example, one of the documents that exist in Couchbase would look something like this:

```
{
    "_id": "58a23bbc7627c561d77d7a81",
    "_type": "services",
    "name": "Comcast",
    "tenants": [
        "58a23bc27627c561d77d7a82",
        "58a244c89dfa1f66a092776e"
    ]
}
```

In the above example, the `_type` property is populated via the collection name that was passed in at runtime.  Each of the `$oid` properties were compressed to be only the id values.  The document key will match the `_id` value.

## How it Works

When the application starts, each of the runtime flags are consumed and used towards configuring the application.  This includes where the JSON file is located, what collection it came from, and various Couchbase connection information.

After the JSON file has opened, X number of workers are created for processing the lines read from the JSON file.  Remember, each line is a separate document.  Workers are necessary because requests against Couchbase are blocking.  Waiting for Couchbase to respond for each import could take time on large amounts of data which is why multiple imports should happen concurrently.

During the import, each document receives a `_type` and each `$oid` property within the document, no matter how deep, gets compressed.  Compressing will pull out the id value and remove the `$oid` wrapper.

## How to get Help

If you need help using this project, feel free to try one of the following:

Reach me on Twitter - [@nraboy](https://www.twitter.com)

Visit the forums - [https://forums.couchbase.com](https://forums.couchbase.com)

## Resources

Couchbase - [https://www.couchbase.com](https://www.couchbase.com)

The Polyglot Developer - [https://www.thepolyglotdeveloper.com](https://www.thepolyglotdeveloper.com)