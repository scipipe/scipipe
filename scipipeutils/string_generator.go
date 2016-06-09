package scipipeutils

import "github.com/scipipe/scipipe"

// StringGenerator takes a number of strings and returns a generator process
// which sends the strings, one by one, on its `Out` port
type StringGenerator struct {
	scipipe.Process
	Strings []string
	Out     chan string
}

// Instantiate a new StringGenerator
func NewStringGenerator(strings ...string) *StringGenerator {
	return &StringGenerator{
		Out:     make(chan string, scipipe.BUFSIZE),
		Strings: strings,
	}
}

// Run the StringGenerator
func (proc *StringGenerator) Run() {
	defer close(proc.Out)
	for _, str := range proc.Strings {
		proc.Out <- str
	}
}
