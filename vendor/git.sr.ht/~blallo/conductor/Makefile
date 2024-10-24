TESTOUT ?= ./testout
BROWSER ?= firefox
REPETITIONS ?= 100

.PHONY: test
test:
	@go test -count=1 -race ./...

.PHONY: test-rep
test-rep:
	@for i in $$(seq 1 $(REPETITIONS)); do \
		echo "---------------------------------------------------"; \
		echo "Test run $${i}"; \
		$(MAKE) test; \
		if [ $$? -ne 0 ]; then exit $?; fi; \
	done

$(TESTOUT):
	mkdir $(TESTOUT)

.PHONY: test-cov
test-cov: $(TESTOUT)
	go test -v -coverprofile=$(TESTOUT)/cov.out ./...
	go tool cover -html=$(TESTOUT)/cov.out -o $(TESTOUT)/cov.html
	$(BROWSER) $(TESTOUT)/cov.html

.PHONY: clean
clean:
	rm -rf $(TESTOUT)

.PHONY: lint
lint:
	@gofmt -d -e .
