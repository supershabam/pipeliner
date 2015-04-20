package main

import (
	"fmt"
	"log"
	"time"
)

type Person struct {
	Name string
}

func files(done <-chan struct{}, num int) <-chan string {
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

func filenameToPeople(done <-chan struct{}, file string) (<-chan Person, <-chan error) {
	out := make(chan Person)
	errc := make(chan error, 1)
	go func() {
		defer close(out)
		var err error
		defer func() {
			errc <- err
		}()

		// read json file of people, filter them
		// we'll fake that process
		for i := 0; i < 200; i++ {
			if i == 75 {
				err = fmt.Errorf("oops")
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
	return out, errc
}

//go:generate pipeliner -file=filenames_to_people.go -from=string -func=filenamesToPeople -into=Person -operator=flatMap
func main() {
	done := make(chan struct{})
	filesc := files(done, 200)
	people, peopleErrc := filenamesToPeople(done, 10, filesc, filenameToPeople)
	for person := range people {
		fmt.Printf("%+v\n", person)
	}
	if err := <-peopleErrc; err != nil {
		log.Fatal(err)
	}
	log.Println("done")
}
