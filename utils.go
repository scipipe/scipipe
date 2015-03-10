package scipipe

func check(e error, id string) {
	if e != nil {
		panic(e)
	}
}
