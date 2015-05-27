VERSION := "v0.0.1"

clean:
	rm -rf out

build:
	mkdir -p out/darwin out/linux
	GOOS=darwin go build -o out/darwin/c2u c2u.go
	GOOS=linux go build -o out/linux/c2u c2u.go

