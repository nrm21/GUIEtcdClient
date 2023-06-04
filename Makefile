NAME        ?= GUIEtcdClient.exe
OUTPUT_BIN  ?= .\bin\${NAME}
SOURCE      ?= .\src
GO_FLAGS    ?=
GO_TAGS     ?= walk_use_cgo
CGO_ENABLED ?= 1
#VERSION     := $(shell for /f "tokens=1,2,3 delims=." %%a in (VERSION) do set var1=%%a && set var2=%%b && set /a var3=%%c+1 && %var1%.%var2%.%var3% )
#LD_FLAGS    ?= "-w -s -H=windowsgui -X main.version=${VERSION}"
LD_FLAGS    ?= "-w -s -H=windowsgui -X main.version=1.1.9"


default:
	@move /y ${OUTPUT_BIN} ${OUTPUT_BIN}.bak
	@make build

build:
	@set CGO_ENABLED=${CGO_ENABLED}
	go build -tags ${GO_TAGS} ${GO_FLAGS} -ldflags=${LD_FLAGS} -o ${OUTPUT_BIN} ${SOURCE}
