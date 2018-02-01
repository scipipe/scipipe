SciPipe can stream the output via UNIX [named pipes (or "FIFO files")](https://en.wikipedia.org/wiki/Named_pipe).

Streaming can be turned on, on out-ports when creating processes with
`NewProc()`, by using `{os:outport_name}` as placeholder, instead of the
normal `{o:outport_name}` (note the addisional "s")

You can see how this is used in [this example on GitHub](https://github.com/scipipe/scipipe/blob/master/examples/fifo/fifo.go#L14).

Note that when streaming, you will not get an output file for the output in
question.

Note also that you still have to provide a path formatting strategy (via some
of the `Process.SetPath...()` functions, or by manually adding one to
`Process.PathFormatters`. This is because a uniqe file name is needed in
order to create any audit files, as well as to give a unique name for the named
pipe.

## See also

- [Streaming example on GitHub](https://github.com/scipipe/scipipe/blob/master/examples/fifo/fifo.go#L14).
