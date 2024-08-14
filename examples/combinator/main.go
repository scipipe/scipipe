package main

import (
	"fmt"
	"path/filepath"
	"strings"
	"sync"

	"github.com/scipipe/scipipe"
	sp "github.com/scipipe/scipipe"
)

// ------------------------------------------------------------------------------
// Demo workflow
// ------------------------------------------------------------------------------

func main() {
	// Create a workflow, using 4 cpu cores
	wf := sp.NewWorkflow("my_workflow", 1)

	// Initialize combinator process
	generateCombinations := func() (paths map[string][]string, params map[string][]string) {
		paths = map[string][]string{
			"infile": []string{},
		}
		params = map[string][]string{
			"l": []string{},
			"n": []string{},
			"u": []string{},
		}

		globPaths, err := filepath.Glob("infiles/*.txt")
		scipipe.Check(err)
		for _, f := range globPaths {
			for _, l := range []string{"a", "b", "c"} {
				for _, n := range []string{"1", "2", "3"} {
					for _, u := range []string{"A", "B", "C"} {
						paths["infile"] = append(paths["infile"], f)
						params["l"] = append(params["l"], l)
						params["n"] = append(params["n"], n)
						params["u"] = append(params["u"], u)
					}
				}
			}
		}

		return paths, params
	}
	combinator := NewCombinator(wf, "combinator", generateCombinations)

	// Initialize and Connect Greeter
	greeter := wf.NewProc("fooer", "echo $(cat {i:basefile}){p:l}{p:n}{p:u} > {o:combinations}")
	greeter.In("basefile").From(combinator.OutPort("infile"))
	greeter.InParam("l").From(combinator.OutParamPort("l"))
	greeter.InParam("n").From(combinator.OutParamPort("n"))
	greeter.InParam("u").From(combinator.OutParamPort("u"))
	greeter.SetOutFunc("combinations", func(t *scipipe.Task) string {
		basePath := strings.ReplaceAll(filepath.Base(t.InPath("basefile")), ".txt", "")
		return fmt.Sprintf("out/%s-%s%s%s.txt", basePath, t.Param("l"), t.Param("n"), t.Param("u"))
	})

	// Run the workflow
	wf.Run()
}

// ------------------------------------------------------------------------------
// Combinator component implementation
// ------------------------------------------------------------------------------

type Combinator struct {
	scipipe.BaseProcess
	fun func() (map[string][]string, map[string][]string)
}

func NewCombinator(wf *scipipe.Workflow, name string, newFun func() (map[string][]string, map[string][]string)) *Combinator {
	comb := &Combinator{
		BaseProcess: scipipe.NewBaseProcess(wf, name),
		fun:         newFun,
	}
	paths, params := comb.fun()
	for pName, _ := range paths {
		if _, ok := comb.OutPorts()[pName]; !ok {
			comb.InitOutPort(comb, pName)
		}
	}
	for pName, _ := range params {
		if _, ok := comb.OutParamPorts()[pName]; !ok {
			comb.InitOutParamPort(comb, pName)
		}
	}
	wf.AddProc(comb)
	return comb
}

func (p *Combinator) Run() {
	defer p.CloseAllOutPorts()

	paths, params := p.fun()

	wgf := &sync.WaitGroup{}
	for pName, pathList := range paths {
		pName := pName
		pathList := pathList
		wgf.Add(1)
		go func() {
			for _, path := range pathList {
				ip, err := scipipe.NewFileIP(path)
				scipipe.Check(err)
				p.OutPort(pName).Send(ip)
			}
			wgf.Done()
		}()
	}
	wgp := &sync.WaitGroup{}
	for pName, paramList := range params {
		pName := pName
		paramList := paramList
		wgp.Add(1)
		go func() {
			for _, param := range paramList {
				p.OutParamPort(pName).Send(param)
			}
			wgp.Done()
		}()
	}
	wgf.Wait()
	wgp.Wait()
}
