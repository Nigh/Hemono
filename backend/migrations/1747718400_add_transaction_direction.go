package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		collection, err := app.FindCollectionByNameOrId("transactions")
		if err != nil {
			return err
		}

		field := &core.SelectField{
			Name:   "direction",
			Values: []string{"EXPENSE", "INCOME"},
		}
		collection.Fields.Add(field)

		return app.Save(collection)
	}, func(app core.App) error {
		collection, err := app.FindCollectionByNameOrId("transactions")
		if err != nil {
			return err
		}

		collection.Fields.RemoveByName("direction")

		return app.Save(collection)
	})
}
