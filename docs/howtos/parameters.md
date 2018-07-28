Parameters are arguments sent to commands as flags, or unnamed values, or
sometimes just the occurance of flags.

SciPipe does not provide one unified way to handle parameters, but instead
suggest a few different strategies, dependent on the usage pattern. This is
because it turns out that there is a very large variety in how parameters can
be used with shell commands.

To keep SciPipe a small and flexible tool, we instead mostly leave the choice
up to the workflow author to create a solution for each case, using a few helper
tools provided with SciPipe, but also all the programming facilities built in to
the Go programming language.

Below we will discuss how to handle the most common uses for for parameters in
SciPipe. For any more complicated use cases not covered here, please refer to
the [mailing list](https://groups.google.com/forum/#!forum/scipipe) or the
[chat](https://gitter.im/scipipe/scipipe), to ask your question.

## Static parameters

If parameters in your shell command is always, the same, you can just add them
"manually" to the shell command pattern used to create your process.

For example, if you always want to write the string "hello" to output files,
you could create your processes with this string added manually:

```go
helloWriter := scipipe.NewProc("helloWriter", "echo hello > {o:outfile}")
```

### See also

- [Static parameters example](https://github.com/scipipe/scipipe/blob/master/examples/static_params/staticparams.go)

## Receive parameters dynamically

Receiving parameters dynamically is a much more technically demandning solution
than using static parameters.

The idea is that by using placeholders for parameter values in a command, each
parameter for a particular process, will automatically get a channel of type
string, on which it can receive values. When the process is ready to execute
another shell command, it receives one item on each parameter ports, in
addition to receiving one file on each (file-)in-port, and merges the values
into the shell command, before executing it.

An example of this would be a little too complicated to cover briefly on this
page, so please instead see the [dynamic parameters example](https://github.com/scipipe/scipipe/blob/master/examples/param_channels/params.go).
In the [Run method of the Combinatorics task](https://github.com/scipipe/scipipe/blob/master/examples/param_channels/params.go#L58-L70)
you will find the code used to send values (all combinations of values in three
arrays of lenght 3, in this case).

### See also

- [Dynamic parameters example](https://github.com/scipipe/scipipe/blob/master/examples/param_channels/params.go)

## Handle boolean flags

*Topic coming soon. Please add it as a support request in the [issue tracker](https://github.com/scipipe/scipipe/issues)
if you need this information fast, and we can prioritize writing it asap.*

## Handling parameters in re-usable components

*Topic coming soon. Please add it as a support request in the [issue tracker](https://github.com/scipipe/scipipe/issues)
if you need this information fast, and we can prioritize writing it asap.*

## Relevant examples

- [Static parameters](https://github.com/scipipe/scipipe/blob/master/examples/static_params/staticparams.go)
- [Receive parameters dynamically](https://github.com/scipipe/scipipe/blob/master/examples/param_channels/params.go)
