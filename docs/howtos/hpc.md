This is [being worked on right now (issue #38)](https://github.com/scipipe/scipipe/issues/38).

What you can do right now, is to use the `Prepend` field in processes, to add a
[salloc](https://slurm.schedmd.com/salloc.html) command string (in the case of
SLURM), or any analogous blocking command to other resource managers.

So, something like this (See on the third line how the salloc-line is added to the process):

```go
wf := scipipe.NewWorkflow("Hello_World_Workflow", 4)
myProc := wf.NewProc("hello_world", "echo Hello World; sleep 10;")
myProc.Prepend = "salloc -A projectABC123 -p core -t 1:00 -J HelloWorld"
```

*(Beware: This is not a full code example, and won't compile without some more boilerplate, which you can find in the introductory examples)*

You can find the updated GoDoc for the process struct [here](http://godoc.org/github.com/scipipe/scipipe#Process).
