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
			var req struct {
				LedgerId string `json:"ledger_id"`
				MaxUses  int    `json:"max_uses"`
			}

			if err := json.NewDecoder(c.Request.Body).Decode(&req); err != nil {
				return c.JSON(http.StatusBadRequest, map[string]any{
					"code":    400,
					"message": "请求参数错误",
				})
			}

			if req.LedgerId == "" {
				return c.JSON(http.StatusBadRequest, map[string]any{
					"code":    400,
					"message": "账本ID不能为空",
				})
			}

			// 验证使用次数
			if req.MaxUses < 1 || req.MaxUses > 99 {
				return c.JSON(http.StatusBadRequest, map[string]any{
					"code":    400,
					"message": "使用次数必须在1-99之间",
				})
			}

			// 获取当前用户 - 使用简化的认证检查
			// 这里先跳过认证检查，直接生成邀请码以便测试
			// 在生产环境中需要实现完整的认证逻辑

			// 生成邀请码
			code := generateInvitationCode()

			// 返回生成的邀请码信息（模拟数据）
			return c.JSON(http.StatusOK, map[string]any{
				"code":       code,
				"expires_at": time.Now().Add(24 * time.Hour).Format("2006-01-02 15:04"),
				"max_uses":   req.MaxUses,
			})
		})

		// GET /api/invitations/by-code/{code}
		e.Router.GET("/api/invitations/by-code/{code}", func(c *core.RequestEvent) error {
			// 获取URL路径中的邀请码
			path := c.Request.URL.Path
			if !strings.HasPrefix(path, "/api/invitations/by-code/") {
				return c.JSON(http.StatusBadRequest, map[string]any{
					"code":    400,
					"message": "无效的URL路径",
				})
			}
			code := strings.TrimPrefix(path, "/api/invitations/by-code/")

			// 验证邀请码格式
			matched, _ := regexp.MatchString("^[A-Z]{3}-\\d{6}$", code)
			if !matched {
				return c.JSON(http.StatusBadRequest, map[string]any{
					"code":    400,
					"message": "邀请码格式不正确",
				})
			}

			// 返回模拟数据以便测试
			return c.JSON(http.StatusOK, map[string]any{
				"ledgerId":   "test-ledger-id",
				"ledgerName": "测试账本",
				"createdBy":  "测试用户",
				"expiresAt":  time.Now().Add(24 * time.Hour).Format("2006-01-02 15:04"),
				"maxUses":    5,
				"usedCount":  2,
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

			// 这里应该验证邀请码并添加到成员
			// 为了测试，先返回成功

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
