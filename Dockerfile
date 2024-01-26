# Utilisez l'image officielle de Go pour la phase de build
FROM golang:1.22rc2-bullseye as builder

# Définissez le répertoire de travail dans le conteneur
WORKDIR /app

# Copiez les fichiers go.mod et go.sum
COPY go.mod go.sum ./

# Téléchargez toutes les dépendances
RUN go mod download

# Copiez le code source dans le conteneur
COPY . .

# Compilez l'application
RUN CGO_ENABLED=0 GOOS=linux go build -v -o main

# Utilisez une image alpine pour la phase de runtime
FROM alpine:latest

# Ajoutez ca-certificates pour les appels HTTPS
RUN apk --no-cache add ca-certificates

# Copiez l'exécutable compilé à partir de la phase de build
COPY --from=builder /app/main /app/main

# Exposez le port sur lequel votre application s'exécute
EXPOSE 8080

# Définissez l'exécutable par défaut pour le conteneur
ENTRYPOINT ["/app/main"]