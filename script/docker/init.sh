type docker-compose >/dev/null || {
    curl -L https://github.com/docker/compose/releases/download/1.11.2/run.sh > docker-compose
    chmod +x docker-compose
    DOCKER_COMPOSE='./docker-compose'
}
export COMPOSE_OPTIONS="-e USER_ID=$(id -ur) -e GROUP_ID=$(id -gr) -e GOOS=$GOOS -e GOARCH=$GOARCH"
DOCKER_COMPOSE=${DOCKER_COMPOSE:-docker-compose}
