package main

import (
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/plugins/migratecmd"

	_ "hemono/backend/migrations"
)

// 生成邀请码
func generateInvitationCode() string {
	letters := "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	var letterPart strings.Builder
	for i := 0; i < 3; i++ {
		letterPart.WriteByte(letters[rand.Intn(len(letters))])
	}

	var numberPart strings.Builder
	for i := 0; i < 6; i++ {
		numberPart.WriteByte(byte('0' + rand.Intn(10)))
	}

	return letterPart.String() + "-" + numberPart.String()
}

func main() {
	app := pocketbase.New()
	rand.Seed(time.Now().UnixNano())

	migratecmd.MustRegister(app, app.RootCmd, migratecmd.Config{
		Dir: "migrations",
	})

	// 添加邀请码API路由
	app.OnServe().BindFunc(func(e *core.ServeEvent) error {
		// POST /api/invitations/generate
		e.Router.POST("/api/invitations/generate", func(c *core.RequestEvent) error {
			// 这里需要根据实际 PocketBase 版本调整 API 调用
			// 暂时返回成功以便测试前端
			return c.JSON(http.StatusOK, map[string]any{
				"code":       "ABC-123456",
				"expires_at": time.Now().Add(24 * time.Hour).Format("2006-01-02 15:04"),
			})
		})

		// GET /api/invitations/by-code/{code}
		e.Router.GET("/api/invitations/by-code/{code}", func(c *core.RequestEvent) error {
			// 验证邀请码格式
			code := c.Request.URL.Path[len("/api/invitations/by-code/"):]
			matched, _ := regexp.MatchString("^[A-Z]{3}-\\d{6}$", code)
			if !matched {
				return c.JSON(http.StatusBadRequest, map[string]any{
					"code":    400,
					"message": "邀请码格式不正确",
				})
			}

			// 暂时返回模拟数据以便测试前端
			return c.JSON(http.StatusOK, map[string]any{
				"ledgerId":   "test-ledger-id",
				"ledgerName": "测试账本",
				"createdBy":  "测试用户",
				"expiresAt":  time.Now().Add(24 * time.Hour).Format("2006-01-02 15:04"),
			})
		})

		// POST /api/invitations/join
		e.Router.POST("/api/invitations/join", func(c *core.RequestEvent) error {
			var req struct {
				Code string `json:"code"`
			}

			if err := json.NewDecoder(c.Request.Body).Decode(&req); err != nil {
				return c.JSON(http.StatusBadRequest, map[string]any{
					"code":    400,
					"message": "请求参数错误",
				})
			}

			if req.Code == "" {
				return c.JSON(http.StatusBadRequest, map[string]any{
					"code":    400,
					"message": "邀请码不能为空",
				})
			}

			// 暂时返回成功以便测试前端
			return c.JSON(http.StatusOK, map[string]any{
				"success":    true,
				"ledgerId":   "test-ledger-id",
				"ledgerName": "测试账本",
				"message":    "成功加入账本",
			})
		})

		return e.Next()
	})

	app.OnRecordAfterCreateSuccess("ledgers").BindFunc(func(e *core.RecordEvent) error {
		ledgerId := e.Record.Id
		ownerId := e.Record.GetString("owner")

		collection, err := app.FindCollectionByNameOrId("ledger_members")
		if err != nil {
			return err
		}

		record := core.NewRecord(collection)
		record.Set("ledger", ledgerId)
		record.Set("user", ownerId)
		record.Set("role", "admin")

		return app.Save(record)
	})

	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
}
