package main

import (
	"context"
	"testing"

	pb "runo/plagiarism_checker"
)

func TestCheckPlagiarism(t *testing.T) {
	server := &plagiarismCheckerServer{}

	request := &pb.CheckPlagiarismRequest{
		MessageText: "Test message",
	}

	response, err := server.CheckPlagiarism(context.Background(), request)

	if err != nil {
		t.Fatalf("Error calling CheckPlagiarism: %v", err)
	}

	if response.IsPlagiarized {
		t.Errorf("Expected IsPlagiarized to be false, got true")
	}
}
