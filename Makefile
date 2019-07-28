ensure-deps:
	go mod tidy && go mod vendor

run:
	go run main.go -api-token=${TELEGRAM_APITOKEN} -log-level=${LOG_LEVEL}

build:
	GOOS=darwin go build -o dist/darwin/telegram_bot_ex
	GOOS=linux go build -o dist/linux/telegram_bot_ex
