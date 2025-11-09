#!/usr/bin/env bash
export DOCKER_HOST=$(docker context inspect --format '{{.Endpoints.docker.Host}}' $(docker context show))

if command -v docker-compose &> /dev/null; then
    export DOCKER_COMPOSE_BIN="docker-compose"
else
    export DOCKER_COMPOSE_BIN="docker compose"
fi

