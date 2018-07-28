[SciPipe 0.8.0](https://github.com/scipipe/scipipe/releases/tag/v0.8.0)
introduced experimental support for converting audit logs (those
`.audit.json` files produced to accompany all output files from SciPipe)
into other formats, such as HTML, TeX (for further conversion to PDF) or
even executable Bash-scripts. Here's how to do it.

## Convert audit log to HTML

Given that you have an audit log file with the name `myfile.audit.json`,
then execute:

```bash
scipipe audit2html myfile.audit.json
```

This will produce an HTML file named `myfile.audit.html`, which you can
view in a web browser.

## Convert audit log to TeX

Given that you have an audit log file with the name `myfile.audit.json`,
then execute:

```bash
scipipe audit2tex myfile.audit.json
```

This will produce an HTML file named `myfile.audit.tex`, which you can
either edit manually, or convert directly to PDF using the `pdflatex`
command like so:

```bash
pdflatex myfile.audit.tex
```

Converting to PDF requires that you have a TeX installation on your system.
On Ubuntu, you can install the base package or TeX live with `sudo apt-get
install texlive-base`.

## Convert audit log to Bash

Given that you have an audit log file with the name `myfile.audit.json`,
then execute:

```bash
scipipe audit2bash myfile.audit.json
```

This will produce a Bash-file named `myfile.audit.sh`, which you can
execute like so:

```bash
sh myfile.audit.sh
```

... in order to reproduce the file again from scratch, if it is removed,
given that you have all the dependent files and tools installed on your
system.