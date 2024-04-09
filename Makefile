PACKAGE_NAME = depositter
VERSION = 1.0
DESTDIR = ./bin/

.PHONY: clean

all: go-build package clean

go-build:
	go build -o ./bin/depositter

package:
	mkdir -p $(DESTDIR)/usr/local/bin
	mkdir -p $(DESTDIR)/DEBIAN
	cp $(DESTDIR)/depositter $(DESTDIR)/usr/local/bin
	cp control $(DESTDIR)/DEBIAN
	dpkg-deb --build $(DESTDIR) $(PACKAGE_NAME)_$(VERSION)_amd64.deb

clean:
	rm -rf bin/*
