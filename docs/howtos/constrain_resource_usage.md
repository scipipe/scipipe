It is important to carefully manage how much resources (CPU and memory) your
workflows are using, so that you don't overbook you compute node(s).

In SciPipe you can do that using two settings:

- Max concurrent tasks, which is set on the workflow level, when initiating a new workflow.
- Cores per tasks, that can be set on processes after they are initialized.

Max concurrent tasks is a required setting when initializing workflows, while
cores per task can be left to the default, which is 1 core per task.

You might want to change this number if for example you have a software that
uses more memory than the available memory on your computer divided by the max
concurrent tasks number you have set.

For example, if you have 8GB of free memory, and have set max concurrent tasks
on your workflow to 4, but you have a process whose commandline application
uses not 2GB of memory, but 4GB, then you might want to set cores per tasks for
that process to 2, so that it gets the double amount of memory.

In practice, you set cores per task by setting the field `CoresPerTask` on the process struct, after it is initiated. 

## Example

```go
foo := scipipe.NewProc("foo", "echo foo > {o:foofile}")
foo.CoresPerTask = 2
```
