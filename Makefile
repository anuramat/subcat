.SILENT:
.PHONY: hack demo clean

hack:
	make -L --no-print-directory demo

demo: result
	slow() { echo [$$2]; for i in $$(seq 1 $$1); do echo output $$i; sleep 1; done; }; \
		./result/bin/subcat <(slow 5 "slower") <(slow 3 "faster")

SOURCES = $(shell git ls-files)
result: $(SOURCES)
	nix build

clean:
	rm result
