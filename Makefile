BPF_DIR := bpf
BPF_OBJ := $(BPF_DIR)/execve.bpf.o
BPF_HEADER := $(BPF_DIR)/vmlinux.h

# Default target
all: $(BPF_OBJ) build

# Generate vmlinux.h from BTF if it doesn't exist
$(BPF_HEADER):
	@echo "Generating vmlinux.h from BTF..."
	bpftool btf dump file /sys/kernel/btf/vmlinux format c > $@

# Compile BPF program using clang
$(BPF_OBJ): $(BPF_DIR)/execve.bpf.c $(BPF_HEADER)
	@echo "Compiling eBPF program..."
	clang -O2 -g -Wall -target bpf -c $< -o $@

# Build Blooview
build:
	@echo "Building Blooview"
	go build -o bin/blooview main.go

clean:
	rm -f $(BPF_OBJ)
	rm -f bin/blooview
	rm -f $(BPF_HEADER)

.PHONY: all build clean

