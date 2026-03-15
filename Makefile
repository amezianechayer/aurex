test-driver-postgres:
	go test -v -coverprofile=coverage.out -coverpkg=./... ./... -storage-driver 'postgres'

test:
	go test -v -coverprofile=coverage.out -coverpkg=./... ./...

bench:
	test:
	go test -v -coverprofile=coverage.out -coverpkg=./... ./...

bench:
	go test -test.bench=. -run=^a ./...

fetch-horizon:
	cd cmd/horizon && \
	curl -OL https://github.com/amezianechayer/horizon/releases/latest/download/horizon.tar.gz && \
	tar -xvf horizon.tar.gz && \
	rm horizon.tar.gz