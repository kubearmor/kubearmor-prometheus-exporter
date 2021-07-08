CURDIR=$(shell pwd)

.PHONY: build
build:
	cd $(CURDIR); go mod tidy
	cd $(CURDIR); go build -o kubearmor-prometheus-exporter main.go

.PHONY: run
run: $(CURDIR)/kubearmor-prometheus-exporter
	cd $(CURDIR); ./kubearmor-prometheus-exporter

.PHONY: build-image
build-image:
	cd $(CURDIR); docker build -t kubearmor/kubearmor-prometheus-exporter:latest .

.PHONY: push-image
push-image:
	cd $(CURDIR); docker push kubearmor/kubearmor-prometheus-exporter:latest

.PHONY: clean
clean:
	cd $(CURDIR); sudo rm -f kubearmor-prometheus-exporter
	#cd $(CURDIR); find . -name go.sum | xargs -I {} rm -f {}
