package components

import (
	"os"
	"sort"
	"testing"

	"github.com/scipipe/scipipe"
)

func TestIPSorter(t *testing.T) {
	tempDir, err := os.MkdirTemp(".", ".sptsttmp-ipsorter-*")
	scipipe.Check(err)

	fileNames := []string{}
	for _, f := range []string{"b", "c", "a", "z", "d"} {
		fileName := tempDir + "/" + f + ".txt"
		fileNames = append(fileNames, fileName)
		fh, err := os.Create(fileName)
		scipipe.Check(err)

		_, err = fh.WriteString(f)
		scipipe.Check(err)

		err = fh.Close()
		scipipe.Check(err)
	}

	wf := scipipe.NewWorkflow("wf", 4)

	fileSource := NewFileSource(wf, "file-source", fileNames...)

	sorter := NewIPSorter(wf, "sorter", func(ip *scipipe.FileIP) string {
		return ip.Path()
	})
	sorter.In().From(fileSource.Out())

	checker := NewSortChecker(wf, "sort-checker", t)
	checker.In().From(sorter.Out())

	wf.Run()
}

type SortChecker struct {
	scipipe.BaseProcess
	test *testing.T
}

func NewSortChecker(wf *scipipe.Workflow, name string, test *testing.T) *SortChecker {
	p := &SortChecker{
		scipipe.NewBaseProcess(wf, name),
		test,
	}
	p.InitInPort(p, "in")
	wf.AddProc(p)
	return p
}

func (p *SortChecker) In() *scipipe.InPort { return p.InPort("in") }

func (p *SortChecker) Run() {
	ips := map[string]*scipipe.FileIP{}
	sortedStrings := []string{}
	unSortedStrings := []string{}

	// Read all IPs on in-port into data structures
	for inIP := range p.In().Chan {
		sortStr := inIP.Path()
		ips[sortStr] = inIP
		sortedStrings = append(sortedStrings, sortStr)
		unSortedStrings = append(unSortedStrings, sortStr)
	}

	// Sort the strings representing IPs
	sort.Strings(sortedStrings)

	for i, str := range unSortedStrings {
		if str != sortedStrings[i] {
			p.test.Fatalf("Strings were not equal: %s <> %s ... so turns out not sorted", str, sortedStrings[i])
		}
	}
}
