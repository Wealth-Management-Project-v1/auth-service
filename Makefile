docker-build:
	git pull
	docker build -t auth-service .