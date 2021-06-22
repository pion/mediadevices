.PHONY: darwin-static
all: clean darwin windows clean
darwin: clean darwin-basic darwin-strip stat
windows: clean windows-static windows-static-strip stat

clean: 
	rm -rf ./bin

stat: 
	ls -hl ./bin

darwin-basic:
	GOENV="/Users/dbl0null/Library/Application Support/go/env" \
	GOOS=darwin GOARCH=amd64 \
	CGO_ENABLED=1 \
	go build -x -v \
	-o ./bin/streamer.bin \
	&& objdump -a -g -h ./bin/streamer.bin

darwin-strip:
	GOENV="/Users/dbl0null/Library/Application Support/go/env" \
	GOOS=darwin GOARCH=amd64 \
	CGO_ENABLED=1 \
	go build -x -v \
	--ldflags '-s -w' \
	-o ./bin/streamer.strip.bin \
	&& objdump -a -g -h ./bin/streamer.strip.bin

darwin-static:
	GOENV="/Users/dbl0null/Library/Application Support/go/env" \
	GOOS=darwin GOARCH=amd64 \
	CGO_ENABLED=1 \
	go build -x -v \
	--ldflags '--extldflags "-static -static-libgcc -static-libstdc++"' \
	-o ./bin/streamer.static.bin \
	&& objdump -a -g -h ./bin/streamer.static.bin

darwin-static-strip:
	GOENV="/Users/dbl0null/Library/Application Support/go/env" \
	GOOS=darwin GOARCH=amd64 \
	CGO_ENABLED=1 \
	go build -x -v \
	--ldflags '--extldflags "-static -static-libgcc -static-libstdc++" -s -w' \
	-o ./bin/streamer.static.strip.bin \
	&& objdump -p -f -a -g --section-headers ./bin/streamer.static.strip.bin

windows-static:
	GOOS=windows GOARCH=amd64 \
	CGO_ENABLED=1 \
	CXX="x86_64-w64-mingw32-g++" \
	CXX_FOR_TARGET="x86_64-w64-mingw32-g++" \
	CC="x86_64-w64-mingw32-gcc" \
	CC_FOR_TARGET="x86_64-w64-mingw32-gcc" \
	go build -x -v \
	--ldflags '--extldflags "-static -static-libgcc -static-libstdc++"' \
	-o ./bin/streamer.static.exe \
	&& objdump -p -f -a -g -h ./bin/streamer.static.exe

windows-static-strip:
	GOOS=windows GOARCH=amd64 \
	CGO_ENABLED=1 \
	CXX="x86_64-w64-mingw32-g++" \
	CXX_FOR_TARGET="x86_64-w64-mingw32-g++" \
	CC="x86_64-w64-mingw32-gcc" \
	CC_FOR_TARGET="x86_64-w64-mingw32-gcc" \
	go build -x -v \
	--ldflags '--extldflags "-static -static-libgcc -static-libstdc++" -s -w' \
	-o ./bin/streamer.static.strip.exe
	objdump -p -f -a -g --section-headers ./bin/streamer.static.strip.exe

