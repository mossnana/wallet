develop:
	air -c air.toml
develop-compose:
	docker-compose -f docker/docker-compose.develop.yml up -d
run-compose:
	docker-compose -f docker/docker-compose.yml up -d
