package main

import (
	"fmt"
	"log"
	"time"

	"github.com/supershabam/pipeliner"
)

type Person struct {
	Name string
}

func files(ctx pipeliner.Context, num int) <-chan string {
	done := ctx.Done()
	out := make(chan string)
	go func() {
		defer close(out)

		for i := 1; i <= num; i++ {
			select {
			case <-done:
				return
			case out <- fmt.Sprintf("fakefile.%d.json", i):
			}
		}
	}()
	return out
}

func filenameToPeople(ctx pipeliner.Context, file string) <-chan Person {
	done := ctx.Done()
	out := make(chan Person)
	go func() {
		defer close(out)

		// read json file of people, filter them
		// we'll fake that process
		for i := 0; i < 200; i++ {
			if i == 75 {
				err := fmt.Errorf("oops")
				ctx.Cancel(err)
				return
			}
			// can choose to filter out people here
			// simulate a reading lag
			time.Sleep(10 * time.Millisecond)
			person := Person{
				Name: fmt.Sprintf("Lamename %d of %s", i, file),
			}
			select {
			case <-done:
				return
			case out <- person:
			}
		}
	}()
	return out
}

//go:generate pipeliner -file=filenames_to_people.go -from=string -func=filenamesToPeople -into=Person -operator=flatMap
func main() {
	ctx := pipeliner.QueueError()
	filesCh := files(ctx, 200)
	peopleCh := filenamesToPeople(ctx, 10, filesCh, filenameToPeople)
	for person := range peopleCh {
		fmt.Printf("%+v\n", person)
	}
	if len(ctx.Errors()) != 0 {
		log.Fatal(ctx.Errors()[0])
	}
	log.Println("done")
}
