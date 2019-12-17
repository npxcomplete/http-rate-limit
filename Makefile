
FUNC := churn_without_gen

test:
	go test ./...

cover:
	mkdir -p ./target
	go tool cover -html=target/coverage.out -o target/coverage.html

profile:
	mkdir -p ./target
	go test ./src/ \
		-bench=$(FUNC) \
		-benchmem \
		-benchtime=10s\
		-memprofile target/memprofile.out \
		-cpuprofile target/cpuprofile.out

flame:
	go tool pprof -http=":8081" src.test target/cpuprofile.out
