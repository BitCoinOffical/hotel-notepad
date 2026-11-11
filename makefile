SERVICES := auth-service guestbook-service

.PHONY: test test-%

test:
	@for s in $(SERVICES); do $(MAKE) test-$$s || exit 1; done

test-%:
	cd backend/$* && go test -v -race ./...