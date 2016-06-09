package main

import (
	"fmt"
	sp "github.com/scipipe/scipipe"
	plib "github.com/scipipe/scipipe/proclib"
)

const (
	fastq_base_url = "http://bioinfo.perdanauniversity.edu.my/tein4ngs/ngspractice/"
	fastq_file     = "%s.ILLUMINA.low_coverage.4p_%s.fq"
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
	//sp.InitLogDebug()
	gunzipCmdPat := "gunzip -c {i:in} > {o:out}"

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
	gunzipRef := sp.NewFromShell("gunzipRef", gunzipCmdPat)
	pipeRun.AddProcess(gunzipRef)
	gunzipRef.SetPathReplace("in", "out", ".gz", "")
	gunzipRef.In["in"].Connect(dlRefGz.Out["outfile"])

	refFanOut := plib.NewFanOut()
	pipeRun.AddProcess(refFanOut)

	refFanOut.InFile.Connect(gunzipRef.Out["out"])

	// --------------------------------------------------------------------------------
	// Index Reference Genome
	// --------------------------------------------------------------------------------
	idxRef := sp.NewFromShell("Index Ref",
		"bwa index -a bwtsw {i:index}; echo done > {o:done}")
	idxRef.SetPathExtend("index", "done", ".indexed")
	pipeRun.AddProcess(idxRef)
	idxRef.In["index"].Connect(refFanOut.GetOutPort("index_ref"))

	idxRefDoneFanOut := plib.NewFanOut()
	pipeRun.AddProcess(idxRefDoneFanOut)
	idxRefDoneFanOut.InFile.Connect(idxRef.Out["done"])

	outPorts := make(map[string]map[string]map[string]*sp.OutPort)
	for _, individual := range individuals {
		outPorts[individual] = make(map[string]map[string]*sp.OutPort)
		for _, sample := range samples {
			outPorts[individual][sample] = make(map[string]*sp.OutPort)

			// --------------------------------------------------------------------------------
			// Download FastQ component
			// --------------------------------------------------------------------------------
			file_name := fmt.Sprintf(fastq_file, individual, sample)
			dlFastq := sp.NewFromShell("dl_fastq",
				"wget -O {o:fastq} "+fastq_base_url+file_name)
			dlFastq.SetPathStatic("fastq", file_name)
			pipeRun.AddProcess(dlFastq)

			fastQFanOut := plib.NewFanOut()
			fastQFanOut.InFile.Connect(dlFastq.Out["fastq"])
			pipeRun.AddProcess(fastQFanOut)

			outPorts[individual][sample]["fastq"] = sp.NewOutPort()
			outPorts[individual][sample]["fastq"] = fastQFanOut.GetOutPort("merg")

			// --------------------------------------------------------------------------------
			// BWA Align
			// --------------------------------------------------------------------------------
			bwaAln := sp.NewFromShell("bwa_aln",
				"bwa aln {i:ref} {i:fastq} > {o:sai} # {i:idxdone}")
			pipeRun.AddProcess(bwaAln)
			bwaAln.SetPathExtend("fastq", "sai", ".sai")
			// Connect
			bwaAln.In["ref"].Connect(refFanOut.GetOutPort("bwa_aln_" + individual + "_" + sample))
			bwaAln.In["idxdone"].Connect(idxRefDoneFanOut.GetOutPort("bwa_aln_" + individual + "_" + sample))
			bwaAln.In["fastq"].Connect(fastQFanOut.GetOutPort("bwa_aln"))
			// Store in map
			outPorts[individual][sample]["sai"] = sp.NewOutPort()
			outPorts[individual][sample]["sai"] = bwaAln.Out["sai"]
		}

		// --------------------------------------------------------------------------------
		// Merge
		// --------------------------------------------------------------------------------
		individualParamGen := plib.NewStringGenerator(individual)
		pipeRun.AddProcess(individualParamGen)

		merg := sp.NewFromShell("merge_"+individual,
			"bwa sampe {i:ref} {i:sai1} {i:sai2} {i:fq1} {i:fq2} > {o:merged} # {i:refdone} {p:individual}")
		pipeRun.AddProcess(merg)
		merg.PathFormatters["merged"] = func(t *sp.SciTask) string {
			return fmt.Sprintf("%s.merged.sam", t.Params["individual"])
		}

		merg.In["ref"].Connect(refFanOut.GetOutPort("merg_" + individual))
		merg.In["refdone"].Connect(idxRefDoneFanOut.GetOutPort("merg_" + individual))
		merg.In["sai1"].Connect(outPorts[individual]["1"]["sai"])
		merg.In["sai2"].Connect(outPorts[individual]["2"]["sai"])
		merg.In["fq1"].Connect(outPorts[individual]["1"]["fastq"])
		merg.In["fq2"].Connect(outPorts[individual]["2"]["fastq"])
		merg.ParamPorts["individual"].Connect(individualParamGen.Out)

		sink.Connect(merg.Out["merged"])
	}
	// --------------------------------------------------------------------------------
	// Run pipeline
	// --------------------------------------------------------------------------------
	pipeRun.AddProcess(sink)
	pipeRun.Run()
}
