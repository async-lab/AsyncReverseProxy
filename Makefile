MOUDLE_NAME = asrp
TARGET_DIR = ./bin
SRC_DIR = ./cmd
SUFFIX = 
TARGET = $(TARGET_DIR)/$(MOUDLE_NAME)$(SUFFIX)

ifeq ($(OS),Windows_NT)
	set CGO_ENABLED=0
    RM = del /q
    MKDIR = if not exist $(TARGET_DIR) mkdir $(TARGET_DIR)
	SUFFIX=.exe
else
	export CGO_ENABLED=0
    RM = rm -f
    MKDIR = mkdir -p $(TARGET_DIR)
endif

build:
	$(MKDIR)
	go build -ldflags="-s -w" -o "$(TARGET)" "$(SRC_DIR)"
	upx --best --lzma "$(TARGET)"
clean:
	$(RM) "$(TARGET)"

run:
	".\$(TARGET)"

.PHONY: build clean run test