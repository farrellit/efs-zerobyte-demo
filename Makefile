
run-readwrite:
	PASSES=100 go run ./readwrite.go

cleanup:
	find efs/ -type f -delete
	redis-cli -n 1 del fileq

stack: build push stackonly

build:
	GOOS=linux GOARCH=amd64 go build readwrite.go
	docker build -t readwrite -f Dockerfile.readwrite .

push: build
	eval "`docker run --rm farrellit/awscli ecr get-login --no-include-email --region us-east-1`";
	acct=$$(docker run --rm farrellit/awscli --region us-east-1 sts get-caller-identity --query Account --output text); docker tag readwrite:latest $$acct.dkr.ecr.us-east-1.amazonaws.com/efs-testing/readwrite:latest; docker push $$acct.dkr.ecr.us-east-1.amazonaws.com/efs-testing/readwrite:latest

stackonly:
	which python && -mjson.tool stack.json >/dev/null # opportunisticlaly check json syntax before a round trip to the cfn api
	docker run --rm -it -v `pwd`:/code farrellit/awscli --region us-east-1 cloudformation validate-template --template-body file:///code/stack.json 
	docker run -e AWS_DEFAULT_REGION=us-east-1 --rm -it -v `pwd`:/code --entrypoint bash farrellit/awscli /code/stack.sh
