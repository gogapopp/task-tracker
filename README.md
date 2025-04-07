# task-tracker
migrate -database "postgres://user:password@localhost:5432/tasktracker?sslmode=disable" -path ./migrations up   

docker compose -f 'docker-compose.yml' build --no-cache    
docker compose -f 'docker-compose.yml' up -d   