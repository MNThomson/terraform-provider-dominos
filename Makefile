VERSION =9.9.9

build: *.go
	go build -o terraform-provider-dominos .

run:
	clear
	rm -rf .terraform.lock.hcl terraform.tfstat*
	make build
	terraform init -plugin-dir .terraform.d/plugins/
	TF_LOG=INFO terraform apply -auto-approve
	rm -rf .terraform.lock.hcl terraform.tfstat*

watch:
	while true; do \
	    inotifywait -e modify,create,delete -r internal/provider/*.go && make run; \
	done

localSetup:
	mkdir -p .terraform.d/plugins/registry.terraform.io/mnthomson/dominos/$(VERSION)/linux_amd64/
	ln -s $$(pwd)/terraform-provider-dominos .terraform.d/plugins/registry.terraform.io/mnthomson/dominos/$(VERSION)/linux_amd64/terraform-provider-dominos_v$(VERSION)

clean:
	rm -rf .terraform .terraform.d .terraform.lock.hcl
	rm -rf terraform.tfstate*
	rm -rf terraform-provider-dominos
