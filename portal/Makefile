portal: vendor bindata.go *.go
	go build -i -v -ldflags "-X main.appVersion=$$(git rev-parse HEAD)"

portal_linux_amd64: vendor bindata.go *.go
	GOOS=linux GOARCH=amd64 go build -i -v -ldflags "-X main.appVersion=$$(git rev-parse HEAD)" -o portal_linux_amd64

glide-install:
	glide install

bindata.go: static/css/bootstrap.min.css views/*.tmpl
	go-bindata $^

static/css/bootstrap.min.css: node_modules/bootstrap/dist/css/bootstrap.min.css
	mkdir -p static/css
	cp node_modules/bootstrap/dist/css/bootstrap.min.css static/css/bootstrap.min.css

vendor: glide.lock
	glide install

node_modules/bootstrap/dist/css/bootstrap.min.css: package.json
	npm install
	touch node_modules/bootstrap/dist/css/bootstrap.min.css

importteams_linux_amd64: cmd/importteams/main.go
	GOOS=linux GOARCH=amd64 go build -i -v -o importteams_linux_amd64 ./cmd/importteams

deploy: portal_linux_amd64
	cd ../provisioning/portal && ansible-playbook ansible/*deploy.yml -i $(TARGET), $(ANSIBLE_ARGS)

deploy_full: portal_linux_amd64 importteams_linux_amd64
	cd ../provisioning/portal && ansible-playbook ansible/*.yml -i $(TARGET), $(ANSIBLE_ARGS)
	
.PHONY: glide-install
