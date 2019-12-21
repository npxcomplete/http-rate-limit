
FUNC := churn_without_gen

test:
	go test ./...

cover:
	mkdir -p ./target
	go tool cover -html=target/coverage.out -o target/coverage.html

bench:
	mkdir -p ./target
	go test ./src/ \
    	-bench=. \
		-benchtime=1s\
		-cpuprofile ./target/cpuprofile.out

profile:
	mkdir -p ./target
	go test ./src/ \
		-bench=$(FUNC) \
		-benchmem \
		-benchtime=1s\
		-memprofile ./target/memprofile.out \
		-cpuprofile ./target/cpuprofile.out \
		-o ./target/pkg.test

flame:
	go tool pprof -http=":8081" target/pkg.test target/cpuprofile.out

escape_analysis:
	go tool pprof -alloc_space target/pkg.test target/memprofile.out

gcflags:
	go build -gcflags "-m -m" ./... 2>&1 | less -S