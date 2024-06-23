PY=$(CURDIR)/.venv/bin/python3
PIP=$(CURDIR)/.venv/bin/pip

1brc: $(CURDIR)/cmd/1brc
	go build -o "$(CURDIR)/1brc" "$(CURDIR)/cmd/1brc" 

.PHONY: profile
profile: 1brc 1brc.prof
	go tool pprof 1brc 1brc.prof

1brc.prof: measurements.txt 1brc
	cat measurements.txt | "$(CURDIR)/1brc" --profile="$(CURDIR)/1brc.prof"

test: measurements.txt 1brc 
	cat measurements.txt | "$(CURDIR)/1brc"

.venv:
	python3 -m venv "$(CURDIR)/.venv"

.PHONY: pip
pip: .venv
	$(PIP) install --upgrade pip

.PHONY: numpy
numpy: pip
	$(PIP) install numpy

.PHONY: tqdm
tqdm: pip
	$(PIP) install tqdm

.PHONY: polars
polars: pip
	$(PIP) install polars

measurements.txt: | numpy tqdm polars
	$(PY) "$(CURDIR)/scripts/create_measurements.py"

.PHONY: clean
clean:
	rm -rf "$(CURDIR)/.venv"
	rm -f "$(CURDIR)/measurements.txt"
	rm -f "$(CURDIR)/1brc"
	rm -f "$(CURDIR)/1brc.prof"

