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

			// 获取当前用户
			authRecord := c.Auth
			if authRecord == nil {
				return c.JSON(http.StatusUnauthorized, map[string]any{
					"code":    401,
					"message": "未认证",
				})
			}

			// 验证账本是否存在且用户是所有者
			ledgerCollection, err := app.FindCollectionByNameOrId("ledgers")
			if err != nil {
				return c.JSON(http.StatusInternalServerError, map[string]any{
					"code":    500,
					"message": "内部服务器错误",
				})
			}

			ledger, err := app.FindRecordById(ledgerCollection, req.LedgerId)
			if err != nil {
				return c.JSON(http.StatusNotFound, map[string]any{
					"code":    404,
					"message": "账本不存在",
				})
			}

			if ledger.GetString("owner") != authRecord.Id {
				return c.JSON(http.StatusForbidden, map[string]any{
					"code":    403,
					"message": "只有账本所有者可以生成邀请码",
				})
			}

			// 生成邀请码
			code := generateInvitationCode()

			// 保存到数据库
			invitationCollection, err := app.FindCollectionByNameOrId("invitation_codes")
			if err != nil {
				return c.JSON(http.StatusInternalServerError, map[string]any{
					"code":    500,
					"message": "内部服务器错误",
				})
			}

			record := core.NewRecord(invitationCollection)
			record.Set("code", code)
			record.Set("ledger", req.LedgerId)
			record.Set("created_by", authRecord.Id)
			record.Set("expires_at", time.Now().Add(24*time.Hour).Format("2006-01-02 15:04:05"))
			record.Set("max_uses", req.MaxUses)
			record.Set("used_count", 0)

			if err := app.Save(record); err != nil {
				return c.JSON(http.StatusInternalServerError, map[string]any{
					"code":    500,
					"message": "保存邀请码失败",
				})
			}

			// 返回生成的邀请码信息
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

			// 查询邀请码
			invitationCollection, err := app.FindCollectionByNameOrId("invitation_codes")
			if err != nil {
				return c.JSON(http.StatusInternalServerError, map[string]any{
					"code":    500,
					"message": "内部服务器错误",
				})
			}

			// 使用filter查询特定的邀请码
			foundRecord, err := app.FindFirstRecordByData(invitationCollection, "code", code)
			if err != nil {
				return c.JSON(http.StatusNotFound, map[string]any{
					"code":    404,
					"message": "邀请码不存在",
				})
			}

			// 检查邀请码是否过期
			expiresAtStr := foundRecord.GetString("expires_at")
			expiresAt, err := time.Parse("2006-01-02 15:04:05", expiresAtStr)
			if err != nil {
				return c.JSON(http.StatusInternalServerError, map[string]any{
					"code":    500,
					"message": "内部服务器错误",
				})
			}

			if time.Now().After(expiresAt) {
				return c.JSON(http.StatusGone, map[string]any{
					"code":    410,
					"message": "邀请码已过期",
				})
			}

			// 检查邀请码是否已用完
			maxUses := foundRecord.GetInt("max_uses")
			usedCount := foundRecord.GetInt("used_count")
			if usedCount >= maxUses {
				return c.JSON(http.StatusConflict, map[string]any{
					"code":    409,
					"message": "邀请码已被使用完",
				})
			}

			// 获取账本信息
			ledgerId := foundRecord.GetString("ledger")
			ledgerCollection, err := app.FindCollectionByNameOrId("ledgers")
			if err != nil {
				return c.JSON(http.StatusInternalServerError, map[string]any{
					"code":    500,
					"message": "内部服务器错误",
				})
			}

			ledger, err := app.FindRecordById(ledgerCollection, ledgerId)
			if err != nil {
				return c.JSON(http.StatusNotFound, map[string]any{
					"code":    404,
					"message": "账本不存在",
				})
			}

			// 获取创建者信息
			createdById := foundRecord.GetString("created_by")
			userCollection, err := app.FindCollectionByNameOrId("users")
			if err != nil {
				return c.JSON(http.StatusInternalServerError, map[string]any{
					"code":    500,
					"message": "内部服务器错误",
				})
			}

			creator, err := app.FindRecordById(userCollection, createdById)
			if err != nil {
				return c.JSON(http.StatusInternalServerError, map[string]any{
					"code":    500,
					"message": "内部服务器错误",
				})
			}

			// 返回邀请码信息
			return c.JSON(http.StatusOK, map[string]any{
				"ledgerId":   ledgerId,
				"ledgerName": ledger.GetString("name"),
				"createdBy":  creator.GetString("name"),
				"expiresAt":  expiresAtStr,
				"maxUses":    maxUses,
				"usedCount":  usedCount,
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

			// 验证邀请码格式
			matched, _ := regexp.MatchString("^[A-Z]{3}-\\d{6}$", req.Code)
			if !matched {
				return c.JSON(http.StatusBadRequest, map[string]any{
					"code":    400,
					"message": "邀请码格式不正确",
				})
			}

			// 获取当前用户
			authRecord := c.Auth
			if authRecord == nil {
				return c.JSON(http.StatusUnauthorized, map[string]any{
					"code":    401,
					"message": "未认证",
				})
			}

			// 查询邀请码
			invitationCollection, err := app.FindCollectionByNameOrId("invitation_codes")
			if err != nil {
				return c.JSON(http.StatusInternalServerError, map[string]any{
					"code":    500,
					"message": "内部服务器错误",
				})
			}

			foundRecord, err := app.FindFirstRecordByData(invitationCollection, "code", req.Code)
			if err != nil {
				return c.JSON(http.StatusNotFound, map[string]any{
					"code":    404,
					"message": "邀请码不存在",
				})
			}

			// 检查邀请码是否过期
			expiresAtStr := foundRecord.GetString("expires_at")
			expiresAt, err := time.Parse("2006-01-02 15:04:05", expiresAtStr)
			if err != nil {
				return c.JSON(http.StatusInternalServerError, map[string]any{
					"code":    500,
					"message": "内部服务器错误",
				})
			}

			if time.Now().After(expiresAt) {
				return c.JSON(http.StatusGone, map[string]any{
					"code":    410,
					"message": "邀请码已过期",
				})
			}

			// 检查邀请码是否已用完
			maxUses := foundRecord.GetInt("max_uses")
			usedCount := foundRecord.GetInt("used_count")
			if usedCount >= maxUses {
				return c.JSON(http.StatusConflict, map[string]any{
					"code":    409,
					"message": "邀请码已被使用完",
				})
			}

			// 获取账本ID
			ledgerId := foundRecord.GetString("ledger")

			// 检查用户是否已经是该账本的成员
			ledgerMembersCollection, err := app.FindCollectionByNameOrId("ledger_members")
			if err != nil {
				return c.JSON(http.StatusInternalServerError, map[string]any{
					"code":    500,
					"message": "内部服务器错误",
				})
			}

			// 查询所有成员记录并检查用户是否已经是成员
			allMemberRecords, err := app.Dao().FindRecords(
				ledgerMembersCollection,
				"",
			)

			if err != nil {
				return c.JSON(http.StatusInternalServerError, map[string]any{
					"code":    500,
					"message": "内部服务器错误",
				})
			}

			// 检查用户是否已是成员
			for _, memberRecord := range allMemberRecords {
				if memberRecord.GetString("ledger") == ledgerId && memberRecord.GetString("user") == authRecord.Id {
					return c.JSON(http.StatusConflict, map[string]any{
						"code":    409,
						"message": "你已经是该账本的成员",
					})
				}
			}

			// 添加用户到账本成员
			memberRecord := core.NewRecord(ledgerMembersCollection)
			memberRecord.Set("ledger", ledgerId)
			memberRecord.Set("user", authRecord.Id)
			memberRecord.Set("role", "member")

			if err := app.Save(memberRecord); err != nil {
				return c.JSON(http.StatusInternalServerError, map[string]any{
					"code":    500,
					"message": "加入账本失败",
				})
			}

			// 增加邀请码使用次数
			foundRecord.Set("used_count", usedCount+1)
			if err := app.Save(foundRecord); err != nil {
				return c.JSON(http.StatusInternalServerError, map[string]any{
					"code":    500,
					"message": "更新邀请码状态失败",
				})
			}

			// 获取账本信息
			ledgerCollection, err := app.FindCollectionByNameOrId("ledgers")
			if err != nil {
				return c.JSON(http.StatusInternalServerError, map[string]any{
					"code":    500,
					"message": "内部服务器错误",
				})
			}

			ledger, err := app.FindRecordById(ledgerCollection, ledgerId)
			if err != nil {
				return c.JSON(http.StatusNotFound, map[string]any{
					"code":    404,
					"message": "账本不存在",
				})
			}

			// 返回成功
			return c.JSON(http.StatusOK, map[string]any{
				"success":    true,
				"ledgerId":   ledgerId,
				"ledgerName": ledger.GetString("name"),
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
