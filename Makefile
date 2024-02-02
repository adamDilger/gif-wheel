build: main.go
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o bootstrap main.go
	zip myFunction.zip bootstrap
	aws lambda update-function-code --function-name gif-wheel \
		--zip-file fileb://myFunction.zip



	# aws lambda create-function --function-name myFunction \
	# 	--runtime provided.al2023 --handler bootstrap \
	# 	--architectures arm64 \
	# 	--role arn:aws:iam::178665545081:role/gif-wheel-deploy \
	# 	--zip-file fileb://myFunction.zip
