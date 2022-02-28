changelog-install:
	@echo "Installing git-chglog...\r"
	brew tap git-chglog/git-chglog
	brew install git-chglog
.PHONY: changelog-install

changelog:
	@echo "Updating CHANGELOG...\r"
	git-chglog --output CHANGELOG.md
.PHONY: changelog
