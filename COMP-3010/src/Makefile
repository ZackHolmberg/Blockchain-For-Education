GO=go
BUILD=$(GO) build
TEST=$(GO) test
CLEAN=$(GO) clean
BINARY=executable
    
all: test build
build:
	@echo "\nBuilding executable...\n"
	@cd blockchain
	@$(BUILD) .
	@cd ..
	@$(BUILD) -o $(BINARY) -v
test: 
	@$(TEST) -v
clean:
	@$(CLEAN)
	@rm -f $(BINARY)
run:
	@echo "\nBuilding executable...\n"
	@cd blockchain
	@$(BUILD) .
	@cd ..
	@$(BUILD) -o $(BINARY) -v
	@echo "\nRunning executable...\n"
	@./$(BINARY)
	@$(CLEAN)
	@rm -f $(BINARY)