.PHONY: init
init: dist
	mkdir phpmyadmin
	mkdir mariadb
	touch service.json

.PHONY: dist
dist:
	curl -L -Ss https://github.com/traPtitech/knoQ-UI/releases/latest/download/dist.tar.gz | tar zxv -C ./

.PHONY: clean
clean:
	rm -rf dist/
	rm -rf mariadb
	rm -rf phpmyadmin
	rm -f service.json
