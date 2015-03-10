package scipipe

func check(e error) {
	if e != nil {
		panic(e)
	}
}
