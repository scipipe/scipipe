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

	. "github.com/scipipe/scipipe"
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

	InitLogDebug()

	// --------------------------------------------------------------------------------
	// Initialize pipeline runner
	// --------------------------------------------------------------------------------

	wf := NewWorkflow("resequencing_wf")
	sink := NewSink("sink")

	// --------------------------------------------------------------------------------
	// Download Reference Genome
	// --------------------------------------------------------------------------------
	downloadRefCmd := "wget -O {o:outfile} " + ref_base_url + ref_file_gz
	downloadRef := wf.NewProc("download_ref", downloadRefCmd)
	downloadRef.SetPathStatic("outfile", ref_file_gz)

	// --------------------------------------------------------------------------------
	// Unzip ref file
	// --------------------------------------------------------------------------------
	ungzipRefCmd := "gunzip -c {i:in} > {o:out}"
	ungzipRef := wf.NewProc("ugzip_ref", ungzipRefCmd)
	ungzipRef.SetPathReplace("in", "out", ".gz", "")
	ungzipRef.In("in").Connect(downloadRef.Out("outfile"))

	// Create a FanOut so multiple downstream processes can read from the
	// ungzip process
	refFanOut := components.NewFanOut("ref_fanout")
	refFanOut.InFile.Connect(ungzipRef.Out("out"))
	wf.Add(refFanOut)

	// --------------------------------------------------------------------------------
	// Index Reference Genome
	// --------------------------------------------------------------------------------
	indexRef := wf.NewProc("index_ref", "bwa index -a bwtsw {i:index}; echo done > {o:done}")
	indexRef.SetPathExtend("index", "done", ".indexed")
	indexRef.In("index").Connect(refFanOut.Out("index_ref"))

	indexDoneFanOut := components.NewFanOut("indexdone_fanout")
	indexDoneFanOut.InFile.Connect(indexRef.Out("done"))
	wf.Add(indexDoneFanOut)

	// Create (multi-level) maps where we can gather outports from processes
	// for each for loop iteration and access them in the merge step later
	outPorts := map[string]map[string]map[string]*FilePort{}
	for _, indv := range individuals {
		outPorts[indv] = map[string]map[string]*FilePort{}
		for _, smpl := range samples {
			outPorts[indv][smpl] = map[string]*FilePort{}

			// --------------------------------------------------------------------------------
			// Download FastQ component
			// --------------------------------------------------------------------------------
			file_name := fmt.Sprintf(fastq_file_pat, indv, smpl)
			downloadFastQCmd := "wget -O {o:fastq} " + fastq_base_url + file_name
			downloadFastQ := wf.NewProc("download_fastq_"+indv+"_"+smpl, downloadFastQCmd)
			downloadFastQ.SetPathStatic("fastq", file_name)

			fastQFanOut := components.NewFanOut("fastq_fanout")
			fastQFanOut.InFile.Connect(downloadFastQ.Out("fastq"))
			wf.Add(fastQFanOut)

			// Save outPorts for later use
			outPorts[indv][smpl]["fastq"] = fastQFanOut.Out("merg")

			// --------------------------------------------------------------------------------
			// BWA Align
			// --------------------------------------------------------------------------------
			bwaAlignCmd := "bwa aln {i:ref} {i:fastq} > {o:sai} # {i:idxdone}"
			bwaAlign := wf.NewProc("bwa_aln", bwaAlignCmd)
			bwaAlign.SetPathExtend("fastq", "sai", ".sai")
			bwaAlign.In("ref").Connect(refFanOut.Out("bwa_aln_" + indv + "_" + smpl))
			bwaAlign.In("idxdone").Connect(indexDoneFanOut.Out("bwa_aln_" + indv + "_" + smpl))
			bwaAlign.In("fastq").Connect(fastQFanOut.Out("bwa_aln"))

			// Save outPorts for later use
			outPorts[indv][smpl]["sai"] = bwaAlign.Out("sai")
		}

		// --------------------------------------------------------------------------------
		// Merge
		// --------------------------------------------------------------------------------
		// This one is is needed so bwaMergecan take a proper parameter for
		// individual, which it uses to generate output paths
		indParamGen := components.NewStringGen(indv)
		wf.Add(indParamGen)

		// bwa sampe process
		bwaMergeCmd := "bwa sampe {i:ref} {i:sai1} {i:sai2} {i:fq1} {i:fq2} > {o:merged} # {i:refdone} {p:indv}"
		bwaMerge := wf.NewProc("merge_"+indv, bwaMergeCmd)
		bwaMerge.SetPathCustom("merged", func(t *SciTask) string { return fmt.Sprintf("%s.merged.sam", t.Params["indv"]) })
		bwaMerge.In("ref").Connect(refFanOut.Out("bwa_merge_" + indv))
		bwaMerge.In("refdone").Connect(indexDoneFanOut.Out("bwa_merge_" + indv))
		bwaMerge.In("sai1").Connect(outPorts[indv]["1"]["sai"])
		bwaMerge.In("sai2").Connect(outPorts[indv]["2"]["sai"])
		bwaMerge.In("fq1").Connect(outPorts[indv]["1"]["fastq"])
		bwaMerge.In("fq2").Connect(outPorts[indv]["2"]["fastq"])
		bwaMerge.PP("indv").Connect(indParamGen.Out)

		sink.Connect(bwaMerge.Out("merged"))
	}

	// --------------------------------------------------------------------------------
	// Run pipeline
	// --------------------------------------------------------------------------------

	wf.SetDriver(sink)
	wf.Run()
}
