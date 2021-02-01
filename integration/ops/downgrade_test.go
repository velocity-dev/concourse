package ops_test

import (
	"strings"
	"testing"

	"github.com/concourse/concourse/integration/internal/dctest"
	"github.com/concourse/concourse/integration/internal/flytest"
)

func TestDowngrade(t *testing.T) {
	t.Parallel()

	devDC := dctest.Init(t, "../docker-compose.yml")

	t.Run("deploy dev", func(t *testing.T) {
		devDC.Run(t, "up", "-d")
	})

	fly := flytest.Init(t, devDC)
	setupUpgradeDowngrade(t, fly)

	latestDC := dctest.Init(t, "../docker-compose.yml", "overrides/latest.yml")

	t.Run("migrate down to latest from clean deploy", func(t *testing.T) {
		// just to see what it was before
		devDC.Run(t, "run", "web", "migrate", "--current-db-version")

		latest := latestDC.Output(t, "run", "web", "migrate", "--supported-db-version")
		latest = strings.TrimRight(latest, "\n")

		devDC.Run(t, "run", "web", "migrate", "--migrate-db-to-version", latest)
	})

	t.Run("deploy latest", func(t *testing.T) {
		latestDC.Run(t, "up", "-d")
	})

	fly = flytest.Init(t, latestDC)
	verifyUpgradeDowngrade(t, fly)

	t.Run("migrate up to dev and deploy dev", func(t *testing.T) {
		devDC.Run(t, "up", "-d")
	})

	fly = flytest.Init(t, devDC)
	verifyUpgradeDowngrade(t, fly)
}
