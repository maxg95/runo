.PHONY: checker main

checker:
	@cd plagiarism_checker_main && go run plagiarism_checker.go
	
main:
	@go run main.go