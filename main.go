package main

import (
    "fmt"
    "os"
    "flag"
    "log"
    "./elasticsearchsetup"
)

const url_c = "http://localhost:9200"
const sniff_c = false
const index_c = "elastic-test"
const mapping_c = "mapping.json"
const field_c = "mo_notes"

func main() {
  var (
    url     = flag.String("url", url_c, "Elasticsearch host.")
    sniff     = flag.Bool("sniff", sniff_c, "Sniff elasticsearch nodes")
    index   = flag.String("index", index_c, "Index name.")
    mapping   = flag.String("mapping", mapping_c, "Mapping file.")
    field   = flag.String("field", field_c, "Field name for text search tests.")
  )
  flag.Parse()

  if *url == "" {
    log.Fatal("missing url")
  }
  if *index == "" {
    log.Fatal("missing index name")
  }
  if *mapping == "" {
    log.Fatal("missing mapping file definition")
  }
  if *field == "" {
    log.Fatal("missing text search field definition")
  }

  client := elasticsearchsetup.SetupTestClient(*url, *sniff)
  elasticsearchsetup.SetupTestClientAndCreateIndexAndAddDocs(client, *index, *mapping, *field, true)
}

func check(err error) {
    if err != nil {
        fmt.Println(err)
        os.Exit(1)
    }
}
