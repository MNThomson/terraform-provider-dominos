VERSION =0.1.1

build: *.go
	go get && go build -o terraform-provider-dominos ./

localInstall:
	make clean
	make build
	mkdir -p .terraform.d/plugins/registry.terraform.io/mnthomson/dominos/$(VERSION)/linux_amd64/
	cp terraform-provider-dominos .terraform.d/plugins/registry.terraform.io/mnthomson/dominos/$(VERSION)/linux_amd64/terraform-provider-dominos_v$(VERSION)
	terraform init -plugin-dir .terraform.d/plugins/

clean:
	rm -rf .terraform .terraform.lock.hcl
	rm -rf terraform.tfstate*
	rm -rf terraform-provider-dominos

localTest:
	clear
	make localInstall
	TF_LOG=TRACE terraform plan
