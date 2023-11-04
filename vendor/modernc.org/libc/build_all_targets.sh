set -e
for tag in none libc.dmesg libc.membrk libc.memgrind libc.strace
do
	echo "-tags=$tag"
	echo "GOOS=darwin GOARCH=amd64"
	GOOS=darwin GOARCH=amd64 go build -tags=$tag -v ./...
	GOOS=darwin GOARCH=amd64 go test -tags=$tag -c -o /dev/null
	echo "GOOS=darwin GOARCH=arm64"
	GOOS=darwin GOARCH=arm64 go build -tags=$tag -v ./...
	GOOS=darwin GOARCH=arm64 go test -tags=$tag -c -o /dev/null
	echo "GOOS=freebsd GOARCH=386"
	GOOS=freebsd GOARCH=386 go build -tags=$tag -v ./...
	GOOS=freebsd GOARCH=386 go test -tags=$tag -c -o /dev/null
	echo "GOOS=freebsd GOARCH=amd64"
	GOOS=freebsd GOARCH=amd64 go build -tags=$tag -v ./...
	GOOS=freebsd GOARCH=amd64 go test -tags=$tag -c -o /dev/null
	echo "GOOS=freebsd GOARCH=arm"
	GOOS=freebsd GOARCH=arm go build -tags=$tag -v ./...
	GOOS=freebsd GOARCH=arm go test -tags=$tag -c -o /dev/null
	echo "GOOS=linux GOARCH=386"
	GOOS=linux GOARCH=386 go build -tags=$tag -v ./...
	GOOS=linux GOARCH=386 go test -tags=$tag -c -o /dev/null
	echo "GOOS=linux GOARCH=amd64"
	GOOS=linux GOARCH=amd64 go build -tags=$tag -v ./...
	GOOS=linux GOARCH=amd64 go test -tags=$tag -c -o /dev/null
	echo "GOOS=linux GOARCH=arm"
	GOOS=linux GOARCH=arm go build -tags=$tag -v ./...
	GOOS=linux GOARCH=arm go test -tags=$tag -c -o /dev/null
	echo "GOOS=linux GOARCH=arm64"
	GOOS=linux GOARCH=arm64 go build -tags=$tag -v ./...
	GOOS=linux GOARCH=arm64 go test -tags=$tag -c -o /dev/null
	echo "GOOS=linux GOARCH=loong64"
	GOOS=linux GOARCH=loong64 go build -tags=$tag -v ./...
	GOOS=linux GOARCH=loong64 go test -tags=$tag -c -o /dev/null
	echo "GOOS=linux GOARCH=ppc64le"
	GOOS=linux GOARCH=ppc64le go build -tags=$tag -v ./...
	GOOS=linux GOARCH=ppc64le go test -tags=$tag -c -o /dev/null
	echo "GOOS=linux GOARCH=riscv64"
	GOOS=linux GOARCH=riscv64 go build -tags=$tag -v ./...
	GOOS=linux GOARCH=riscv64 go test -tags=$tag -c -o /dev/null
	echo "GOOS=linux GOARCH=s390x"
	GOOS=linux GOARCH=s390x go build -tags=$tag -v ./...
	GOOS=linux GOARCH=s390x go test -tags=$tag -c -o /dev/null
	echo "GOOS=netbsd GOARCH=amd64"
	GOOS=netbsd GOARCH=amd64 go build -tags=$tag -v ./...
	GOOS=netbsd GOARCH=amd64 go test -tags=$tag -c -o /dev/null
	echo "GOOS=netbsd GOARCH=arm"
	GOOS=netbsd GOARCH=arm go build -tags=$tag -v ./...
	GOOS=netbsd GOARCH=arm go test -tags=$tag -c -o /dev/null
	echo "GOOS=openbsd GOARCH=386"
	GOOS=openbsd GOARCH=386 go build -tags=$tag -v ./...
	GOOS=openbsd GOARCH=386 go test -tags=$tag -c -o /dev/null
	echo "GOOS=openbsd GOARCH=amd64"
	GOOS=openbsd GOARCH=amd64 go build -tags=$tag -v ./...
	GOOS=openbsd GOARCH=amd64 go test -tags=$tag -c -o /dev/null
	echo "GOOS=openbsd GOARCH=arm64"
	GOOS=openbsd GOARCH=arm64 go build -tags=$tag -v ./...
	GOOS=openbsd GOARCH=arm64 go test -tags=$tag -c -o /dev/null
	echo "GOOS=windows GOARCH=386"
	GOOS=windows GOARCH=386 go build -tags=$tag -v ./...
	GOOS=windows GOARCH=386 go test -tags=$tag -c -o /dev/null
	echo "GOOS=windows GOARCH=amd64"
	GOOS=windows GOARCH=amd64 go build -tags=$tag -v ./...
	GOOS=windows GOARCH=amd64 go test -tags=$tag -c -o /dev/null
	echo "GOOS=windows GOARCH=arm64"
	GOOS=windows GOARCH=arm64 go build -tags=$tag -v ./...
	GOOS=windows GOARCH=arm64 go test -tags=$tag -c -o /dev/null
done
