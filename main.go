package main

import (
    "fmt"
    "os"
    "./elasticsearchsetup"
)

const url = "http://localhost:9200"
const sniff = false
const index = "elastic-test"
const mapping = "mapping.json"
const field = "mo_notes"

func main() {
    client := elasticsearchsetup.SetupTestClient(url, sniff)
    elasticsearchsetup.SetupTestClientAndCreateIndexAndAddDocs(client, index, mapping, field, true)
}

func check(err error) {
    if err != nil {
        fmt.Println(err)
        os.Exit(1)
    }
}
