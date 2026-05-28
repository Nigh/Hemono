package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		ledgers, err := app.FindCollectionByNameOrId("ledgers")
		if err != nil {
			return err
		}
		ledgers.ViewRule = ptr("owner = @request.auth.id || ledger_members_via_ledger.user ?= @request.auth.id")
		if err := app.Save(ledgers); err != nil {
			return err
		}

		members, err := app.FindCollectionByNameOrId("ledger_members")
		if err != nil {
			return err
		}
		members.ListRule = ptr("ledger.owner = @request.auth.id || ledger.ledger_members_via_ledger.user ?= @request.auth.id")
		return app.Save(members)
	}, func(app core.App) error {
		ledgers, err := app.FindCollectionByNameOrId("ledgers")
		if err != nil {
			return err
		}
		ledgers.ViewRule = ptr("owner = @request.auth.id")
		if err := app.Save(ledgers); err != nil {
			return err
		}

		members, err := app.FindCollectionByNameOrId("ledger_members")
		if err != nil {
			return err
		}
		members.ListRule = ptr("ledger.owner = @request.auth.id || user = @request.auth.id")
		return app.Save(members)
	})
}

func ptr(s string) *string {
	return &s
}
