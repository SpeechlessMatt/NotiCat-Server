all: submods build

submods:
	$(MAKE) -C mail
	$(MAKE) -C scripts install

build:
	go build -o noticat .

clean:
	$(MAKE) -C mail clean
	$(MAKE) -C scripts clean
	rm -f noticat

.PHONY: all submods build clean
