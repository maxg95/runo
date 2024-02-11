# Tasks and Components

## 1. Message Handling and Forwarding
The bot monitors messages in a specified Telegram channel and then forwarded them another channel, (text and media messages with captions).

## 2. Database Interaction
Utilizing a PostgreSQL database, the app stores essential information about forwarded messages. The database schema includes a "messages" table with fields such as message text, timestamp, channel username, and message URL.

## 3. Plagiarism Checking
This component communicates with a plagiarism checker microservice through gRPC. It employs Jaccard similarity coefficients to detect plagiarism in text messages. If plagiarism is identified, the original message is promptly deleted from the database.

## 4. Daily Email Summaries
The app sends daily email summaries of messages posted in the last 24 hours. This automated email is scheduled to be dispatched at 4 AM UTC and includes links to the original messages in the channel along with their corresponding text.

## 5. Testing
The testing process is facilitated through the inclusion of `plagiarism_checker_test.go` and `main_test.go`. These files are specifically designed for testing the plagiarism checker and mail sending functions, ensuring the robustness of the implemented features.

## 6. Redis
The app employs Redis for multiple purposes, acting as a message broker, cache, and NoSQL database. This versatile usage enhances the system's efficiency and responsiveness.

## 7. Web Interface
To interact with the app, users can utilize a secure web interface. This includes HTTPS (TLS) for encryption, a REST API for seamless communication, and an HTML template for a user-friendly experience.

## 8. Docker
For easy deployment and management, the app is containerized using Docker. Three Docker images and containers are utilized - one each for Redis, PostgreSQL, and the application itself. The Dockerfile and docker-compose configuration ensure a streamlined deployment process.

## 9. Kubernetes
The app's deployment and orchestration are handled through Kubernetes. Deployment and service files are provided for Redis, PostgreSQL, and the main application, simplifying scalability and management.

## 10. AWS
For cloud deployment, the app is integrated with Amazon Elastic Kubernetes Service (EKS). This integration facilitates seamless scaling, efficient resource utilization, and robust performance in an AWS environment.


### Configuration
Configuration parameters, including the Telegram bot token, database connection string, plagiarism checker address, etc., are centralized in the `config` package for easy access and modification.

### Deployment in Videos

- **Docker:**
  https://youtu.be/JWK47EQJLQM

- **Kubernetes:**
  https://youtu.be/n2quIJnFLlk

- **AWS:**
  https://youtu.be/32rkJt0_cI4
  https://youtu.be/-942zhizHug
