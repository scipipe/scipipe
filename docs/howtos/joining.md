If you want to join multiple files, you can do that by first using the [StreamToSubstream](https://godoc.org/github.com/scipipe/scipipe/components#StreamToSubStream) component in
combination with an in-port command pattern where the `join` pattern is used.

A concrete usage of this can be seen [in the DNA Cancer analysis workflow](https://github.com/pharmbio/scipipe-demo/blob/fdb9888/dnacanceranalysis/dnacanceranalysiswf.go#L89-L90).
In this example, we see that the `bams` inport is defined like so: `{i:bams|join: }`. This means that
it will receive IPs on the `bams` inport, and join their filenames, with a space between each.
On the next row, we connect this in-port to the [OutSubStream](https://godoc.org/github.com/scipipe/scipipe/components#StreamToSubStream.OutSubStream) of the [StreamToSubstream](https://godoc.org/github.com/scipipe/scipipe/components#StreamToSubStream) component (in this
example, the StreamToSubstream component was stored in a map, but that is specific to the example
code, and not for using it in general).

## More info

See the [Concatenator component](https://godoc.org/github.com/scipipe/scipipe/components#Concatenator).

Also see the [page about Scatter/Gather](/howtos/scatter_gather/).
