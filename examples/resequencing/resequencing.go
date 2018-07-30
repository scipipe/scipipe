// Implementation (work in progress) of the resequencing analysis pipeline used
// to teach the introductory NGS bioinformatics analysis course at SciLifeLab
// as described on this page:
// http://uppnex.se/twiki/do/view/Courses/NgsIntro1502/ResequencingAnalysis.html
//
// Prerequisites: Samtools, BWA, Picard, GATK.  You can install all tools
// except GATK on X/L/K/Ubuntu, with this command:
//
//   sudo apt-get install samtools bwa picard-tools
//
// GATK needs to be downloaded and installed manually from:
// http://www.broadinstitute.org/gatk
package main

import (
	"fmt"

	sp "github.com/scipipe/scipipe"
)

// ------------------------------------------------------------------------------------
// Set up static stuff like paths
// ------------------------------------------------------------------------------------
const (
	fastqBaseUrl = "http://bioinfo.perdanauniversity.edu.my/tein4ngs/ngspractice/"
	fastQFilePtn = "%s.ILLUMINA.low_coverage.4p_%s.fq"
	refBaseUrl   = "http://ftp.ensembl.org/pub/release-75/fasta/homo_sapiens/dna/"
	refFile      = "Homo_sapiens.GRCh37.75.dna.chromosome.17.fa"
	refFileGz    = "Homo_sapiens.GRCh37.75.dna.chromosome.17.fa.gz"
	vcfBaseUrl   = "http://ftp.1000genomes.ebi.ac.uk/vol1/ftp/phase1/analysis_results/integrated_call_sets/"
	vcfFile      = "ALL.chr17.integrated_phase1_v3.20101123.snps_indels_svs.genotypes.vcf.gz"
)

// ------------------------------------------------------------------------------------
// Set up the main input parameters to the workflow
// ------------------------------------------------------------------------------------
var (
	individuals = []string{"NA06984", "NA12489"}
	samples     = []string{"1", "2"}
)

func main() {
	// --------------------------------------------------------------------------------
	// Initialize Workflow
	// --------------------------------------------------------------------------------
	wf := sp.NewWorkflow("resequencing_workflow", 4)

	// --------------------------------------------------------------------------------
	// Download Reference Genome
	// --------------------------------------------------------------------------------
	dlRef := wf.NewProc("download_ref", "wget -O {o:outfile} "+refBaseUrl+refFileGz)
	dlRef.SetOut("outfile", refFileGz)

	// --------------------------------------------------------------------------------
	// Unzip ref file
	// --------------------------------------------------------------------------------
	ungzipRef := wf.NewProc("ugzip_ref", "gunzip -c {i:in} > {o:out}")
	ungzipRef.In("in").From(dlRef.Out("outfile"))
	ungzipRef.SetOut("out", "{i:in|%.gz}")

	// --------------------------------------------------------------------------------
	// Index Reference Genome
	// --------------------------------------------------------------------------------
	indexRef := wf.NewProc("index_ref", "bwa index -a bwtsw {i:ref}; echo done > {o:done}")
	indexRef.In("ref").From(ungzipRef.Out("out"))
	indexRef.SetOut("done", "{i:ref}.indexed")

	// Create (multi-level) maps where we can gather outports from processes
	// for each for loop iteration and access them in the merge step later
	outs := map[string]map[string]map[string]*sp.OutPort{}
	for _, ind := range individuals {
		outs[ind] = map[string]map[string]*sp.OutPort{}
		for _, spl := range samples {
			outs[ind][spl] = map[string]*sp.OutPort{}
			indSpl := "_" + ind + "_" + spl

			// ------------------------------------------------------------------------
			// Download FastQ component
			// ------------------------------------------------------------------------
			fastqFile := fmt.Sprintf(fastQFilePtn, ind, spl)
			dlFastq := wf.NewProc("download_fastq"+indSpl, "wget -O {o:fastq} "+fastqBaseUrl+fastqFile)
			dlFastq.SetOut("fastq", fastqFile)

			// Save outPorts for later use
			outs[ind][spl]["fastq"] = dlFastq.Out("fastq")

			// ------------------------------------------------------------------------
			// BWA Align
			// ------------------------------------------------------------------------
			align := wf.NewProc("bwa_aln"+indSpl, "bwa aln {i:ref} {i:fastq} > {o:sai} # {i:idxdone}")
			align.In("ref").From(ungzipRef.Out("out"))
			align.In("idxdone").From(indexRef.Out("done"))
			align.In("fastq").From(dlFastq.Out("fastq"))
			align.SetOut("sai", "{i:fastq}.sai")

			// Save outPorts for later use
			outs[ind][spl]["sai"] = align.Out("sai")
		}

		// ---------------------------------------------------------------------------
		// Merge
		// ---------------------------------------------------------------------------
		merge := wf.NewProc("merge_"+ind, "bwa sampe {i:ref} {i:sai1} {i:sai2} {i:fq1} {i:fq2} > {o:merged} # {i:refdone} {p:ind}")
		merge.InParam("ind").FromStr(ind)
		merge.In("ref").From(ungzipRef.Out("out"))
		merge.In("refdone").From(indexRef.Out("done"))
		merge.In("sai1").From(outs[ind]["1"]["sai"])
		merge.In("sai2").From(outs[ind]["2"]["sai"])
		merge.In("fq1").From(outs[ind]["1"]["fastq"])
		merge.In("fq2").From(outs[ind]["2"]["fastq"])
		merge.SetOut("merged", "{p:ind}.merged.sam")
	}
	// -------------------------------------------------------------------------------
	// Run Workflow
	// -------------------------------------------------------------------------------
	wf.Run()
}
