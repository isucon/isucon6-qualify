isucon6q: isucon6q/isucon6q-bench-worker isucon6q/isucon6q-bench

clean:
	rm isucon6q/isucon6q-bench-worker isucon6q/isucon6q-bench

isucon6q/isucon6q-bench-worker:
	GOOS=$(GOOS) go build -ldflags="-s -w" -o $@ ./worker

isucon6q/isucon6q-bench:
	GOOS=$(GOOS) go build -ldflags="-s -w" -o $@

.PHONY: isucon6q clean
