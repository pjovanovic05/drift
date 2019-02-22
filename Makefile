all:
	GOOS=linux GOARCH=amd64 go build -tags osusergo drift.go client.go server.go