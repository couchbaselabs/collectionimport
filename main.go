package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"sync"

	"github.com/couchbase/gocb"
)

var waitGroup sync.WaitGroup
var data chan string
var bucket *gocb.Bucket

func main() {
	fmt.Println("Starting the import process...")

	flagInputFile := flag.String("input-file", "", "file with path which contains documents")
	flagWorkerCount := flag.Int("workers", 20, "concurrent workers for importing data")
	flagCollectionName := flag.String("collection", "", "mongodb collection name")
	flagCouchbaseHost := flag.String("couchbase-host", "", "couchbase cluster host")
	flagCouchbaseBucket := flag.String("couchbase-bucket", "", "couchbase bucket name")
	flagCouchbaseBucketPassword := flag.String("couchbase-bucket-password", "", "couchbase bucket password")
	flag.Parse()

	cluster, _ := gocb.Connect("couchbase://" + *flagCouchbaseHost)
	bucket, _ = cluster.OpenBucket(*flagCouchbaseBucket, *flagCouchbaseBucketPassword)

	file, _ := os.Open(*flagInputFile)
	defer file.Close()

	data = make(chan string)

	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)

	for i := 0; i < *flagWorkerCount; i++ {
		waitGroup.Add(1)
		go worker(*flagCollectionName)
	}

	for scanner.Scan() {
		data <- scanner.Text()
	}

	close(data)

	waitGroup.Wait()

	fmt.Println("The import has completed!")
}

func worker(collection string) {
	defer waitGroup.Done()
	for {
		document, ok := <-data
		if !ok {
			break
		}
		cbimport(document, collection)
	}
}

func cbimport(document string, collection string) {
	var mapDocument map[string]interface{}
	json.Unmarshal([]byte(document), &mapDocument)
	mapDocument["_type"] = collection
	compressObjectIds(mapDocument)
	bucket.Insert(mapDocument["_id"].(string), mapDocument, 0)
}

func compressObjectIds(mapDocument map[string]interface{}) string {
	var objectIdValue string
	for key, value := range mapDocument {
		switch value.(type) {
		case string:
			if key == "$oid" && len(mapDocument) == 1 {
				return value.(string)
			}
		case map[string]interface{}:
			objectIdValue = compressObjectIds(value.(map[string]interface{}))
			if objectIdValue != "" {
				mapDocument[key] = objectIdValue
			}
		case []interface{}:
			for index, element := range value.([]interface{}) {
				objectIdValue = compressObjectIds(element.(map[string]interface{}))
				if objectIdValue != "" {
					value.([]interface{})[index] = objectIdValue
				}
			}
		}
	}
	return ""
}
