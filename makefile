# ref: https://www.alexedwards.net/blog/serverless-api-with-go-and-aws-lambda

check: *.go
	go build -o bin/check

linux: *.go
	env GOOS=linux GOARCH=amd64 go build -o bin/check

main: check

lambda: linux
	zip -j bin/main.zip bin/check

role:
	aws iam create-role --role-name lambda-status-checker-executor --assume-role-policy-document file://./trust-policy.json
	aws iam attach-role-policy --role-name lambda-status-checker-executor --policy-arn arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole

create:
	# account=248699473389
	aws lambda create-function --function-name status-checker --description 'For checking webpages status' --timeout 900 --runtime go1.x --role arn:aws:iam::${account}:role/lambda-status-checker-executor --handler check --zip-file fileb://./bin/main.zip

delete:
	aws lambda delete-function --function-name status-checker

invoke:
	aws lambda invoke --function-name status-checker output.json && cat output.json

update: lambda
	aws lambda update-function-code --function-name status-checker --zip-file fileb://./bin/main.zip

create-api:
	aws apigateway create-rest-api --name status-checker

get-root-id:
	# api-id=3h2xo7vssh
	aws apigateway get-resources --rest-api-id ${api-id}

create-api-path:
	# root-id=x3vw38xbm9
	aws apigateway create-resource --rest-api-id ${api-id} --parent-id ${root-id} --path-part check

set-api-method:
	# path-id=73j833
	aws apigateway put-method --rest-api-id ${api-id} --resource-id ${path-id} --http-method ANY --authorization-type NONE

set-api-integration:
	# uri=arn:aws:apigateway:ap-southeast-1:lambda:path/2015-03-31/functions/arn:aws:lambda:ap-southeast-1:248699473389:function:status-checker/invocations
	aws apigateway put-integration --rest-api-id ${api-id} --resource-id ${path-id} --http-method ANY --type AWS_PROXY --integration-http-method POST --uri ${uri}

	# account=248699473389 api-id=3h2xo7vssh
	aws lambda add-permission --function-name status-checker --statement-id sc-statement-id --action lambda:InvokeFunction --principal apigateway.amazonaws.com --source-arn arn:aws:execute-api:ap-southeast-1:${account}:${api-id}:/*/*/*

	aws apigateway create-deployment --rest-api-id ${api-id} --stage-name prod

test-api:
	# api-id=3h2xo7vssh path-id=73j833
	aws apigateway test-invoke-method --rest-api-id ${api-id} --resource-id ${path-id} --http-method "GET"

clean:
	rm -rf bin
