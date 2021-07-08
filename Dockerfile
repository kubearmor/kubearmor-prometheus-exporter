FROM golang
WORKDIR /home/test
COPY . .
RUN GOOS=linux GOARCH=amd64 go build -o main main.go
CMD ["./main"]