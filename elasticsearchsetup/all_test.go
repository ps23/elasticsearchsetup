package elasticsearchsetup

import (
  "testing"
	"log"
  "os"

	elastic "github.com/olivere/elastic"
)

const url = "http://localhost:9200"
const sniff = false
const index = "elastic-test"
const mapping = "../mapping.json"
const field = "mo_notes"

func TestMain(m *testing.M) {
	// call flag.Parse() here if TestMain uses flags
  log.Printf("Setting up tests")
  client := SetupTestClient(url, sniff)
  SetupTestClientAndCreateIndexAndAddDocs(client, index, mapping, field, true)
  log.Printf("Running tests")
	os.Exit(m.Run())
}

func TestDummy(t *testing.T) {
  type args struct {
    client *elastic.Client
    index string
		field string
    numSlices int
  }

	// Create an Elasticsearch client
	client := SetupTestClient(url, sniff)

  tests := []struct {
    name string
    args args
    want bool
  }{
    {"test0", args{client, "elastic-test", "mo_notes", 4}, true},
  }

  for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) {
      if got := Dummy(tt.args.client, tt.args.index, tt.args.field, tt.args.numSlices); got != tt.want {
        t.Errorf("Dummy() = %v, want %v", got, tt.want)
      }
    })
  }
}
