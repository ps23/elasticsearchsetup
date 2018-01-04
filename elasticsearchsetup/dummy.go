package elasticsearchsetup

import (
	"fmt"
	"log"
	"io"
	"sync"
  "sync/atomic"
	"time"
  "context"
	"encoding/json"
	"errors"

  "golang.org/x/sync/errgroup"
	elastic "github.com/olivere/elastic"
)

func Dummy(client *elastic.Client, index string, field string, numSlices int) bool {
	log.Printf("running func Dummy")
	var err error
	// Setup a group of goroutines from the excellent errgroup package
	g, ctx := errgroup.WithContext(context.TODO())

	// Hits channel will be sent to from the first set of goroutines and consumed by the second
	type hit struct {
		Slice int
		Hit   elastic.SearchHit
	}
	hitsc := make(chan hit)

	begin := time.Now()

	// Start a number of goroutines to parallelize scrolling
	var wg sync.WaitGroup
	for i := 0; i < numSlices; i++ {
		wg.Add(1)

		slice := i

		// Prepare the query
		var query elastic.Query
		if field == "" {
			query = elastic.NewMatchAllQuery()
		} else {
			query = elastic.NewExistsQuery(field)
		}

		// Prepare the slice
		sliceQuery := elastic.NewSliceQuery().Id(i).Max(numSlices)

		// Start goroutine for this sliced scroll
		g.Go(func() error {
			defer wg.Done()

			ss := elastic.NewSearchSource().Query(query)

			src, err := ss.Source()
			if err != nil {
				return err
			}
			data, _ := json.Marshal(src)
			log.Printf("%s", string(data))

			svc := client.Scroll(index).SearchSource(ss).Slice(sliceQuery)

			for {
				res, err := svc.Do(ctx)
				if err == io.EOF {
					break
				}
				if err != nil {
    			// Get *elastic.Error which contains additional information
    			e, ok := err.(*elastic.Error)
    			if !ok {
        		// This shouldn't happen
    			}
    			log.Printf("Elastic failed with status %d and error %s.", e.Status, e.Details.Reason)
				}

				for _, searchHit := range res.Hits.Hits {
					// Pass the hit to the hits channel, which will be consumed below
					select {
					case hitsc <- hit{Slice: slice, Hit: *searchHit}:
					case <-ctx.Done():
						return ctx.Err()
					}
				}
			}
			return nil
		})
	}

	go func() {
		// Wait until all scrolling is done
		wg.Wait()
		close(hitsc)
	}()

	// Second goroutine will consume the hits sent from the workers in first set of goroutines
	var total uint64
	var bulkSize int = 1000
	totals := make([]uint64, numSlices)
	g.Go(func() error {
		bulk := client.Bulk().Index(index)

		for hit := range hitsc {

			abc := []map[string]string{{"term": "test"}}
			type doc struct {
								Extractions []map[string]string `json:"extractions"`
				}

			d := doc {
				Extractions: abc,
			}

			req := elastic.NewBulkUpdateRequest().Id(hit.Hit.Id).Type("doc").Doc(d)

			// Enqueue the document
			bulk.Add(req)
			if bulk.NumberOfActions() >= bulkSize {
				// Commit
				res, err := bulk.Do(ctx)
				if err != nil {
					return err
				}
				if res.Errors {
						// Look up the failed documents with res.Failed(), and e.g. recommit
						fmt.Println(res.Failed());
						return errors.New("bulk commit failed")
				}
				// "bulk" is reset after Do, so you can reuse it
			}

			select {
			default:
			case <-ctx.Done():
				return ctx.Err()
			}

			// Count the hits here.
			atomic.AddUint64(&totals[hit.Slice], 1)

			select {
			default:
			case <-ctx.Done():
				return ctx.Err()
			}
		}

		// Commit the final batch before exiting
		if bulk.NumberOfActions() > 0 {
			fmt.Println("Commit the final batch before exiting");
			_, err = bulk.Do(ctx)
			if err != nil {
				return err
			}
		}

		return nil
	})

	// Wait until all goroutines are finished
	if err := g.Wait(); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Scrolled through a total of %d documents in %v\n", total, time.Since(begin))
	for i := 0; i < numSlices; i++ {
		fmt.Printf("Slice %2d received %d documents\n", i, totals[i])
	}

  return true;
}
