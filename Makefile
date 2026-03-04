CONTAINER_RUNTIME := "docker"
SIGNAL := "traces"
COUNT := 3

telemetrygen:
	$(CONTAINER_RUNTIME) compose --profile scripts run --rm telemetrygen --otlp-endpoint collector:4317 $(SIGNAL) --otlp-insecure --$(SIGNAL) $(COUNT)
