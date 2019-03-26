# DNA Reverse complement example

A simple example workflow implemented with SciPipe. The workflow computes the
reverse base complement of a string of DNA, using standard UNIX tools.

## Detailed explanation of the code

See the revcomplement.go file for the source code.

On line 4, the SciPipe library is imported, to be later accessed as scipipe. On
line 7, a short string of DNA is defined. On line 9-33, the full workflow is
implemented in the program’s main() function, meaning that it will be executed
when the resulting program is executed. On line 11, a new workflow object (or
“struct” in Go terms) is initiated with a name and the maximum number of cores
to use. On lines 15-25, the workflow components, or processes, are initiated,
each with a name and a shell command pattern. Input file names are defined with
a placeholder on the form {i:INPORTNAME} and outputs on the form
{o:OUTPORTNAME}. The port-name will be used later to access the corresponding
ports for setting up data dependencies. On line 16, a component that writes the
previously defined DNA string to a file is initiated, and on line 17, the file
path pattern for the out-port dna is defined (in this case a static file name).
On line 20, a component that translates each DNA base to its complementary
counterpart is initiated. On line 21, the file path pattern for its only
out-port is defined. In this case, reusing the file path of the file it will
receive on its in-port named in, thus the {i:in} part. The %.txt part removes
.txt from the input path. On line 24, a component that will reverse the DNA
string is initiated. On lines 27-29, data dependencies are defined via the in-
and out-ports defined earlier as part of the shell command patterns. On line
32, the workflow is being run.

## How to run

To run the example, given that you have the [Go toolchain](https://golang.org)
installed (a vertion from at least 1.9), you can run it like this:

```bash
$ go run revcomplement.go 
```

You are then expected to see some log output similar to the following:

```bash
AUDIT   2019/03/26 22:59:43 | workflow:DNA Base Complement Workflow | Starting workflow (Writing log to log/scipipe-20190326-225943-dna-base-complement-workflow.log)
AUDIT   2019/03/26 22:59:43 | Make DNA                         | Executing: echo AAAGCCCGTGGGGGACCTGTTC > dna.txt
AUDIT   2019/03/26 22:59:43 | Make DNA                         | Finished: echo AAAGCCCGTGGGGGACCTGTTC > dna.txt
AUDIT   2019/03/26 22:59:43 | Base Complement                  | Executing: cat ../dna.txt | tr ATCG TAGC > dna.compl.txt
AUDIT   2019/03/26 22:59:43 | Base Complement                  | Finished: cat ../dna.txt | tr ATCG TAGC > dna.compl.txt
AUDIT   2019/03/26 22:59:43 | Reverse                          | Executing: cat ../dna.compl.txt | rev > dna.compl.rev.txt
AUDIT   2019/03/26 22:59:43 | Reverse                          | Finished: cat ../dna.compl.txt | rev > dna.compl.rev.txt
AUDIT   2019/03/26 22:59:43 | workflow:DNA Base Complement Workflow | Finished workflow (Log written to log/scipipe-20190326-225943-dna-base-complement-workflow.log)
```
