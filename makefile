build:
	go build -o output/server ./

run: build
	./output/server

watch:
	ulimit -n 1000 #increase the file watch limit, might required on MacOS
	./bin/reflex -s -r '\.go$$' make run