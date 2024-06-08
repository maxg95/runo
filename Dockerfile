FROM golang:1.20-alpine
.
WORKDIR /app

COPY . .

RUN go mod tidy && go mod download && go build -o main . && chmod +x ./main

EXPOSE 50051
EXPOSE 9999

COPY entrypoint.sh /app/entrypoint.sh
RUN chmod +x /app/entrypoint.sh

CMD ["/app/entrypoint.sh"]
