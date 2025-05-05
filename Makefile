docker-build-and-push:
	docker build --platform linux/arm64/v8,linux/amd64 -t git.kmsign.ru/royalcat/konfa-voice:latest --push .