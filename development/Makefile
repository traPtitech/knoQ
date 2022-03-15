.PHONY: init
init: dist
	go install github.com/fuji8/gotypeconverter/cmd/gotypeconverter@v0.1.10
	touch service.json
	mkdir phpmyadmin
	mkdir mariadb

.PHONY: dist
dist:
	curl -L -Ss https://github.com/traPtitech/knoQ-UI/releases/latest/download/dist.tar.gz | tar zxv -C ./

.PHONY: clean
clean:
	rm -rf dist/
	rm -rf mariadb
	rm -rf phpmyadmin
	rm -f service.json

.PHONY: up
up:
	docker-compose up --build

.PHONY: down
down:
	docker-compose down