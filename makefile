#-----------------------#
# Go make utils #
generate-schema:
	@echo "Now generating schemas for new graphql model"
	go get github.com/99designs/gqlgen
	go run github.com/99designs/gqlgen generate .

go-format: # Formats the entire directory
	@echo "Now formatting v2 directory"
	go fmt .