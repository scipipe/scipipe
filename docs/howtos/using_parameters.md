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

TBC...

## Receive parameters dynamically

TBC...

## Handle boolean flags

TBC...

## Handling parameters in re-usable components

TBC...

## Relevant examples

- [Static parameters](https://github.com/scipipe/scipipe/blob/master/examples/static_params/staticparams.go)
- [Receive parameters dynamically](https://github.com/scipipe/scipipe/blob/master/examples/param_channels/params.go)
