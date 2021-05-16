# Build the Go API
FROM golang:latest AS backend
ADD ./backend /backend
WORKDIR /backend
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-w" -a -o . && mv charityyeti /usr/local/charityyeti
WORKDIR /
RUN rm -rf backend/

# make the Go binary executable and run
WORKDIR /usr/local/
RUN chmod +x ./charityyeti
EXPOSE 8080
CMD ./charityyeti
