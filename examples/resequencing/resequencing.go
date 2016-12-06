// Implementation (work in progress) of the resequencing analysis pipeline used
// to teach the introductory NGS bioinformatics analysis course at SciLifeLab
// as described on this page:
// http://uppnex.se/twiki/do/view/Courses/NgsIntro1502/ResequencingAnalysis.html
// Prerequisites:
// - Samtools
// - BWA
// - Picard
// - GATK
// Install all tools except GATK like this on X/L/K/Ubuntu:
// sudo apt-get install samtools bwa picard-tools
// (GATK needs to be downloaded and installed manually from www.broadinstitute.org/gatk)
package main

import (
	"fmt"

	sp "github.com/scipipe/scipipe"
	plib "github.com/scipipe/scipipe/proclib"
)

const (
	fastq_base_url = "http://bioinfo.perdanauniversity.edu.my/tein4ngs/ngspractice/"
	fastq_file_pat = "%s.ILLUMINA.low_coverage.4p_%s.fq"
	ref_base_url   = "http://ftp.ensembl.org/pub/release-75/fasta/homo_sapiens/dna/"
	ref_file       = "Homo_sapiens.GRCh37.75.dna.chromosome.17.fa"
	ref_file_gz    = "Homo_sapiens.GRCh37.75.dna.chromosome.17.fa.gz"
	vcf_base_url   = "http://ftp.1000genomes.ebi.ac.uk/vol1/ftp/phase1/analysis_results/integrated_call_sets/"
	vcf_file       = "ALL.chr17.integrated_phase1_v3.20101123.snps_indels_svs.genotypes.vcf.gz"
)

var (
	individuals = []string{"NA06984", "NA12489"}
	samples     = []string{"1", "2"}
)

func main() {
	// Create ungzip command pattern for later use
	ungzCmdPat := "gunzip -c {i:in} > {o:out}"

	// --------------------------------------------------------------------------------
	// Initialize pipeline runner
	// --------------------------------------------------------------------------------

	pipeRun := sp.NewPipelineRunner()
	sink := sp.NewSink()

	// --------------------------------------------------------------------------------
	// Download Reference Genome
	// --------------------------------------------------------------------------------

	dlRefGz := sp.NewFromShell("dl_gzipped",
		"wget -O {o:outfile} "+ref_base_url+ref_file_gz)
	pipeRun.AddProcess(dlRefGz)
	dlRefGz.SetPathStatic("outfile", ref_file_gz)

	// --------------------------------------------------------------------------------
	// Unzip ref file
	// --------------------------------------------------------------------------------

	ungzRef := sp.NewFromShell("ungzRef", ungzCmdPat)
	ungzRef.SetPathReplace("in", "out", ".gz", "")
	ungzRef.In["in"].Connect(dlRefGz.Out["outfile"])
	pipeRun.AddProcess(ungzRef)

	// Create a FanOut so multiple downstream processes can read from the
	// ungzip process
	refFOut := plib.NewFanOut()
	refFOut.InFile.Connect(ungzRef.Out["out"])
	pipeRun.AddProcess(refFOut)

	// --------------------------------------------------------------------------------
	// Index Reference Genome
	// --------------------------------------------------------------------------------

	indxRef := sp.NewFromShell("Index Ref", "bwa index -a bwtsw {i:index}; echo done > {o:done}")
	indxRef.SetPathExtend("index", "done", ".indexed")
	indxRef.In["index"].Connect(refFOut.GetOutPort("index_ref"))
	pipeRun.AddProcess(indxRef)

	idxDnFO := plib.NewFanOut()
	idxDnFO.InFile.Connect(indxRef.Out["done"])
	pipeRun.AddProcess(idxDnFO)

	// Create (multi-level) maps where we can gather outports from processes
	// for each for loop iteration and access them in the merge step later
	outPorts := make(map[string]map[string]map[string]*sp.FilePort)
	for _, indv := range individuals {
		outPorts[indv] = make(map[string]map[string]*sp.FilePort)
		for _, smpl := range samples {
			outPorts[indv][smpl] = make(map[string]*sp.FilePort)

			// --------------------------------------------------------------------------------
			// Download FastQ component
			// --------------------------------------------------------------------------------

			file_name := fmt.Sprintf(fastq_file_pat, indv, smpl)
			dlFastq := sp.NewFromShell("dl_fastq",
				"wget -O {o:fastq} "+fastq_base_url+file_name)
			dlFastq.SetPathStatic("fastq", file_name)
			pipeRun.AddProcess(dlFastq)

			fqFnOut := plib.NewFanOut()
			fqFnOut.InFile.Connect(dlFastq.Out["fastq"])
			pipeRun.AddProcess(fqFnOut)

			// Save outPorts for later use
			outPorts[indv][smpl]["fastq"] = fqFnOut.GetOutPort("merg")

			// --------------------------------------------------------------------------------
			// BWA Align
			// --------------------------------------------------------------------------------

			bwaAlgn := sp.NewFromShell("bwa_aln",
				"bwa aln {i:ref} {i:fastq} > {o:sai} # {i:idxdone}")
			bwaAlgn.SetPathExtend("fastq", "sai", ".sai")
			bwaAlgn.In["ref"].Connect(refFOut.GetOutPort("bwa_aln_" + indv + "_" + smpl))
			bwaAlgn.In["idxdone"].Connect(idxDnFO.GetOutPort("bwa_aln_" + indv + "_" + smpl))
			bwaAlgn.In["fastq"].Connect(fqFnOut.GetOutPort("bwa_aln"))
			pipeRun.AddProcess(bwaAlgn)

			// Save outPorts for later use
			outPorts[indv][smpl]["sai"] = bwaAlgn.Out["sai"]
		}

		// --------------------------------------------------------------------------------
		// Merge
		// --------------------------------------------------------------------------------

		// This one is is needed so bwaMerg can take a proper parameter for
		// individual, which it uses to generate output paths
		indParamGen := plib.NewStringGenerator(indv)
		pipeRun.AddProcess(indParamGen)

		// bwa sampe process
		bwaMerg := sp.NewFromShell("merge_"+indv,
			"bwa sampe {i:ref} {i:sai1} {i:sai2} {i:fq1} {i:fq2} > {o:merged} # {i:refdone} {p:indv}")
		bwaMerg.PathFormatters["merged"] = func(t *sp.SciTask) string {
			return fmt.Sprintf("%s.merged.sam", t.Params["indv"])
		}
		// Connect
		bwaMerg.In["ref"].Connect(refFOut.GetOutPort("merg_" + indv))
		bwaMerg.In["refdone"].Connect(idxDnFO.GetOutPort("merg_" + indv))
		bwaMerg.In["sai1"].Connect(outPorts[indv]["1"]["sai"])
		bwaMerg.In["sai2"].Connect(outPorts[indv]["2"]["sai"])
		bwaMerg.In["fq1"].Connect(outPorts[indv]["1"]["fastq"])
		bwaMerg.In["fq2"].Connect(outPorts[indv]["2"]["fastq"])
		bwaMerg.ParamPorts["indv"].Connect(indParamGen.Out)
		// Add to runner
		pipeRun.AddProcess(bwaMerg)

		sink.Connect(bwaMerg.Out["merged"])
	}

	// --------------------------------------------------------------------------------
	// Run pipeline
	// --------------------------------------------------------------------------------

	pipeRun.AddProcess(sink)
	pipeRun.Run()
}
