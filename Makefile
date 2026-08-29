GORELEASER := go tool goreleaser
SVU := go tool svu

.PHONY: release-check release-snapshot release release-local next-version release-prepare

release-check:
	$(GORELEASER) check

release-snapshot:
	$(GORELEASER) release --clean --snapshot

next-version:
	@$(SVU) next

release-prepare:
	@./scripts/prepare-release.sh

# Create and push a semver tag; the Release workflow publishes to GitHub.
release: release-check
	@test -n "$(TAG)" || (echo "usage: make release TAG=v0.14.0" && exit 1)
	@test -z "$$(git status --porcelain)" || (echo "working tree not clean" && exit 1)
	git tag -a "$(TAG)" -m "Release $(TAG)"
	git push origin "$(TAG)"

# Publish a release from the current tag (local maintainer use).
release-local: release-check
	@test -n "$$GITHUB_TOKEN" || (echo "GITHUB_TOKEN required" && exit 1)
	$(GORELEASER) release --clean
