COMMIT := $(shell git describe --always --dirty)
PROJECT_ROOT := $(shell git rev-parse --show-toplevel)
IS_DB_RUNNING := $(shell docker inspect -f '{{ .State.Running }}' sidan_sql 2>/dev/null)
IS_SERVICE_RUNNING := $(shell docker inspect -f '{{ .State.Running }}' sidan_sql 2>/dev/null)

default: run

.PHONY: run
run: db-run-$(COMMIT)

.PHONY: db-build-$(COMMIT) db-run-$(COMMIT) db-stop
db-build-$(COMMIT):
ifeq (,$(shell docker images -q "sidan-db:$(COMMIT)"))
	@docker build -f Dockerfile.sql -t "sidan-db:$(COMMIT)" .
else
	$(NOOP)
endif

db-run-$(COMMIT): db-build-$(COMMIT) docker-network-create
ifndef IS_DB_RUNNING
	@docker run \
	--net backend-network \
	-p 3306:3306/tcp \
	-v $(PROJECT_ROOT)/db:/docker-entrypoint-initdb.d \
	--rm -d --name sidan_sql \
	"sidan-db:$(COMMIT)"
else
	$(NOOP)
endif

db-stop:
ifdef IS_DB_RUNNING
	@docker kill sidan_sql
else
	$(NOOP)
endif

.PHONY: docker-network-create docker-network-clean
docker-network-create:
ifeq (,$(shell docker network ls -q --filter name=backend-network))
	@docker network create backend-network
else
	$(NOOP)
endif

docker-network-clean: db-stop
	@docker network rm backend-network
