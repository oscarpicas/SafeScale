GO?=go

.PHONY: clean generate nfs

VERSIONFILE := version.go
REV := `git rev-parse HEAD 2>/dev/null || echo ""`

all: nfs generate

vet:
	@$(GO) vet ./...

nfs:
	@(cd nfs && $(MAKE) all)

generate: clean gensrc
	@(cd nfs && $(MAKE) $(@))
	@(go generate .)

gensrc:
	@$(RM) $(VERSIONFILE) || true
	@echo "package system" > $(VERSIONFILE)
	@echo "const (" >> $(VERSIONFILE)
	@echo "  REV = \"$(REV)\"" >> $(VERSIONFILE)
	@echo ")" >> $(VERSIONFILE)

clean:
	@(cd nfs && $(MAKE) $(@))
	@$(RM) rice-box.go || true
