NAME        ?= GUIEtcdClient.exe
OUTPUT_BIN  ?= .\bin\${NAME}
SOURCE      ?= .\cmd
GO_FLAGS    ?=
GO_TAGS     ?= walk_use_cgo
CGO_ENABLED ?= 1
#VERSION     ?= $(shell for /f "tokens=1,2,3 delims=." %%a in (VERSION) do set /a var1=%%a && set /a var2=%%b && set /a var3=%%c+1 && <NUL set /p = "%var1%.%var2%.%var3%")
#VERSION     ?= $(shell powershell.exe -Command "$file = Get-Content .\VERSION -Raw ; $parts = $file.Split('.', 3) ; [int]$buildnum = $parts[2] ; $buildnum++ ; $parts[2] = [string]$buildnum ; $parts[0],$parts[1],$parts[2] -join '.'")
VERSION     ?= $(shell git rev-list --count HEAD)-$(shell git rev-parse --short=7 HEAD)
LD_FLAGS    ?= "-w -s -H=windowsgui -X main.version=${VERSION}"

default: move build

move:
	@echo "Moving output..."
	@IF EXIST ${OUTPUT_BIN} ( @move /y ${OUTPUT_BIN} ${OUTPUT_BIN}.bak )

build:
	@echo "Building from source..."
	@set CGO_ENABLED=${CGO_ENABLED}
	go build -v -tags ${GO_TAGS} ${GO_FLAGS} -ldflags=${LD_FLAGS} -o ${OUTPUT_BIN} ${SOURCE}

#versioninc:
#	@echo "Incremention version file..."
#	@echo ${VERSION} > VERSION
