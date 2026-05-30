package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		jsonData := `[
			{
				"createRule": "user = @request.auth.id",
				"deleteRule": "user = @request.auth.id",
				"fields": [
					{
						"autogeneratePattern": "[a-z0-9]{15}",
						"hidden": false,
						"id": "text3208210256",
						"max": 15,
						"min": 15,
						"name": "id",
						"pattern": "^[a-z0-9]+$",
						"presentable": false,
						"primaryKey": true,
						"required": true,
						"system": true,
						"type": "text"
					},
					{
						"cascadeDelete": false,
						"collectionId": "_pb_users_auth_",
						"hidden": false,
						"id": "relation4019283746",
						"maxSelect": 1,
						"minSelect": 0,
						"name": "user",
						"presentable": false,
						"required": true,
						"system": false,
						"type": "relation"
					},
					{
						"autogeneratePattern": "",
						"hidden": false,
						"id": "text3829174650",
						"max": 100,
						"min": 1,
						"name": "name",
						"pattern": "",
						"presentable": true,
						"primaryKey": false,
						"required": true,
						"system": false,
						"type": "text"
					},
					{
						"autogeneratePattern": "",
						"hidden": false,
						"id": "text2948173650",
						"max": 64,
						"min": 64,
						"name": "token_hash",
						"pattern": "^[a-f0-9]+$",
						"presentable": false,
						"primaryKey": false,
						"required": true,
						"system": false,
						"type": "text"
					},
					{
						"autogeneratePattern": "",
						"hidden": false,
						"id": "text1827364509",
						"max": 12,
						"min": 12,
						"name": "token_prefix",
						"pattern": "",
						"presentable": false,
						"primaryKey": false,
						"required": true,
						"system": false,
						"type": "text"
					},
					{
						"hidden": false,
						"id": "autodate2990389176",
						"name": "created",
						"onCreate": true,
						"onUpdate": false,
						"presentable": false,
						"system": false,
						"type": "autodate"
					},
					{
						"hidden": false,
						"id": "autodate3332085495",
						"name": "updated",
						"onCreate": true,
						"onUpdate": true,
						"presentable": false,
						"system": false,
						"type": "autodate"
					}
				],
				"id": "pbc_5091827364",
				"indexes": [
					"CREATE INDEX ` + "`" + `idx_api_tokens_token_prefix` + "`" + ` ON ` + "`" + `api_tokens` + "`" + ` (` + "`" + `token_prefix` + "`" + `)",
					"CREATE INDEX ` + "`" + `idx_api_tokens_user` + "`" + ` ON ` + "`" + `api_tokens` + "`" + ` (` + "`" + `user` + "`" + `)"
				],
				"listRule": "user = @request.auth.id",
				"name": "api_tokens",
				"system": false,
				"type": "base",
				"updateRule": null,
				"viewRule": "user = @request.auth.id"
			}
		]`

		return app.ImportCollectionsByMarshaledJSON([]byte(jsonData), false)
	}, func(app core.App) error {
		collection, err := app.FindCollectionByNameOrId("api_tokens")
		if err != nil {
			return nil
		}
		return app.Delete(collection)
	})
}
