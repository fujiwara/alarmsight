export GOOS=linux
export GOARCH=amd64
export CGO_ENABLED=0

help:
	@echo "make bootstrap"

bootstrap:
	cd ../ && make && mv alarmsight bootstrap
	mv ../bootstrap .

deploy: bootstrap
	lambroll deploy

clean:
	rm -f bootstrap
	cd ../ && make clean
