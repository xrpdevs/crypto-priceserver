ifndef $(GOPATH)
    GOPATH=$(shell go env GOPATH)
    export GOPATH
endif

.ONESHELL:
default:
	echo "Installing dependencies"; \
	go get github.com/mattn/go-sqlite3; \
	go get gopkg.in/yaml.v2; \
	echo "Building priceserver"; \
	go build priceserver.go; \
	echo 'All done! Now try "make run" or "make install"'


run:
	go run priceserver.go --config=./priceserver.yml --dbfile=./db/data.db

.ONESHELL:
install:
	sudo su
	echo "Creating /var/lib/priceserver/ for sqlite3 database"; \
	mkdir /var/lib/priceserver/ ; \
	chown nobody:nogroup /var/lib/priceserver -R
	echo "Copying priceserver to /use/local/bin/" ; \
	cp ./priceserver.yml /etc/ ; \
	cp ./priceserver /usr/local/bin/priceserver ; \
	cp ./priceserver.service /etc/systemd/system/ ;\
	systemctl daemon-reload
	echo "All done!"
	exit
