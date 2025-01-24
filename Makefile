swagger-user:
	swag init -g cmd/user-service/main.go -o cmd/user-service/docs

swagger-product:
	swag init -g cmd/product-service/main.go -o cmd/product-service/docs

swagger-recommendation:
	swag init -g cmd/recommendation-service/main.go -o cmd/recommendation-service/docs

swagger-analytics:
	swag init -g cmd/analytics-service/main.go -o cmd/analytics-service/docs

swagger-sso:
	swag init -g cmd/sso-service/main.go -o cmd/sso-service/docs

swagger: swagger-user swagger-product swagger-recommendation swagger-analytics swagger-sso

backup:
	./backup.sh

clean:
	find . -name __pycache__ -type d -print0|xargs -0 rm -r --
	rm -rf .idea/

docker_clean:
	sudo docker stop $$(sudo docker ps -a -q) || true
	sudo docker rm $$(sudo docker ps -a -q) || true

docker_down:
	docker compose -f docker-compose.yaml down && docker network prune --force

mkdocs_build:
	mkdocs build

mkdocs:
	mkdocs serve

start:
	docker compose up --build