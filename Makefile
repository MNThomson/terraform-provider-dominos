build: *.go
	go get && go build -o terraform-provider-dominos ./

localinstall:
	make clean
	make build
	mkdir -p ~/.terraform.d/plugins/terraform.local/mnthomson/dominos/0.1.0/linux_amd64/
	cp terraform-provider-dominos ~/.terraform.d/plugins/terraform.local/mnthomson/dominos/0.1.0/linux_amd64/terraform-provider-dominos_v0.1.0

clean:
	rm -rf .terraform .terraform.lock.hcl
	rm -rf terraform-provider-dominos
