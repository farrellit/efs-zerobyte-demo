
run-readwrite:
	PASSES=100 go run ./readwrite.go

cleanup:
	find efs/ -type f -delete
	redis-cli -n 1 del fileq

stack: build push stackonly

build:
	GOOS=linux GOARCH=amd64 go build readwrite.go
	docker build -t readwrite -f Dockerfile.readwrite .
	eval "`docker run --rm farrellit/awscli ecr get-login --no-include-email --region us-east-1`";

push:
	docker tag readwrite:latest 097202842911.dkr.ecr.us-east-1.amazonaws.com/efs-testing/readwrite:latest
	docker push 097202842911.dkr.ecr.us-east-1.amazonaws.com/efs-testing/readwrite:latest

stackonly:
	python -mjson.tool stack.json >/dev/null
	docker run --rm -it -v `pwd`:/code farrellit/awscli --region us-east-1 cloudformation validate-template --template-body file:///code/stack.json 
	docker run -e AWS_DEFAULT_REGION=us-east-1 --rm -it -v `pwd`:/code --entrypoint bash farrellit/awscli -c 'aws cloudformation describe-stacks --stack-name efs-testing-readwrite; if [ $$? == 0 ]; then action=update-stack; else action=create-stack; fi; aws cloudformation $$action --template-body file:///code/stack.json --stack-name efs-testing-readwrite --region us-east-1 --capabilities CAPABILITY_IAM'
