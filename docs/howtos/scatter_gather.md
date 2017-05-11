There is [work going on to add better scatter/gather support in SciPipe (Issue #20)](https://github.com/scipipe/scipipe/issues/20).
In the meanwhile, have a look at [this example on GitHub](https://github.com/scipipe/scipipe/blob/master/examples/scatter_gather/scattergather.go)
which demonstrates one way of doing a scatter gather operation, using the Splitter component ([see line 24](https://github.com/scipipe/scipipe/blob/master/examples/scatter_gather/scattergather.go#L24))
and two concatenator components ([see lines 38-39](https://github.com/scipipe/scipipe/blob/master/examples/scatter_gather/scattergather.go#L38-L39))
to do the scatter, and gather, operations respectively.
