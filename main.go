package main
/
import (
	"context"
	"crypto/tls"
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/smtp"
	"sync"
	"time"

	"runo/plagiarism_checker"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"google.golang.org/grpc"

	_ "github.com/lib/pq"

	"runo/config"

	"github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"
)

type Message struct {
	ID              int
	MessageText     string
	Timestamp       time.Time
	ChannelUsername string
	MessageURL      string
}

func (m Message) TimestampFormat() string {
	return m.Timestamp.Format("02.01.2006 в 15:04:05")
}

func getMessagesHandler(w http.ResponseWriter, r *http.Request) {
	db, err := sql.Open("postgres", config.DBConnStr)
	if err != nil {
		log.Println("Error connecting to the database:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	rows, err := db.Query("SELECT id, message_text, timestamp, channel_username, message_url FROM messages")
	if err != nil {
		log.Println("Error fetching messages from the database:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var messages []Message
	for rows.Next() {
		var message Message
		err := rows.Scan(&message.ID, &message.MessageText, &message.Timestamp, &message.ChannelUsername, &message.MessageURL)
		if err != nil {
			log.Println("Error scanning rows:", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		messages = append(messages, message)
	}

	tmpl, err := template.ParseFiles("template.html")
	if err != nil {
		log.Println("Error parsing template:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	err = tmpl.Execute(w, messages)
	if err != nil {
		log.Println("Error executing template:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

// Sending email task Part 1.
func sendEmail(subject, body string) error {
	from := config.MailFrom
	password := config.MailPassword
	to := config.MailTo

	log.Println("Preparing to send email")

	msg := "From: " + from + "\n" +
		"To: " + to + "\n" +
		"Subject: " + subject + "\n\n" +
		body

	auth := smtp.PlainAuth("", from, password, "smtp.gmail.com")

	err := smtp.SendMail("smtp.gmail.com:587", auth, from, []string{to}, []byte(msg))
	if err != nil {
		log.Println("Error sending email:", err)
		return err
	}
	return nil
}

func main() {
	// Create a new Telegram bot instance.
	bot, err := tgbotapi.NewBotAPI(config.BotToken)
	if err != nil {
		log.Fatal("Error creating bot:", err)
	}

	// Set the bot to get updates.
	bot.Debug = false
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)
	if err != nil {
		log.Fatal("Error getting updates:", err)
	}

	// Connect to the PostgreSQL database.
	db, err := sql.Open("postgres", config.DBConnStr)
	if err != nil {
		log.Fatal("Error connecting to the database:", err)
	}
	defer db.Close()

	// Create the "messages" table.
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS messages (
		id SERIAL PRIMARY KEY,
		message_text TEXT NOT NULL,
		timestamp TIMESTAMP NOT NULL,
		channel_username TEXT NOT NULL,
		message_url TEXT NOT NULL
	)`)
	if err != nil {
		log.Fatal("Error creating table:", err)
	}

	redisClient := redis.NewClient(&redis.Options{
		Addr:     "172.17.0.3:6379",
		Password: "",
		DB:       0,
	})
	defer redisClient.Close()

	pubsub := redisClient.Subscribe(context.Background(), "plagiarism_channel")
	defer pubsub.Close()

	// Set up a gRPC connection to the plagiarism_checker microservice.
	plagiarismCheckerConn, err := grpc.Dial(config.PlagiarismCheckerAddress, grpc.WithInsecure())
	if err != nil {
		log.Fatal("Error connecting to plagiarism_checker:", err)
	}
	defer plagiarismCheckerConn.Close()

	// Create a plagiarism_checker client.
	plagiarismCheckerClient := plagiarism_checker.NewPlagiarismCheckerClient(plagiarismCheckerConn)

	// Use a WaitGroup to ensure all Goroutines are finished before exiting.
	var wg sync.WaitGroup
	wg.Add(1)

	// Goroutine for handling updates, forwarding messages, and saving them to the database.
	go func() {
		defer wg.Done()
		for update := range updates {
			if update.ChannelPost != nil {
				if update.ChannelPost.Text != "" {
					fmt.Println("Received new message:", update.ChannelPost.Text)
					// Redirect the message to chat.
					msg := tgbotapi.NewMessage(config.ChatID, update.ChannelPost.Text)
					msg.ParseMode = "Markdown"
					bot.Send(msg)

					// Construct the URL to the exact post in the channel.
					channelURL := fmt.Sprintf("https://t.me/%s/%d", update.ChannelPost.Chat.UserName, update.ChannelPost.MessageID)

					// Save the redirected text messages and the URL to the database.
					_, err := db.Exec(
						"INSERT INTO messages (message_text, timestamp, channel_username, message_url) VALUES ($1, $2, $3, $4)",
						update.ChannelPost.Text,
						time.Now(),
						"@"+update.ChannelPost.Chat.UserName,
						channelURL,
					)
					if err != nil {
						log.Println("Error saving message to the database:", err)
					}

					// Check plagiarism using the plagiarism_checker microservice.
					checkPlagiarismRequest := &plagiarism_checker.CheckPlagiarismRequest{
						MessageText: update.ChannelPost.Text,
					}

					go func() {
						// Publish to Redis
						_, err := redisClient.Publish(context.Background(), "plagiarism_channel", update.ChannelPost.Text).Result()
						if err != nil {
							log.Println("Error publishing message to plagiarism_channel:", err)
						}

						// Perform plagiarism check
						fmt.Println("Sending plagiarism check request:", update.ChannelPost.Text)
						plagiarismResponse, err := plagiarismCheckerClient.CheckPlagiarism(context.Background(), checkPlagiarismRequest)
						if err != nil {
							log.Println("Error checking plagiarism:", err)
							return
						}

						if plagiarismResponse.IsPlagiarized {
							log.Println("Plagiarism detected! Deleting message.")
							_, err := db.Exec("DELETE FROM messages WHERE message_text = $1", update.ChannelPost.Text)
							if err != nil {
								log.Println("Error deleting message from the database:", err)
							}
						}
					}()
				} else if update.ChannelPost.Caption != "" {
					// Handle media messages with captions (like photos, videos, etc.).
					// And redirect the caption to chat.
					msg := tgbotapi.NewMessage(config.ChatID, update.ChannelPost.Caption)
					msg.ParseMode = "Markdown"
					bot.Send(msg)

					// Construct the URL to the exact post in the channel.
					channelURL := fmt.Sprintf("https://t.me/%s/%d", update.ChannelPost.Chat.UserName, update.ChannelPost.MessageID)

					// Save the redirected captions and the URL to the database.
					_, err := db.Exec(
						"INSERT INTO messages (message_text, timestamp, channel_username, message_url) VALUES ($1, $2, $3, $4)",
						update.ChannelPost.Caption,
						time.Now(),
						"@"+update.ChannelPost.Chat.UserName,
						channelURL,
					)
					if err != nil {
						log.Println("Error saving caption to the database:", err)
					}
				}
			}
		}
	}()

	// Sending email task Part 2.
	// Calculate the duration until the next 4 AM UTC
	now := time.Now().UTC()
	next4AM := time.Date(now.Year(), now.Month(), now.Day(), 4, 0, 0, 0, time.UTC)
	if now.After(next4AM) {
		next4AM = next4AM.Add(24 * time.Hour)
	}
	timeUntilNext4AM := next4AM.Sub(now)

	// Create a timer to run the email sending process at the next 4 AM UTC
	emailTimer := time.NewTimer(timeUntilNext4AM)
	defer emailTimer.Stop()

	// Use a goroutine to handle the email sending process.
	go func() {
		for {
			<-emailTimer.C

			last24Hours := time.Now().Add(-24 * time.Hour)
			rows, err := db.Query("SELECT message_text, timestamp, message_url FROM messages WHERE timestamp > $1", last24Hours)
			if err != nil {
				log.Println("Error fetching messages from the database:", err)
				continue
			}
			defer rows.Close()

			var emailBody string
			for rows.Next() {
				var messageText, timestamp, messageUrl string
				err := rows.Scan(&messageText, &timestamp, &messageUrl)
				if err != nil {
					log.Println("Error scanning rows:", err)
					continue
				}
				emailBody += fmt.Sprintf("\nЛінк: %s\nПовідомлення: %s\n\n", messageUrl, messageText)
			}

			if emailBody != "" {
				date := time.Now()
				subject := fmt.Sprintf("Інформація за %s", date.Format("02-1-2006"))
				err := sendEmail(subject, emailBody)
				if err != nil {
					log.Println("Error sending email:", err)
				} else {
					log.Println("Email sent successfully!")
				}
			}

			// Calculate the duration until the next 4 AM UTC for the next day
			next4AM = next4AM.Add(24 * time.Hour)
			timeUntilNext4AM = next4AM.Sub(time.Now().UTC())
			emailTimer.Reset(timeUntilNext4AM)
		}
	}()

	mux := mux.NewRouter()
	// Define API routes
	mux.HandleFunc("/", getMessagesHandler)
	// Serve static files
	mux.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("."))))

	// Start the HTTPS server
	go func() {
		tlsConfig := &tls.Config{
			CurvePreferences: []tls.CurveID{tls.X25519, tls.CurveP256},
		}

		srv := &http.Server{
			Addr:        ":9999",
			Handler:     mux,
			TLSConfig:   tlsConfig,
			IdleTimeout: time.Minute,
			ReadTimeout: 5 * time.Second,
		}

		log.Print("Starting server...")

		err := srv.ListenAndServeTLS("./tls/cert.pem", "./tls/key.pem")
		log.Fatal(err)
	}()

	// Wait for all Goroutines to finish before exiting.
	wg.Wait()

	// Fetch the saved messages from the database in ascending order by the timestamp.
	rows, err := db.Query("SELECT message_text, timestamp, channel_username, message_url FROM messages ORDER BY timestamp ASC")
	if err != nil {
		log.Println("Error fetching messages from the database:", err)
	}
	defer rows.Close()

	// Process the fetched rows.
	for rows.Next() {
		var messageText, timestamp, channelUsername, messageURL string
		err := rows.Scan(&messageText, &timestamp, &channelUsername, &messageURL)
		if err != nil {
			log.Println("Error scanning rows:", err)
		}
		cacheKey := fmt.Sprintf("message:%s", messageText)
		cachedMessage, err := redisClient.Get(context.Background(), cacheKey).Result()
		if err == nil {
			log.Printf("Cached Message: %s", cachedMessage)
		} else {
			log.Printf("Caching Message: %s", messageText)
			redisClient.Set(context.Background(), cacheKey, messageText, time.Duration(5*time.Minute))
		}
		logEntryKey := fmt.Sprintf("log:%s", time.Now().UTC().Format(time.RFC3339Nano))
		logEntry := fmt.Sprintf("Message: %s, Timestamp: %s, Channel: %s, URL: %s", messageText, timestamp, channelUsername, messageURL)
		redisClient.Set(context.Background(), logEntryKey, logEntry, 0)

		// Process the fetched data.
		log.Printf("Message: %s, Timestamp: %s, Channel: %s, URL: %s", messageText, timestamp, channelUsername, messageURL)
	}
	if err := rows.Err(); err != nil {
		log.Println("Error iterating rows:", err)
	}
	// Fetch logs from Redis
	logKeys, err := redisClient.Keys(context.Background(), "log:*").Result()
	if err != nil {
		log.Println("Error fetching log keys from Redis:", err)
	}

	// Process fetched logs
	for _, logKey := range logKeys {
		logEntry, err := redisClient.Get(context.Background(), logKey).Result()
		if err != nil {
			log.Println("Error fetching log entry from Redis:", err)
			continue
		}
		log.Printf("Retrieved Log: %s", logEntry)
	}
}
