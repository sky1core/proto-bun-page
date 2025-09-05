PROTO_DIR := proto
GO_PKG := github.com/sky1core/proto-bun-page

.PHONY: proto
proto:
	@echo "Generating Go from proto..."
	protoc -I $(PROTO_DIR) \
		--go_out=. --go_opt=paths=source_relative \
		$(PROTO_DIR)/pager/v1/pager.proto
	@echo "Done."

