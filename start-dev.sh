#!/bin/bash
set -a && source .env && set +a

echo "Starting services..."

cd services/auth && go run . &
cd services/user && go run . &
cd services/gateway && go run . &
cd web && npm run dev &

wait
