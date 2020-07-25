build:
	cd margarita && docker build -t poolside/margarita:latest .

dev:
	docker-compose -f docker-compose.dev.yml up

test-server:
	python3 -m http.server 8000 --bind 127.0.0.1