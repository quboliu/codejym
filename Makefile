.PHONY: test test-backend test-frontend smoke release-check deploy-local down

test:
	./scripts/test-local.sh gate

test-backend:
	./scripts/test-local.sh backend

test-frontend:
	./scripts/test-local.sh frontend

smoke:
	./scripts/test-local.sh smoke

release-check:
	./scripts/test-local.sh release

deploy-local:
	KEEP_SERVICES=1 ./scripts/test-local.sh smoke

down:
	docker compose -p codejym-smoke -f config/docker-compose.yml down -v
