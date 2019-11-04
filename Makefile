ensure-deps:
	go mod tidy && go mod vendor

run:
	GOOGLE_APPLICATION_CREDENTIALS=./gcloud_cred.json go run main.go --config "./conf.yml"

init-db:
	go run main.go db init --config "./conf.yml"

build:
	GOOS=darwin go build -o dist/darwin/telegram_bot_ex
	GOOS=linux go build -o dist/linux/telegram_bot_ex
