build-local:
	docker buildx create --use
	docker buildx build --platform linux/amd64,linux/arm64 -t ghcr.io/lastmilesolutions/php-observability:latest .


build-push:
	docker buildx create --use
	docker buildx build --platform linux/amd64,linux/arm64 -t ghcr.io/lastmilesolutions/php-observability:latest --push .
