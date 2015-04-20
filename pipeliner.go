package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"text/template"
)

var (
	operator = flag.String("operator", "", "operator to generate (batch, flatMap)")
	fn       = flag.String("func", "", "func name to generate")
	from     = flag.String("from", "", "from this type")
	into     = flag.String("into", "", "into this type")
	file     = flag.String("file", "", "output file to write")
)

type pipelineConfig struct {
	Package string
	Version string
	Fn      string
	From    string
	Into    string
}

const batch = `// AUTOMATICALLY GENERATED FILE - DO NOT EDIT
// generated by pipeliner@v{{.Version}}
package {{.Package}}

// {{.Fn}} is a generated implementation of the batch operator
func {{.Fn}}(done <-chan struct{}, max int, in <-chan {{.From}}) <-chan []{{.From}} {
	if max <= 0 {
		panic("batch max must be greater than zero")
	}
	out := make(chan []{{.From}})
	go func() {
		defer close(out)
	Start:
		batch := []{{.From}}{}
		for {
			if len(batch) >= max {
				select {
				case <-done:
					return
				case out <- batch:
				}
				goto Start
			}
			item, active := <-in
			if !active {
				if len(batch) > 0 {
					select {
					case <-done:
						return
					case out <- batch:
					}
				}
				return
			}
			batch = append(batch, item)
		}
	}()
	return out
}`

const flatMap = `// AUTOMATICALLY GENERATED FILE - DO NOT EDIT
// generated by pipeliner@v{{.Version}}
package {{.Package}}

// {{.Fn}} is a generated implementation of the flatMap operator
func {{.Fn}}(outerDone <-chan struct{}, concurrency int, in <-chan {{.From}}, fn func(<-chan struct{}, {{.From}}) (<-chan {{.Into}}, <-chan error)) (<-chan {{.Into}}, <-chan error) {
	out := make(chan {{.Into}})
	errc := make(chan error, 1)
	go func() {
		defer close(out)
		var outerErr error
		defer func() {
			errc <- outerErr
		}()
		// create our own done function that we have the ability to
		// close upon error. Wrap a once function around closing
		// the done function to protect ourselves.
		done := make(chan struct{})
		once := sync.Once{}
		stop := func(err error) {
			once.Do(func() {
				if err != nil {
					outerErr = err
				}
				close(done)
			})
		}
		// outerDone signals stop
		go func() {
			select {
			case <-outerDone:
				stop(nil)
			case <-done:
			}
		}()
		// stop must be called at least once to ensure the
		// goroutine exits.
		defer stop(nil) 

		wg := sync.WaitGroup{}
		wg.Add(concurrency)
		for i := 0; i < concurrency; i++ {
			go func() {
				defer wg.Done()
				for v := range in {
					vch, verrc := fn(done, v)
					for w := range vch {
						select {
						case <-done:
							return
						case out <- w:
						}
					}
					err := <-verrc
					if err != nil {
						stop(err)
						return
					}
				}
			}()
		}
		wg.Wait()
	}()
	return out, errc
}
`

func usage() {
	fmt.Fprintf(os.Stderr, "usage: pipeliner\n")
	flag.PrintDefaults()
	os.Exit(2)
}

func write(config pipelineConfig, source, file string) error {
	out, err := os.Create(file)
	if err != nil {
		return err
	}
	defer out.Close()
	tmpl, err := template.New("pipeliner").Parse(source)
	if err != nil {
		return err
	}
	err = tmpl.Execute(out, config)
	if err != nil {
		return err
	}
	return exec.Command("goimports", "-w", file).Run()
}

func main() {
	flag.Parse()
	// TODO: validate inputs
	config := pipelineConfig{
		Package: os.Getenv("GOPACKAGE"),
		Version: "0.1.0",
		Fn:      *fn,
		From:    *from,
		Into:    *into,
	}

	var err error
	switch *operator {
	case "batch":
		err = write(config, batch, *file)
	case "flatMap":
		err = write(config, flatMap, *file)
	default:
		usage()
	}
	if err != nil {
		log.Fatal(err)
	}
}