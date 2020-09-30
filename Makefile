dev:
	docker-compose -f docker-compose.mac.yml up

test-server:
	python3 -m http.server 8000 --bind 0.0.0.0