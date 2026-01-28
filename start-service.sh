#!/bin/bash

set -e

cd "$(dirname "$0")"

# Create .env from example if missing
if [ ! -f ".env" ] && [ -f ".env.example" ]; then
    cp .env.example .env
    echo "Created .env from .env.example"
fi

case "${1:-start}" in
    start)
        docker-compose up -d --build
        echo "Services started. Run '$0 logs' to view logs."
        ;;
    stop)
        docker-compose down
        ;;
    logs)
        docker-compose logs -f
        ;;
    status)
        docker-compose ps
        ;;
    clean)
        docker-compose down -v
        ;;
    *)
        echo "Usage: $0 {start|stop|logs|status|clean}"
        exit 1
        ;;
esac
