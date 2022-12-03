all: samler

libsml:
	make -C libsml/sml
	make -C libsml/examples

samler: libsml
	go build
	go test

.PHONY: clean libsml
clean:
	rm -f *.o
	rm -f samler
	make -C libsml clean
