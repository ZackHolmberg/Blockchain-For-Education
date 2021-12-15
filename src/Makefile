GO=go
BUILD=$(GO) build
TEST=$(GO) test
CLEAN=$(GO) clean
    
all: test build
build:
	@echo "\nBuilding executables...\n"
	@cd blockchain && $(BUILD) .
	@cd peer && $(BUILD) -o peerBinary
	@cd middleware && $(BUILD) -o middlewareBinary
test: 
	@$(TEST) -v
clean:
	@$(CLEAN)
runPeer:
	@echo "\nBuilding peer executable...\n"
	@cd blockchain && $(BUILD) .
	@cd peer && $(BUILD) -o peerBinary
	@echo "\nRunning peer executable...\n"
	@cd peer && ./peerBinary
runMiddleware:
	@echo "\nBuilding middleware executable...\n"
	@cd blockchain && $(BUILD) .
	@cd middleware && $(BUILD) -o middlewareBinary
	@echo "\nRunning middleware executable...\n"
	@cd middleware && ./middlewareBinary

