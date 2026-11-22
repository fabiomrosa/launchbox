build:
	./build.sh
dry:
	@if [ -z "$(v)" ]; then \
		echo "⚠️  Missing version bump type!"; \
		echo "👉  Usage: make dry v=patch|minor|major|x.y.z"; \
		exit 1; \
	fi
	@DRY_RUN=true ./release.sh $(v)
major:
	./release.sh major
minor:
	./release.sh minor
patch:
	./release.sh patch
release:
	@if [ -z "$(v)" ]; then \
		echo "Usage: make release v=x.y.z"; \
		exit 1; \
	fi
	@./release.sh $(v)
