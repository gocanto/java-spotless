# =============================================================================
# Makefile - Spotless Wrapper Commands
# =============================================================================

.PHONY: spotless-check spotless-apply spotless-check-all spotless-apply-all spotless-build spotless-help spotless-online spotless-offline spotless-qa spotless-staging spotless-main spotless-apply-online spotless-apply-offline spotless-apply-qa spotless-apply-staging spotless-apply-main

REPO_DIR ?= .
SPOTLESS_OFFLINE ?= true
SPOTLESS_BIN = go run ./cmd/spotless

## Check code formatting on changed files
## Usage: make spotless-check TARGET_BASE_BRANCH=<branch>
spotless-check:
ifndef TARGET_BASE_BRANCH
	$(error TARGET_BASE_BRANCH is required. Usage: make spotless-check TARGET_BASE_BRANCH=<branch>)
endif
	@$(SPOTLESS_BIN) check --repo-dir "$(REPO_DIR)" --base-branch "$(TARGET_BASE_BRANCH)" --offline=$(SPOTLESS_OFFLINE)

## Apply code formatting on changed files
## Usage: make spotless-apply TARGET_BASE_BRANCH=<branch>
spotless-apply:
ifndef TARGET_BASE_BRANCH
	$(error TARGET_BASE_BRANCH is required. Usage: make spotless-apply TARGET_BASE_BRANCH=<branch>)
endif
	@$(SPOTLESS_BIN) apply --repo-dir "$(REPO_DIR)" --base-branch "$(TARGET_BASE_BRANCH)" --offline=$(SPOTLESS_OFFLINE)

## Check code formatting on the whole project
spotless-check-all:
	@$(SPOTLESS_BIN) check --repo-dir "$(REPO_DIR)" --all --offline=$(SPOTLESS_OFFLINE)

## Apply code formatting on the whole project
spotless-apply-all:
	@$(SPOTLESS_BIN) apply --repo-dir "$(REPO_DIR)" --all --offline=$(SPOTLESS_OFFLINE)

## Build spotless docker image with required env vars
spotless-build:
	@HOST_UID=$$(id -u) HOST_GID=$$(id -g) CODE_STYLE_ENV_ENABLED=true REPO_DIR=$$(cd "$(REPO_DIR)" && pwd) \
		docker-compose -f docker-compose.spotless.yml --profile spotless build --no-cache

## Helper targets
spotless-online:
ifndef TARGET_BASE_BRANCH
	$(error TARGET_BASE_BRANCH is required. Usage: make spotless-online TARGET_BASE_BRANCH=<branch>)
endif
	@$(MAKE) spotless-check SPOTLESS_OFFLINE=false TARGET_BASE_BRANCH=$(TARGET_BASE_BRANCH)

spotless-offline:
ifndef TARGET_BASE_BRANCH
	$(error TARGET_BASE_BRANCH is required. Usage: make spotless-offline TARGET_BASE_BRANCH=<branch>)
endif
	@$(MAKE) spotless-check SPOTLESS_OFFLINE=true TARGET_BASE_BRANCH=$(TARGET_BASE_BRANCH)

spotless-qa:
	@$(MAKE) spotless-check TARGET_BASE_BRANCH=qa

spotless-staging:
	@$(MAKE) spotless-check TARGET_BASE_BRANCH=staging

spotless-main:
	@$(MAKE) spotless-check TARGET_BASE_BRANCH=main

spotless-apply-online:
ifndef TARGET_BASE_BRANCH
	$(error TARGET_BASE_BRANCH is required. Usage: make spotless-apply-online TARGET_BASE_BRANCH=<branch>)
endif
	@$(MAKE) spotless-apply SPOTLESS_OFFLINE=false TARGET_BASE_BRANCH=$(TARGET_BASE_BRANCH)

spotless-apply-offline:
ifndef TARGET_BASE_BRANCH
	$(error TARGET_BASE_BRANCH is required. Usage: make spotless-apply-offline TARGET_BASE_BRANCH=<branch>)
endif
	@$(MAKE) spotless-apply SPOTLESS_OFFLINE=true TARGET_BASE_BRANCH=$(TARGET_BASE_BRANCH)

spotless-apply-qa:
	@$(MAKE) spotless-apply TARGET_BASE_BRANCH=qa

spotless-apply-staging:
	@$(MAKE) spotless-apply TARGET_BASE_BRANCH=staging

spotless-apply-main:
	@$(MAKE) spotless-apply TARGET_BASE_BRANCH=main

## Show spotless help
spotless-help:
	@sh -c '\
	  line() { \
	    cmd="$$1"; desc="$$2"; \
	    pad=$$((62 - $${#cmd})); \
	    if [ $$pad -lt 2 ]; then pad=2; fi; \
	    dots=$$(printf "%*s" $$pad "" | tr " " "."); \
	    printf "  \033[1m%s\033[0m \033[2;90m%s\033[0m %s\n" "$$cmd" "$$dots" "$$desc"; \
	  }; \
	  printf "\n\033[1;36mCode Style Commands\033[0m \033[2m(changed files by default)\033[0m\n\n"; \
	  line "make spotless-check TARGET_BASE_BRANCH=main" "Check formatting on changed files"; \
	  line "make spotless-apply TARGET_BASE_BRANCH=main" "Apply formatting on changed files"; \
	  line "make spotless-check-all" "Check formatting on all Java files"; \
	  line "make spotless-apply-all" "Apply formatting on all Java files"; \
	  line "make spotless-build" "Build spotless image"; \
	  printf "\n\033[1;33mBranch Shortcuts\033[0m\n\n"; \
	  line "make spotless-main" "Check formatting vs main"; \
	  line "make spotless-staging" "Check formatting vs staging"; \
	  line "make spotless-qa" "Check formatting vs qa"; \
	  line "make spotless-apply-main" "Apply formatting vs main"; \
	  line "make spotless-apply-staging" "Apply formatting vs staging"; \
	  line "make spotless-apply-qa" "Apply formatting vs qa"; \
	  printf "\n\033[1;35mOnline / Offline\033[0m\n\n"; \
	  line "make spotless-online TARGET_BASE_BRANCH=main" "Check formatting online"; \
	  line "make spotless-offline TARGET_BASE_BRANCH=main" "Check formatting offline"; \
	  line "make spotless-apply-online TARGET_BASE_BRANCH=main" "Apply formatting online"; \
	  line "make spotless-apply-offline TARGET_BASE_BRANCH=main" "Apply formatting offline"; \
	  printf "\n\033[1;32mEnv\033[0m\n\n"; \
	  line "REPO_DIR=/path/to/java/repo" "Repository to run Spotless against (default: .)"; \
	  line "SPOTLESS_OFFLINE=true|false" "Offline mode for local Docker runs (default: true)"; \
	  printf "\n"; \
	'
