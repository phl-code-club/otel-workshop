CONTAINER_RUNTIME := ${CONTAINER_RUNTIME}
ifeq ($(strip $(CONTAINER_RUNTIME)),)
	CONTAINER_RUNTIME := docker
endif
COMPOSE = $(CONTAINER_RUNTIME) compose
SIGNAL := traces
COUNT := 3

telemetrygen:
	$(COMPOSE) --profile scripts run --rm telemetrygen --otlp-endpoint collector:4317 $(SIGNAL) --otlp-insecure --$(SIGNAL) $(COUNT)

up:
	$(COMPOSE) up $(ARGS)

logs:
	$(COMPOSE) logs $(SERVICE) $(ARGS)
