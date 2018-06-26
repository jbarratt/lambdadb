.PHONY: deploy

main.zip: cmd/baconlambda/main.go
	GOOS=linux GOARCH=amd64 go build -o main cmd/baconlambda/main.go
	cp data/bacon.gob .
	zip main.zip main bacon.gob
	rm bacon.gob
	mv main.zip terraform/

deploy: main.zip
	cd terraform; terraform plan -out theplan; terraform apply theplan

baconcli: cmd/baconcli/main.go bacon/graph.go
	go build -o baconcli cmd/baconcli/main.go
