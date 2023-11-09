# Overview

This Telegram bot application is designed to perform several tasks, including monitoring messages in a channel, checking for plagiarism using a microservice, saving messages to a PostgreSQL database, and sending daily email summaries. The application is written in Go and follows a modular structure with components for the main application, plagiarism checker microservice, Dockerization, and Protocol Buffers for communication.

## Components

### `main.go`
The `main.go` file is the core of the application, responsible for handling Telegram updates, forwarding messages to a specified channel, saving them to the database, and and sending daily email summaries. It utilizes the Telegram Bot API, connects to a PostgreSQL database, and communicates with the plagiarism checker microservice.

### `plagiarism_checker.go`
The `plagiarism_checker.go` file implements the plagiarism checker microservice. It calculates Jaccard similarity coefficients between new and existing texts in the database to identify potential plagiarism. The microservice is accessed by the main application through gRPC.

### `Dockerfile`
The `Dockerfile` enables containerization of the application. It specifies the base image, sets up the working directory, copies necessary files, and builds the Go application.

### `plagiarism_checker.proto`
The `plagiarism_checker.proto` file defines the Protocol Buffers service and messages for communication between the main application and the plagiarism checker microservice. It establishes a standardized interface for gRPC communication.

## Tasks

1. **Message Handling and Forwarding**
   - The bot monitors messages in a specified Telegram channel.
   - Text messages are forwarded to another channel, while media messages with captions are redirected to the same channel.

2. **Database Interaction**
   - Utilizes a PostgreSQL database to store information about forwarded messages.
   - Creates a "messages" table with fields for message text, timestamp, channel username, and message URL.

3. **Plagiarism Checking**
   - Communicates with the plagiarism checker microservice through gRPC.
   - Checks for plagiarism in text messages using Jaccard similarity coefficients.
   - If plagiarism is detected, the original message is deleted from the database.

4. **Daily Email Summaries**
   - Sends daily email summaries of messages posted in the last 24 hours at 4 AM UTC.
   - Emails include links to the original messages in the channel and their corresponding text.

### Configuration
The configuration parameters exists in the `config` package, including Telegram bot token, database connection string, plagiarism checker address, etc.