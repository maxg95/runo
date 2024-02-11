package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net"
	"strings"

	pb "runo/plagiarism_checker"

	"google.golang.org/grpc"

	_ "github.com/lib/pq"

	"runo/config"

	"github.com/go-redis/redis/v8"
)

const similarityThreshold = 0.3

type plagiarismCheckerServer struct {
	pb.UnimplementedPlagiarismCheckerServer
}

func (s *plagiarismCheckerServer) CheckPlagiarism(ctx context.Context, req *pb.CheckPlagiarismRequest) (*pb.CheckPlagiarismResponse, error) {
	redisClient := redis.NewClient(&redis.Options{
		Addr:     "172.17.0.3:6379",
		Password: "",
		DB:       0,
	})
	defer redisClient.Close()

	pubsub := redisClient.Subscribe(context.Background(), "plagiarism_channel")
	defer pubsub.Close()

	msg, err := pubsub.ReceiveMessage(context.Background())
	if err != nil {
		log.Println("Error receiving message from plagiarism_channel:", err)
	} else {
		log.Printf("Received message from plagiarism_channel: %s", msg.Payload)
	}

	db, err := sql.Open("postgres", config.DBConnStr)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	rows, err := db.Query("SELECT message_text FROM messages")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	existingTexts := make([]string, 0)
	for rows.Next() {
		var existingText string
		err := rows.Scan(&existingText)
		if err != nil {
			return nil, err
		}
		log.Printf("Processing message: %s", existingText)
		existingTexts = append(existingTexts, existingText)
	}

	newText := strings.Fields(req.GetMessageText())
	fmt.Println("newText:", newText)
	for _, existingText := range existingTexts {
		if existingText == req.GetMessageText() {
			continue
		}
		existingWords := strings.Fields(existingText)
		fmt.Println("existingWords:", existingWords)
		similarity := jaccardSimilarity(newText, existingWords)
		fmt.Println("Similarity:", similarity)
		if similarity > similarityThreshold {
			return &pb.CheckPlagiarismResponse{
				IsPlagiarized: true,
			}, nil
		}
	}

	return &pb.CheckPlagiarismResponse{
		IsPlagiarized: false,
	}, nil
}

func jaccardSimilarity(set1, set2 []string) float64 {
	fmt.Println("Set1:", set1)
	fmt.Println("Set2:", set2)

	intersection := make(map[string]bool)

	// Calculate the intersection between set1 and set2
	for _, s := range set1 {
		for _, t := range set2 {
			if s == t {
				intersection[s] = true
				break
			}
		}
	}

	intersectionCount := float64(len(intersection))
	unionCount := float64(len(set1) + len(set2) - len(intersection))

	if unionCount == 0 {
		return 0.0
	}

	// Calculate the Jaccard similarity coefficient
	similarity := intersectionCount / unionCount

	return similarity
}

func main() {
	lis, err := net.Listen("tcp", config.PlagiarismCheckerAddress)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterPlagiarismCheckerServer(grpcServer, &plagiarismCheckerServer{})

	fmt.Printf("Plagiarism checker listening on %s\n", config.PlagiarismCheckerAddress)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
