.PHONY: data clean help

# Default target
help:
	@echo "Available targets:"
	@echo "  make data  - Generate PNG test data"
	@echo "  make clean - Remove generated test data"
	@echo "  make help  - Show this help message"

# Generate test data
data:
	@echo "Generating PNG test data..."
	@go run datacreator/cmd/main.go
	@echo "Test data generation complete!"

# Clean generated test data
clean:
	@echo "Cleaning test data..."
	@rm -rf testdata/
	@echo "Test data cleaned!"