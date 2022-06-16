VERSION =0.1.1

build: *.go
	go get && go build -o terraform-provider-dominos ./

localInstall:
	make clean
	make build
	mkdir -p ~/.terraform.d/plugins/terraform.local/mnthomson/dominos/$(VERSION)/linux_amd64/
	cp terraform-provider-dominos ~/.terraform.d/plugins/terraform.local/mnthomson/dominos/$(VERSION)/linux_amd64/terraform-provider-dominos_v$(VERSION)
	terraform init

clean:
	rm -rf .terraform .terraform.lock.hcl
	rm -rf terraform-provider-dominos

localTest:
	clear
	make localInstall
	TF_LOG=TRACE terraform plan
