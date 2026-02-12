#!/bin/sh
# =============================================================================
# LOCAL DEVELOPMENT ONLY - Entrypoint for Spotless container
# =============================================================================

if [ "$CODE_STYLE_ENV_ENABLED" != "true" ]; then
    echo "ERROR: CODE_STYLE_ENV_ENABLED environment variable must be set to 'true'"
    echo "This container is for local development only."
    echo ""
    echo "Usage: make spotless-check"
    echo "       make spotless-apply"
    echo "       make spotless-check-all"
    echo "       make spotless-apply-all"
    exit 1
fi

exec "$@"
