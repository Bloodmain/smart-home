migrate-up:
	migrate -path=./migrations -database postgres://postgres:postgres@127.0.0.1:5432/db?sslmode=disable up

migrate-down:
	migrate -path=./migrations -database postgres://postgres:postgres@127.0.0.1:5432/db?sslmode=disable down

controller-build:
	docker build -t homecontroller:v1 .

controller-run:
	docker run -p 8080:8080 -p 8000:8000 --link db --net dbNetwork homecontroller:v1