# Acknowledgements

- SciPipe is very heavily dependent on the proven principles form [Flow-Based
  Programming (FBP)](http://www.jpaulmorrison.com/fbp), as invented by [John Paul Morrison](http://www.jpaulmorrison.com/fbp).
  From Flow-based programming, SciPipe uses the ideas of separate network
  (workflow dependency graph) definition, named in- and out-ports,
  sub-networks/sub-workflows and bounded buffers (already available in Go's
  channels) to make writing workflows as easy as possible.
- This library is has been much influenced/inspired also by the
  [GoFlow](https://github.com/trustmaster/goflow) library by [Vladimir Sibirov](https://github.com/trustmaster/goflow).
- Thanks to [Egon Elbre](http://twitter.com/egonelbre) for helpful input on the
  design of the internals of the pipeline, and processes, which greatly
  simplified the implementation.
- This work is financed by faculty grants and other financing for the [Pharmaceutical Bioinformatics group](http://pharmb.io) of [Dept. of
  Pharmaceutical Biosciences](http://www.farmbio.uu.se) at [Uppsala University](http://www.uu.se), and by [Swedish Research Council](http://vr.se)
  through the Swedish [National Bioinformatics Infrastructure Sweden](http://nbis.se).
- Supervisor for the project is [Ola Spjuth](http://www.farmbio.uu.se/research/researchgroups/pb/olaspjuth).
