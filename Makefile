default: tests

tests: golangci-lint unittests

unittests:
	@sh -c "'$(CURDIR)/scripts/unit_tests.sh'"

golangci-lint:
	@sh -c "'$(CURDIR)/scripts/golangci_lint_check.sh'"

build:
	@sh -c "'$(CURDIR)/scripts/build.sh'"

.PHONY: tests unittests golangci-lint build
