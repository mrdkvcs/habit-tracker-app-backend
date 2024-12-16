module github.com/mrdkvcs/go-base-backend

go 1.22.4

require github.com/joho/godotenv v1.5.1

require github.com/google/uuid v1.6.0

require (
	github.com/golang-jwt/jwt/v5 v5.2.1
	github.com/gorilla/websocket v1.5.3
	github.com/jub0bs/cors v0.2.0
	github.com/lib/pq v1.10.9
	github.com/sashabaranov/go-openai v1.30.3
	golang.org/x/crypto v0.24.0
	golang.org/x/oauth2 v0.24.0
)

require (
	cloud.google.com/go/compute/metadata v0.3.0 // indirect
	golang.org/x/net v0.25.0 // indirect
	golang.org/x/text v0.16.0 // indirect
)
