changelog-install:
	@echo "Installing git-chglog...\r"
	brew tap git-chglog/git-chglog
	brew install git-chglog
.PHONY: changelog-install

changelog:
	@echo "Run like 'git tag vX.X.X' to create a tag first\r"
	@echo "Updating CHANGELOG...\r"
	git-chglog --output CHANGELOG.md
.PHONY: changelog
