all: stack config remove install ui run

geosim: **/*.go cmd/install-files/*
	go build .

.PHONY: stack
stack: *.tf
	terraform init
	terraform apply -auto-approve

.PHONY: destroy
destroy: *.tf
	terraform destroy -auto-approve

.PHONY: ui
ui: stack
	@open -n -a "Google Chrome" --args --guest \
		$(shell terraform output -raw user-kibana_endpoint)
	@echo -n "User Name: " && \
		terraform output -raw user-elastic_username && \
		echo
	@terraform output -raw user-elastic_password | pbcopy

.PHONY: config
config: geosim stack
	@rm geosim-config.yaml >& /dev/null || true
	@./geosim config \
		-p $(shell terraform output -raw user-elastic_password) \
		-u $(shell terraform output -raw user-elastic_username) \
		-U $(shell terraform output -raw user-elasticsearch_endpoint) \
		-K $(shell terraform output -raw user-kibana_endpoint) \
		-c geosim-config.yaml

.PHONY: remove
remove: geosim
	./geosim remove -f

.PHONY: install
install: geosim
	./geosim install

run: geosim
	./geosim run