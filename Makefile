YTDLP_BIN := bin/yt-dlp
YTDLP_URL_MACOS := https://github.com/yt-dlp/yt-dlp/releases/latest/download/yt-dlp_macos
YTDLP_URL_LINUX := https://github.com/yt-dlp/yt-dlp/releases/latest/download/yt-dlp_linux
YTDLP_URL_LINUX_ARM := https://github.com/yt-dlp/yt-dlp/releases/latest/download/yt-dlp_linux_aarch64

all: clean yt-dlp build

clean:
	# rm -rf bin/*
	rm -rf downloads/*
	find . -name ".DS_Store" -type f -delete
	@echo "[CLEAN] bin/* и .DS_Store успешно удалены."

build:
	go build -o bin/bot ./cmd/app

yt-dlp:
	@echo "Detecting platform..."
	@if [ ! -d bin ]; then mkdir -p bin; fi
	@OS=$$(uname -s); ARCH=$$(uname -m); \
	if [ "$$OS" = "Darwin" ] && [ "$$ARCH" = "arm64" ]; then \
		echo "Installing yt-dlp for macOS M1..."; \
		curl -Lo $(YTDLP_BIN) $(YTDLP_URL_MACOS); \
	elif [ "$$OS" = "Linux" ] && [ "$$ARCH" = "x86_64" ]; then \
		echo "Installing yt-dlp for Linux x86_64..."; \
		curl -Lo $(YTDLP_BIN) $(YTDLP_URL_LINUX); \
	elif [ "$$OS" = "Linux" ] && [ "$$ARCH" = "aarch64" ]; then \
		echo "Installing yt-dlp for Linux ARM64 (aarch64)..."; \
		curl -Lo $(YTDLP_BIN) $(YTDLP_URL_LINUX_ARM); \
	else \
		echo "Unsupported platform: $$OS $$ARCH"; \
		exit 1; \
	fi
	@chmod +x $(YTDLP_BIN)