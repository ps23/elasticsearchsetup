package elasticsearchsetup

import (
  "context"
  "log"
  "fmt"
  "io/ioutil"
  elastic "github.com/olivere/elastic"
  recipes "github.com/ps23/elasticrecipes"
)

const pipeline string = "opennlp-pipeline"
//const testIndexName string = "elastic-test"

func SetupOpenNlpPipeline(client *elastic.Client, id string, field string) {
  fmt.Println("Adding ingest pipeline: " + pipeline)

  var err error

  // Read in nginx_json_template
  buf, err := ioutil.ReadFile(pipeline + ".json")
  if err != nil {
    log.Fatal(err)
  }

  putres, err := client.IngestPutPipeline("opennlp-pipeline").
  BodyString(string(buf)).
  Do(context.Background())

  if err != nil {
    log.Fatalf("expected no error; got: %v", err)
  }
  if putres == nil {
    log.Fatalf("expected response; got: %v", putres)
  }
  if !putres.Acknowledged {
    log.Fatalf("expected ingest pipeline to be ack'd; got: %v", putres.Acknowledged)
  }
}

func SetupTestClient(url string, sniff bool) *elastic.Client {
  if( ! recipes.CheckStatus(url, 10) ) { log.Fatal("Could not find any elastic instance on this url") }

  var err error

  client, err := elastic.NewClient(elastic.SetURL(url), elastic.SetSniff(sniff))
  if err != nil {
    log.Fatal(err)
  }

  return client
}

func SetupTestClientAndCreateIndex(client *elastic.Client, index string, mapping string, clean bool)  {
  log.Printf("func SetupTestClientAndCreateIndex")

  if clean {
    client.DeleteIndex(index).Do(context.TODO())
  }

  // Create index
  createIndex, err := recipes.SetMap(client, index, "mapping.json")
  if err != nil {
    log.Fatal(err)
  }
  if createIndex == false {
    log.Fatalf("expected result to be != nil; got: %v", createIndex)
  }
}

func SetupTestClientAndCreateIndexAndAddDocs(client *elastic.Client, index string, mapping string, field string, clean bool) {
  log.Printf("func SetupTestClientAndCreateIndexAndAddDocs")

  var err error

  SetupTestClientAndCreateIndex(client, index, mapping, clean)
  SetupOpenNlpPipeline(client, pipeline, field)

  // Add docs
  var docs []map[string]string
  docs = append(docs, map[string]string{field: "DESCRIBED OFFENDER APPROCHED SECURE SEMI DETATCHED HOUSE FROM FRONT IP SAW SERCUTIY LIGHT COME ON AT FRONT OG HOUSE AND HEARD A NOISE FROM DOWNSTIARS , IPWENT TO SEE WHAT WAS CASUING THE NOISE AND WENT DOWNSTAIRS , HE SAW THE FRONT DOOR OPEN AJAR AND OFFENDR STOOD OUTSIDE IT IN OPEN PORCH , OFFENDER THEN MADE OFF ON FOOT , NO ENTRY GAINED , NO VISIBLE MARKS AT POINT OF ENTRY.",},)
  docs = append(docs, map[string]string{field: "BETWEEN MATERIAL TIMES OFFENDER(S) N/K ENTERED PREMISES VIA FIRE EXIT STAIRS, BY FORCING WINDOW, ENTERED THROUGH BATHROOM, FORCED UP CARPET/FLOORBOARDS TO GAIN ACCESS TO POST OFFICE BELOW. NO PROPERTY TAKEN FROM PREMISES.",},)
  docs = append(docs, map[string]string{field: "B/M/T TPERSON[S]U/K GAIEND ACCESS TO PLAY AREA OF PUBLIC HOUSE BY CLIMBING TIMBER FENCE FROM SIDE PEDESTRIAN ACESS, OFFENDERS WENT TO SIDE 2 X 8 WINDOW PANE LAMINATED USING U/K IMPLEMENT SMASHE DIN WINDOW , CLIMBED IN OFFENDERS FORCED OPEN GAMIMNG MACHINE STOLE MONIES LEFT BY WAY OF ENTRY.",},)
  docs = append(docs, map[string]string{field: "Kobe Bryant was one of the best basketball players of all times. Not even Michael Jordan has ever scored 81 points in one game. Munich is really an awesome city, but New York is as well. Yesterday has been the hottest day of the year.",},)

  for i, o := range docs {
    id := fmt.Sprintf("%d", i)
    _, err = client.Index().Index(index).Pipeline(pipeline).Type("doc").Id(id).BodyJson(&o).Do(context.TODO())
    if err != nil {
      log.Fatal(err)
    }
  }

  // Flush
  _, err = client.Flush().Index(index).Do(context.TODO())
  if err != nil {
    log.Fatal(err)
  }
}
