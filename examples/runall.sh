for d in $(ls -d * | grep -v "\."); do echo $d; cd $d; go run *.go; cd ..; done
