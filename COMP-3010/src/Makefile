GO=go
BUILD=$(GO) build
TEST=$(GO) test
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

