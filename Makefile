MOUDLE_NAME = asrp
TARGET_DIR = ./bin
SRC_DIR = ./cmd

SUFFIX = 

ifeq ($(OS),Windows_NT)
    RM = del /q
    MKDIR = if not exist $(TARGET_DIR) mkdir $(TARGET_DIR)
	SUFFIX=.exe
else
    RM = rm -f
    MKDIR = mkdir -p $(TARGET_DIR)
endif

build:
	$(MKDIR)
	go build -o "$(TARGET_DIR)/$(MOUDLE_NAME)$(SUFFIX)" "$(SRC_DIR)
clean:
	$(RM) "$(TARGET_DIR)/$(BINARY_NAME).exe"

run:
	.\$(TARGET_DIR)\$(BINARY_NAME)$(SUFFIX)

.PHONY: build clean run test