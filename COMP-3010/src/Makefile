GO=go
BUILD=$(GO) build
TEST=$(GO) test
CLEAN=$(GO) clean
    
all: test build
build:
	@echo "\nBuilding executable...\n"
	@cd blockchain
	@$(BUILD) .
	@cd ..
	@$(BUILD) .
test: 
	@$(TEST) -v
clean:
	@$(CLEAN)
run:
	@echo "\nBuilding executable...\n"
	@cd blockchain
	@$(BUILD) .
	@cd ..
	@$(BUILD) .
	@echo "\nRunning executable...\n"
	@./src
	@$(CLEAN)
	@rm -f $(BINARY)