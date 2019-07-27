ensure-deps:
	go mod tidy && go mod vendor

run:
	go run main.go

build:
	GOOS=darwin go build -o dist/darwin/telegram_bot_ex
	GOOS=linux go build -o dist/linux/telegram_bot_ex
