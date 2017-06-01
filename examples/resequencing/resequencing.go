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
	"github.com/scipipe/scipipe/components"
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

	pr := sp.NewPipelineRunner()
	sink := sp.NewSink()

	// --------------------------------------------------------------------------------
	// Download Reference Genome
	// --------------------------------------------------------------------------------

	downloadRef := pr.NewFromShell("download_ref",
		"wget -O {o:outfile} "+ref_base_url+ref_file_gz)
	downloadRef.SetPathStatic("outfile", ref_file_gz)

	// --------------------------------------------------------------------------------
	// Unzip ref file
	// --------------------------------------------------------------------------------

	ungzipRef := pr.NewFromShell("ugzip_ref", ungzCmdPat)
	ungzipRef.SetPathReplace("in", "out", ".gz", "")
	ungzipRef.In("in").Connect(downloadRef.Out("outfile"))

	// Create a FanOut so multiple downstream processes can read from the
	// ungzip process
	refFanOut := components.NewFanOut()
	refFanOut.InFile.Connect(ungzipRef.Out("out"))
	pr.AddProcess(refFanOut)

	// --------------------------------------------------------------------------------
	// Index Reference Genome
	// --------------------------------------------------------------------------------

	indexRef := pr.NewFromShell("index_ref", "bwa index -a bwtsw {i:index}; echo done > {o:done}")
	indexRef.SetPathExtend("index", "done", ".indexed")
	indexRef.In("index").Connect(refFanOut.GetOutPort("index_ref"))

	indexDoneFanOut := components.NewFanOut()
	indexDoneFanOut.InFile.Connect(indexRef.Out("done"))
	pr.AddProcess(indexDoneFanOut)

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
			downloadFastQ := pr.NewFromShell("download_fastq_"+indv+"_"+smpl,
				"wget -O {o:fastq} "+fastq_base_url+file_name)
			downloadFastQ.SetPathStatic("fastq", file_name)

			fastQFanOut := components.NewFanOut()
			fastQFanOut.InFile.Connect(downloadFastQ.Out("fastq"))
			pr.AddProcess(fastQFanOut)

			// Save outPorts for later use
			outPorts[indv][smpl]["fastq"] = fastQFanOut.GetOutPort("merg")

			// --------------------------------------------------------------------------------
			// BWA Align
			// --------------------------------------------------------------------------------

			bwaAlign := pr.NewFromShell("bwa_aln",
				"bwa aln {i:ref} {i:fastq} > {o:sai} # {i:idxdone}")
			bwaAlign.SetPathExtend("fastq", "sai", ".sai")
			bwaAlign.In("ref").Connect(refFanOut.GetOutPort("bwa_aln_" + indv + "_" + smpl))
			bwaAlign.In("idxdone").Connect(indexDoneFanOut.GetOutPort("bwa_aln_" + indv + "_" + smpl))
			bwaAlign.In("fastq").Connect(fastQFanOut.GetOutPort("bwa_aln"))

			// Save outPorts for later use
			outPorts[indv][smpl]["sai"] = bwaAlign.Out("sai")
		}

		// --------------------------------------------------------------------------------
		// Merge
		// --------------------------------------------------------------------------------

		// This one is is needed so bwaMerg can take a proper parameter for
		// individual, which it uses to generate output paths
		indParamGen := components.NewStringGenerator(indv)
		pr.AddProcess(indParamGen)

		// bwa sampe process
		bwaMerg := pr.NewFromShell("merge_"+indv,
			"bwa sampe {i:ref} {i:sai1} {i:sai2} {i:fq1} {i:fq2} > {o:merged} # {i:refdone} {p:indv}")
		bwaMerg.SetPathCustom("merged", func(t *sp.SciTask) string {
			return fmt.Sprintf("%s.merged.sam", t.Params["indv"])
		})
		// Connect
		bwaMerg.In("ref").Connect(refFanOut.GetOutPort("merg_" + indv))
		bwaMerg.In("refdone").Connect(indexDoneFanOut.GetOutPort("merg_" + indv))
		bwaMerg.In("sai1").Connect(outPorts[indv]["1"]["sai"])
		bwaMerg.In("sai2").Connect(outPorts[indv]["2"]["sai"])
		bwaMerg.In("fq1").Connect(outPorts[indv]["1"]["fastq"])
		bwaMerg.In("fq2").Connect(outPorts[indv]["2"]["fastq"])
		bwaMerg.ParamPort("indv").Connect(indParamGen.Out)
		// Add to runner

		sink.Connect(bwaMerg.Out("merged"))
	}

	// --------------------------------------------------------------------------------
	// Run pipeline
	// --------------------------------------------------------------------------------

	pr.AddProcess(sink)
	pr.Run()
}
