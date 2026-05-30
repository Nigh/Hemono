package main

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	mathrand "math/rand"
	"net/http"
	"os"
	"regexp"
	"sort"
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
		letterPart.WriteByte(letters[mathrand.Intn(len(letters))])
	}

	var numberPart strings.Builder
	for i := 0; i < 6; i++ {
		numberPart.WriteByte(byte('0' + mathrand.Intn(10)))
	}

	return letterPart.String() + "-" + numberPart.String()
}

func generateToken() (string, string, string, error) {
	bytes := make([]byte, 20)
	if _, err := rand.Read(bytes); err != nil {
		return "", "", "", err
	}
	token := "hmn_" + hex.EncodeToString(bytes)
	hash := sha256.Sum256([]byte(token))
	tokenHash := hex.EncodeToString(hash[:])
	tokenPrefix := token[:12]
	return token, tokenHash, tokenPrefix, nil
}

func authenticateRequest(app core.App, c *core.RequestEvent) (*core.Record, error) {
	authHeader := c.Request.Header.Get("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		return nil, fmt.Errorf("missing or invalid Authorization header")
	}
	token := strings.TrimPrefix(authHeader, "Bearer ")
	if len(token) < 12 {
		return nil, fmt.Errorf("invalid token")
	}
	prefix := token[:12]

	tokens, err := app.FindRecordsByFilter(
		"api_tokens",
		"token_prefix = {:prefix}",
		"",
		0,
		0,
		map[string]any{"prefix": prefix},
	)
	if err != nil || len(tokens) == 0 {
		return nil, fmt.Errorf("invalid token")
	}

	hash := sha256.Sum256([]byte(token))
	tokenHash := hex.EncodeToString(hash[:])

	for _, t := range tokens {
		if t.GetString("token_hash") == tokenHash {
			user, err := app.FindRecordById("users", t.GetString("user"))
			if err != nil {
				return nil, fmt.Errorf("user not found")
			}
			return user, nil
		}
	}

	return nil, fmt.Errorf("invalid token")
}

func main() {
	app := pocketbase.New()

	migratecmd.MustRegister(app, app.RootCmd, migratecmd.Config{
		Dir: "migrations",
	})

	// 使用 OnBootstrap 钩子，它在应用初始化时（DB 连接后，Server 启动前）执行
	app.OnServe().BindFunc(func(e *core.ServeEvent) error {
		email := os.Getenv("PB_ADMIN_EMAIL")
		password := os.Getenv("PB_ADMIN_PASSWORD")

		if email != "" && password != "" {
			collection, err := app.FindCollectionByNameOrId(core.CollectionNameSuperusers)
			if err != nil {
				return err
			}

			// 检查管理员是否已存在
			admin, _ := app.FindFirstRecordByData(collection.Id, "email", email)

			if admin == nil {
				log.Printf("正在初始化超级用户: %s", email)

				// 创建新的超级用户记录
				newAdmin := core.NewRecord(collection)
				newAdmin.SetEmail(email)
				newAdmin.SetPassword(password)

				// 保存记录
				if err := app.Save(newAdmin); err != nil {
					log.Printf("超级用户创建失败: %v", err)
				} else {
					log.Println("超级用户初始化成功！")
				}
			}
		}
		return e.Next() // 继续执行初始化链
	})

	// Token 管理 API 路由
	app.OnServe().BindFunc(func(e *core.ServeEvent) error {
		// POST /api/tokens - 创建 token
		e.Router.POST("/api/tokens", func(c *core.RequestEvent) error {
			authRecord := c.Auth
			if authRecord == nil {
				return c.JSON(http.StatusUnauthorized, map[string]any{
					"code":    401,
					"message": "未认证",
				})
			}

			var req struct {
				Name string `json:"name"`
			}
			if err := json.NewDecoder(c.Request.Body).Decode(&req); err != nil || req.Name == "" {
				return c.JSON(http.StatusBadRequest, map[string]any{
					"code":    400,
					"message": "名称不能为空",
				})
			}

			existing, err := app.FindRecordsByFilter(
				"api_tokens",
				"user = {:user}",
				"",
				0,
				0,
				map[string]any{"user": authRecord.Id},
			)
			if err != nil {
				return c.JSON(http.StatusInternalServerError, map[string]any{
					"code":    500,
					"message": "查询 token 失败",
				})
			}
			if len(existing) >= 3 {
				return c.JSON(http.StatusConflict, map[string]any{
					"code":    409,
					"message": "最多创建 3 个 token",
				})
			}

			token, tokenHash, tokenPrefix, err := generateToken()
			if err != nil {
				return c.JSON(http.StatusInternalServerError, map[string]any{
					"code":    500,
					"message": "生成 token 失败",
				})
			}

			collection, err := app.FindCollectionByNameOrId("api_tokens")
			if err != nil {
				return c.JSON(http.StatusInternalServerError, map[string]any{
					"code":    500,
					"message": "内部服务器错误",
				})
			}

			record := core.NewRecord(collection)
			record.Set("user", authRecord.Id)
			record.Set("name", req.Name)
			record.Set("token_hash", tokenHash)
			record.Set("token_prefix", tokenPrefix)

			if err := app.Save(record); err != nil {
				return c.JSON(http.StatusInternalServerError, map[string]any{
					"code":    500,
					"message": "保存 token 失败",
				})
			}

			return c.JSON(http.StatusOK, map[string]any{
				"id":           record.Id,
				"name":         record.GetString("name"),
				"token":        token,
				"token_prefix": tokenPrefix,
				"created":      record.GetDateTime("created").Time().Format(time.RFC3339),
			})
		})

		// GET /api/tokens - 列出当前用户的 tokens
		e.Router.GET("/api/tokens", func(c *core.RequestEvent) error {
			authRecord := c.Auth
			if authRecord == nil {
				return c.JSON(http.StatusUnauthorized, map[string]any{
					"code":    401,
					"message": "未认证",
				})
			}

			tokens, err := app.FindRecordsByFilter(
				"api_tokens",
				"user = {:user}",
				"-created",
				0,
				0,
				map[string]any{"user": authRecord.Id},
			)
			if err != nil {
				return c.JSON(http.StatusInternalServerError, map[string]any{
					"code":    500,
					"message": "查询 token 失败",
				})
			}

			var result []map[string]any
			for _, t := range tokens {
				result = append(result, map[string]any{
					"id":           t.Id,
					"name":         t.GetString("name"),
					"token_prefix": t.GetString("token_prefix"),
					"created":      t.GetDateTime("created").Time().Format(time.RFC3339),
				})
			}

			return c.JSON(http.StatusOK, result)
		})

		// DELETE /api/tokens/{id} - 删除 token
		e.Router.DELETE("/api/tokens/{id}", func(c *core.RequestEvent) error {
			authRecord := c.Auth
			if authRecord == nil {
				return c.JSON(http.StatusUnauthorized, map[string]any{
					"code":    401,
					"message": "未认证",
				})
			}

			tokenId := c.Request.PathValue("id")
			record, err := app.FindRecordById("api_tokens", tokenId)
			if err != nil {
				return c.JSON(http.StatusNotFound, map[string]any{
					"code":    404,
					"message": "Token 不存在",
				})
			}

			if record.GetString("user") != authRecord.Id {
				return c.JSON(http.StatusForbidden, map[string]any{
					"code":    403,
					"message": "无权删除该 Token",
				})
			}

			if err := app.Delete(record); err != nil {
				return c.JSON(http.StatusInternalServerError, map[string]any{
					"code":    500,
					"message": "删除 Token 失败",
				})
			}

			return c.JSON(http.StatusOK, map[string]any{
				"success": true,
				"message": "Token 已删除",
			})
		})

		return e.Next()
	})

	// 业务功能 REST API（Token auth 认证）
	app.OnServe().BindFunc(func(e *core.ServeEvent) error {
		// GET /api/v1/ledgers
		e.Router.GET("/api/v1/ledgers", func(c *core.RequestEvent) error {
			user, err := authenticateRequest(app, c)
			if err != nil {
				return c.JSON(http.StatusUnauthorized, map[string]any{
					"code":    401,
					"message": "未认证",
				})
			}

			ownedLedgers, err := app.FindRecordsByFilter(
				"ledgers",
				"owner = {:user}",
				"-created",
				0,
				0,
				map[string]any{"user": user.Id},
			)
			if err != nil {
				return c.JSON(http.StatusInternalServerError, map[string]any{
					"code":    500,
					"message": "查询账本失败",
				})
			}

			memberships, err := app.FindRecordsByFilter(
				"ledger_members",
				"user = {:user}",
				"",
				0,
				0,
				map[string]any{"user": user.Id},
			)
			if err != nil {
				return c.JSON(http.StatusInternalServerError, map[string]any{
					"code":    500,
					"message": "查询成员关系失败",
				})
			}

			seen := make(map[string]bool)
			var result []map[string]any
			for _, l := range ownedLedgers {
				seen[l.Id] = true
				result = append(result, map[string]any{
					"id":      l.Id,
					"name":    l.GetString("name"),
					"owner":   l.GetString("owner"),
					"created": l.GetDateTime("created").Time().Format(time.RFC3339),
				})
			}
			for _, m := range memberships {
				ledgerId := m.GetString("ledger")
				if seen[ledgerId] {
					continue
				}
				seen[ledgerId] = true
				ledger, err := app.FindRecordById("ledgers", ledgerId)
				if err != nil {
					continue
				}
				result = append(result, map[string]any{
					"id":      ledger.Id,
					"name":    ledger.GetString("name"),
					"owner":   ledger.GetString("owner"),
					"created": ledger.GetDateTime("created").Time().Format(time.RFC3339),
				})
			}

			return c.JSON(http.StatusOK, result)
		})

		// POST /api/v1/ledgers
		e.Router.POST("/api/v1/ledgers", func(c *core.RequestEvent) error {
			user, err := authenticateRequest(app, c)
			if err != nil {
				return c.JSON(http.StatusUnauthorized, map[string]any{
					"code":    401,
					"message": "未认证",
				})
			}

			var req struct {
				Name string `json:"name"`
			}
			if err := json.NewDecoder(c.Request.Body).Decode(&req); err != nil || req.Name == "" {
				return c.JSON(http.StatusBadRequest, map[string]any{
					"code":    400,
					"message": "账本名称不能为空",
				})
			}

			collection, err := app.FindCollectionByNameOrId("ledgers")
			if err != nil {
				return c.JSON(http.StatusInternalServerError, map[string]any{
					"code":    500,
					"message": "内部服务器错误",
				})
			}

			record := core.NewRecord(collection)
			record.Set("name", req.Name)
			record.Set("owner", user.Id)

			if err := app.Save(record); err != nil {
				return c.JSON(http.StatusInternalServerError, map[string]any{
					"code":    500,
					"message": "创建账本失败",
				})
			}

			return c.JSON(http.StatusOK, map[string]any{
				"id":      record.Id,
				"name":    record.GetString("name"),
				"owner":   record.GetString("owner"),
				"created": record.GetDateTime("created").Time().Format(time.RFC3339),
			})
		})

		// GET /api/v1/ledgers/{id}
		e.Router.GET("/api/v1/ledgers/{id}", func(c *core.RequestEvent) error {
			user, err := authenticateRequest(app, c)
			if err != nil {
				return c.JSON(http.StatusUnauthorized, map[string]any{
					"code":    401,
					"message": "未认证",
				})
			}

			ledgerId := c.Request.PathValue("id")
			ledger, err := app.FindRecordById("ledgers", ledgerId)
			if err != nil {
				return c.JSON(http.StatusNotFound, map[string]any{
					"code":    404,
					"message": "账本不存在",
				})
			}

			isOwner := ledger.GetString("owner") == user.Id
			if !isOwner {
				member, _ := app.FindFirstRecordByFilter(
					"ledger_members",
					"ledger = {:ledger} && user = {:user}",
					map[string]any{"ledger": ledgerId, "user": user.Id},
				)
				if member == nil {
					return c.JSON(http.StatusForbidden, map[string]any{
						"code":    403,
						"message": "无权访问该账本",
					})
				}
			}

			return c.JSON(http.StatusOK, map[string]any{
				"id":      ledger.Id,
				"name":    ledger.GetString("name"),
				"owner":   ledger.GetString("owner"),
				"created": ledger.GetDateTime("created").Time().Format(time.RFC3339),
			})
		})

		// DELETE /api/v1/ledgers/{id}
		e.Router.DELETE("/api/v1/ledgers/{id}", func(c *core.RequestEvent) error {
			user, err := authenticateRequest(app, c)
			if err != nil {
				return c.JSON(http.StatusUnauthorized, map[string]any{
					"code":    401,
					"message": "未认证",
				})
			}

			ledgerId := c.Request.PathValue("id")
			ledger, err := app.FindRecordById("ledgers", ledgerId)
			if err != nil {
				return c.JSON(http.StatusNotFound, map[string]any{
					"code":    404,
					"message": "账本不存在",
				})
			}

			if ledger.GetString("owner") != user.Id {
				return c.JSON(http.StatusForbidden, map[string]any{
					"code":    403,
					"message": "只有账本所有者可以删除账本",
				})
			}

			if err := app.Delete(ledger); err != nil {
				return c.JSON(http.StatusInternalServerError, map[string]any{
					"code":    500,
					"message": "删除账本失败",
				})
			}

			return c.JSON(http.StatusOK, map[string]any{
				"success": true,
				"message": "账本已删除",
			})
		})

		// GET /api/v1/ledgers/{id}/members
		e.Router.GET("/api/v1/ledgers/{id}/members", func(c *core.RequestEvent) error {
			user, err := authenticateRequest(app, c)
			if err != nil {
				return c.JSON(http.StatusUnauthorized, map[string]any{
					"code":    401,
					"message": "未认证",
				})
			}

			ledgerId := c.Request.PathValue("id")
			ledger, err := app.FindRecordById("ledgers", ledgerId)
			if err != nil {
				return c.JSON(http.StatusNotFound, map[string]any{
					"code":    404,
					"message": "账本不存在",
				})
			}

			isOwner := ledger.GetString("owner") == user.Id
			if !isOwner {
				member, _ := app.FindFirstRecordByFilter(
					"ledger_members",
					"ledger = {:ledger} && user = {:user}",
					map[string]any{"ledger": ledgerId, "user": user.Id},
				)
				if member == nil {
					return c.JSON(http.StatusForbidden, map[string]any{
						"code":    403,
						"message": "无权访问该账本",
					})
				}
			}

			members, err := app.FindRecordsByFilter(
				"ledger_members",
				"ledger = {:ledger}",
				"",
				0,
				0,
				map[string]any{"ledger": ledgerId},
			)
			if err != nil {
				return c.JSON(http.StatusInternalServerError, map[string]any{
					"code":    500,
					"message": "获取成员失败",
				})
			}

			var result []map[string]any
			for _, m := range members {
				u, err := app.FindRecordById("users", m.GetString("user"))
				if err != nil {
					continue
				}
				result = append(result, map[string]any{
					"id":     u.Id,
					"name":   u.GetString("name"),
					"email":  u.GetString("email"),
					"avatar": u.GetString("avatar"),
				})
			}

			return c.JSON(http.StatusOK, result)
		})

		// GET /api/v1/ledgers/{id}/transactions
		e.Router.GET("/api/v1/ledgers/{id}/transactions", func(c *core.RequestEvent) error {
			user, err := authenticateRequest(app, c)
			if err != nil {
				return c.JSON(http.StatusUnauthorized, map[string]any{
					"code":    401,
					"message": "未认证",
				})
			}

			ledgerId := c.Request.PathValue("id")
			ledger, err := app.FindRecordById("ledgers", ledgerId)
			if err != nil {
				return c.JSON(http.StatusNotFound, map[string]any{
					"code":    404,
					"message": "账本不存在",
				})
			}

			isOwner := ledger.GetString("owner") == user.Id
			if !isOwner {
				member, _ := app.FindFirstRecordByFilter(
					"ledger_members",
					"ledger = {:ledger} && user = {:user}",
					map[string]any{"ledger": ledgerId, "user": user.Id},
				)
				if member == nil {
					return c.JSON(http.StatusForbidden, map[string]any{
						"code":    403,
						"message": "无权访问该账本",
					})
				}
			}

			transactions, err := app.FindRecordsByFilter(
				"transactions",
				fmt.Sprintf("ledger = \"%s\"", ledgerId),
				"-date,-created",
				0,
				0,
				nil,
			)
			if err != nil {
				return c.JSON(http.StatusInternalServerError, map[string]any{
					"code":    500,
					"message": "获取交易失败",
				})
			}

			var result []map[string]any
			for _, tx := range transactions {
				payer, _ := app.FindRecordById("users", tx.GetString("payer"))
				payerName := ""
				payerAvatar := ""
				if payer != nil {
					payerName = payer.GetString("name")
					payerAvatar = payer.GetString("avatar")
				}

				item := map[string]any{
					"id":          tx.Id,
					"ledger":      tx.GetString("ledger"),
					"payer":       tx.GetString("payer"),
					"payerName":   payerName,
					"payerAvatar": payerAvatar,
					"amount":      tx.GetInt("amount"),
					"type":        tx.GetString("type"),
					"direction":   tx.GetString("direction"),
					"note":        tx.GetString("note"),
					"date":        tx.GetDateTime("date").Time().Format("2006-01-02"),
					"created":     tx.GetDateTime("created").Time().Format(time.RFC3339),
				}
				if tx.GetString("type") == "SINGLE" {
					item["beneficiary"] = tx.GetString("beneficiary")
				}
				result = append(result, item)
			}

			return c.JSON(http.StatusOK, result)
		})

		// POST /api/v1/ledgers/{id}/transactions
		e.Router.POST("/api/v1/ledgers/{id}/transactions", func(c *core.RequestEvent) error {
			user, err := authenticateRequest(app, c)
			if err != nil {
				return c.JSON(http.StatusUnauthorized, map[string]any{
					"code":    401,
					"message": "未认证",
				})
			}

			ledgerId := c.Request.PathValue("id")
			ledger, err := app.FindRecordById("ledgers", ledgerId)
			if err != nil {
				return c.JSON(http.StatusNotFound, map[string]any{
					"code":    404,
					"message": "账本不存在",
				})
			}

			isOwner := ledger.GetString("owner") == user.Id
			if !isOwner {
				member, _ := app.FindFirstRecordByFilter(
					"ledger_members",
					"ledger = {:ledger} && user = {:user}",
					map[string]any{"ledger": ledgerId, "user": user.Id},
				)
				if member == nil {
					return c.JSON(http.StatusForbidden, map[string]any{
						"code":    403,
						"message": "无权访问该账本",
					})
				}
			}

			var req struct {
				Amount      int    `json:"amount"`
				Type        string `json:"type"`
				Direction   string `json:"direction"`
				Beneficiary string `json:"beneficiary"`
				Note        string `json:"note"`
				Date        string `json:"date"`
			}
			if err := json.NewDecoder(c.Request.Body).Decode(&req); err != nil {
				return c.JSON(http.StatusBadRequest, map[string]any{
					"code":    400,
					"message": "请求参数错误",
				})
			}

			if req.Amount <= 0 {
				return c.JSON(http.StatusBadRequest, map[string]any{
					"code":    400,
					"message": "金额必须大于 0",
				})
			}
			if req.Type != "AA" && req.Type != "SINGLE" {
				return c.JSON(http.StatusBadRequest, map[string]any{
					"code":    400,
					"message": "类型必须为 AA 或 SINGLE",
				})
			}
			if req.Direction != "EXPENSE" && req.Direction != "INCOME" {
				return c.JSON(http.StatusBadRequest, map[string]any{
					"code":    400,
					"message": "方向必须为 EXPENSE 或 INCOME",
				})
			}
			if req.Type == "SINGLE" && req.Beneficiary == "" {
				return c.JSON(http.StatusBadRequest, map[string]any{
					"code":    400,
					"message": "SINGLE 类型必须指定受益人",
				})
			}

			txDate := time.Now().Format("2006-01-02")
			if req.Date != "" {
				matched, _ := regexp.MatchString("^\\d{4}-\\d{2}-\\d{2}$", req.Date)
				if !matched {
					return c.JSON(http.StatusBadRequest, map[string]any{
						"code":    400,
						"message": "日期格式不正确，应为 YYYY-MM-DD",
					})
				}
				txDate = req.Date
			}

			collection, err := app.FindCollectionByNameOrId("transactions")
			if err != nil {
				return c.JSON(http.StatusInternalServerError, map[string]any{
					"code":    500,
					"message": "内部服务器错误",
				})
			}

			record := core.NewRecord(collection)
			record.Set("ledger", ledgerId)
			record.Set("payer", user.Id)
			record.Set("amount", req.Amount)
			record.Set("type", req.Type)
			record.Set("direction", req.Direction)
			record.Set("note", req.Note)
			record.Set("date", txDate)
			if req.Type == "SINGLE" {
				record.Set("beneficiary", req.Beneficiary)
			}

			if err := app.Save(record); err != nil {
				return c.JSON(http.StatusInternalServerError, map[string]any{
					"code":    500,
					"message": "创建交易失败",
				})
			}

			return c.JSON(http.StatusOK, map[string]any{
				"id":        record.Id,
				"ledger":    record.GetString("ledger"),
				"payer":     record.GetString("payer"),
				"amount":    record.GetInt("amount"),
				"type":      record.GetString("type"),
				"direction": record.GetString("direction"),
				"note":      record.GetString("note"),
				"date":      txDate,
			})
		})

		// DELETE /api/v1/transactions/{id}
		e.Router.DELETE("/api/v1/transactions/{id}", func(c *core.RequestEvent) error {
			user, err := authenticateRequest(app, c)
			if err != nil {
				return c.JSON(http.StatusUnauthorized, map[string]any{
					"code":    401,
					"message": "未认证",
				})
			}

			txId := c.Request.PathValue("id")
			record, err := app.FindRecordById("transactions", txId)
			if err != nil {
				return c.JSON(http.StatusNotFound, map[string]any{
					"code":    404,
					"message": "交易不存在",
				})
			}

			if record.GetString("payer") != user.Id {
				return c.JSON(http.StatusForbidden, map[string]any{
					"code":    403,
					"message": "只有付款人可以删除交易",
				})
			}

			if err := app.Delete(record); err != nil {
				return c.JSON(http.StatusInternalServerError, map[string]any{
					"code":    500,
					"message": "删除交易失败",
				})
			}

			return c.JSON(http.StatusOK, map[string]any{
				"success": true,
				"message": "交易已删除",
			})
		})

		// GET /api/v1/ledgers/{id}/stats
		e.Router.GET("/api/v1/ledgers/{id}/stats", func(c *core.RequestEvent) error {
			user, err := authenticateRequest(app, c)
			if err != nil {
				return c.JSON(http.StatusUnauthorized, map[string]any{
					"code":    401,
					"message": "未认证",
				})
			}

			ledgerId := c.Request.PathValue("id")
			ledger, err := app.FindRecordById("ledgers", ledgerId)
			if err != nil {
				return c.JSON(http.StatusNotFound, map[string]any{
					"code":    404,
					"message": "账本不存在",
				})
			}

			isOwner := ledger.GetString("owner") == user.Id
			if !isOwner {
				member, _ := app.FindFirstRecordByFilter(
					"ledger_members",
					"ledger = {:ledger} && user = {:user}",
					map[string]any{"ledger": ledgerId, "user": user.Id},
				)
				if member == nil {
					return c.JSON(http.StatusForbidden, map[string]any{
						"code":    403,
						"message": "无权访问该账本",
					})
				}
			}

			month := c.Request.URL.Query().Get("month")
			var dateFilter string
			if month != "" {
				matched, _ := regexp.MatchString("^\\d{4}-\\d{2}$", month)
				if !matched {
					return c.JSON(http.StatusBadRequest, map[string]any{
						"code":    400,
						"message": "月份格式不正确，应为 YYYY-MM",
					})
				}
				parsedTime, _ := time.Parse("2006-01", month)
				nextMonth := parsedTime.AddDate(0, 1, 0).Format("2006-01")
				dateFilter = fmt.Sprintf(" && date >= \"%s-01\" && date < \"%s-01\"", month, nextMonth)
			}

			members, err := app.FindRecordsByFilter(
				"ledger_members",
				"ledger = {:ledger}",
				"",
				0,
				0,
				map[string]any{"ledger": ledgerId},
			)
			if err != nil {
				return c.JSON(http.StatusInternalServerError, map[string]any{
					"code":    500,
					"message": "获取成员失败",
				})
			}

			memberMap := make(map[string]*struct {
				userId       string
				name         string
				email        string
				avatar       string
				totalExpense int
				totalBenefit int
				totalIncome  int
				incomeShare  int
			})

			for _, member := range members {
				userId := member.GetString("user")
				u, err := app.FindRecordById("users", userId)
				if err != nil {
					continue
				}
				memberMap[userId] = &struct {
					userId       string
					name         string
					email        string
					avatar       string
					totalExpense int
					totalBenefit int
					totalIncome  int
					incomeShare  int
				}{
					userId:       userId,
					name:         u.GetString("name"),
					email:        u.GetString("email"),
					avatar:       u.GetString("avatar"),
					totalExpense: 0,
					totalBenefit: 0,
					totalIncome:  0,
					incomeShare:  0,
				}
			}

			transactions, err := app.FindRecordsByFilter(
				"transactions",
				fmt.Sprintf("ledger = \"%s\"%s", ledgerId, dateFilter),
				"",
				0,
				0,
				nil,
			)
			if err != nil {
				return c.JSON(http.StatusInternalServerError, map[string]any{
					"code":    500,
					"message": "获取交易失败",
				})
			}

			totalExpense := 0
			totalBenefit := 0
			totalIncome := 0
			totalIncomeShare := 0
			monthlyExpense := 0
			monthlyIncome := 0
			last7DaysExpense := 0
			last7DaysIncome := 0

			now := time.Now()
			monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
			nextMonthStart := monthStart.AddDate(0, 1, 0)
			last7DaysStart := now.AddDate(0, 0, -6)

			for _, tx := range transactions {
				amount := tx.GetInt("amount")
				payer := tx.GetString("payer")
				txType := tx.GetString("type")
				beneficiary := tx.GetString("beneficiary")
				direction := tx.GetString("direction")
				txDate := tx.GetDateTime("date").Time()

				if direction == "INCOME" {
					if member, ok := memberMap[payer]; ok {
						member.totalIncome += amount
						totalIncome += amount
						if !txDate.Before(monthStart) && txDate.Before(nextMonthStart) {
							monthlyIncome += amount
						}
						if !txDate.Before(last7DaysStart) && !txDate.After(now) {
							last7DaysIncome += amount
						}
					}
					if txType == "AA" {
						if len(members) > 0 {
							base := amount / len(members)
							remainder := amount % len(members)
							sort.Slice(members, func(i, j int) bool {
								return members[i].GetString("user") < members[j].GetString("user")
							})
							for i, member := range members {
								userId := member.GetString("user")
								share := base
								if i < remainder {
									share += 1
								}
								if m, ok := memberMap[userId]; ok {
									m.incomeShare += share
									totalIncomeShare += share
								}
							}
						}
					} else if txType == "SINGLE" && beneficiary != "" {
						if m, ok := memberMap[beneficiary]; ok {
							m.incomeShare += amount
							totalIncomeShare += amount
						}
					}
				} else {
					if member, ok := memberMap[payer]; ok {
						member.totalExpense += amount
						totalExpense += amount
						if !txDate.Before(monthStart) && txDate.Before(nextMonthStart) {
							monthlyExpense += amount
						}
						if !txDate.Before(last7DaysStart) && !txDate.After(now) {
							last7DaysExpense += amount
						}
					}
					if txType == "AA" {
						if len(members) > 0 {
							base := amount / len(members)
							remainder := amount % len(members)
							sort.Slice(members, func(i, j int) bool {
								return members[i].GetString("user") < members[j].GetString("user")
							})
							for i, member := range members {
								userId := member.GetString("user")
								benefit := base
								if i < remainder {
									benefit += 1
								}
								if m, ok := memberMap[userId]; ok {
									m.totalBenefit += benefit
									totalBenefit += benefit
								}
							}
						}
					} else if txType == "SINGLE" && beneficiary != "" {
						if m, ok := memberMap[beneficiary]; ok {
							m.totalBenefit += amount
							totalBenefit += amount
						}
					}
				}
			}

			type MemberStat struct {
				UserId       string `json:"userId"`
				Name         string `json:"name"`
				Email        string `json:"email"`
				Avatar       string `json:"avatar"`
				TotalExpense int    `json:"totalExpense"`
				TotalBenefit int    `json:"totalBenefit"`
				TotalIncome  int    `json:"totalIncome"`
				IncomeShare  int    `json:"incomeShare"`
				Balance      int    `json:"balance"`
				Percentage   int    `json:"percentage"`
			}

			var memberStats []MemberStat
			maxBalance := 0

			for _, member := range memberMap {
				balance := member.totalExpense - member.totalIncome - member.totalBenefit + member.incomeShare
				balanceAbs := balance
				if balanceAbs < 0 {
					balanceAbs = -balanceAbs
				}
				if balanceAbs > maxBalance {
					maxBalance = balanceAbs
				}
				memberStats = append(memberStats, MemberStat{
					UserId:       member.userId,
					Name:         member.name,
					Email:        member.email,
					Avatar:       member.avatar,
					TotalExpense: member.totalExpense,
					TotalBenefit: member.totalBenefit,
					TotalIncome:  member.totalIncome,
					IncomeShare:  member.incomeShare,
					Balance:      balance,
				})
			}

			for i := range memberStats {
				balanceAbs := memberStats[i].Balance
				if balanceAbs < 0 {
					balanceAbs = -balanceAbs
				}
				if maxBalance > 0 {
					memberStats[i].Percentage = int(float64(balanceAbs) / float64(maxBalance) * 100)
				} else {
					memberStats[i].Percentage = 0
				}
			}

			sort.Slice(memberStats, func(i, j int) bool {
				balanceI := memberStats[i].Balance
				balanceJ := memberStats[j].Balance
				if balanceI < 0 {
					balanceI = -balanceI
				}
				if balanceJ < 0 {
					balanceJ = -balanceJ
				}
				return balanceI > balanceJ
			})

			return c.JSON(http.StatusOK, map[string]any{
				"totalExpense":     totalExpense,
				"totalBenefit":     totalBenefit,
				"totalIncome":      totalIncome,
				"totalIncomeShare": totalIncomeShare,
				"monthlyExpense":   monthlyExpense,
				"monthlyIncome":    monthlyIncome,
				"last7DaysExpense": last7DaysExpense,
				"last7DaysIncome":  last7DaysIncome,
				"memberStats":      memberStats,
			})
		})

		// GET /api/v1/ledgers/{id}/invitation
		e.Router.GET("/api/v1/ledgers/{id}/invitation", func(c *core.RequestEvent) error {
			user, err := authenticateRequest(app, c)
			if err != nil {
				return c.JSON(http.StatusUnauthorized, map[string]any{
					"code":    401,
					"message": "未认证",
				})
			}

			ledgerId := c.Request.PathValue("id")
			ledger, err := app.FindRecordById("ledgers", ledgerId)
			if err != nil {
				return c.JSON(http.StatusNotFound, map[string]any{
					"code":    404,
					"message": "账本不存在",
				})
			}

			if ledger.GetString("owner") != user.Id {
				return c.JSON(http.StatusForbidden, map[string]any{
					"code":    403,
					"message": "只有账本所有者可以查看邀请码",
				})
			}

			foundRecord, err := app.FindFirstRecordByFilter(
				"invitation_codes",
				"ledger = {:ledger} && used_count < max_uses && expires_at > {:now}",
				map[string]any{"ledger": ledgerId, "now": time.Now().Format("2006-01-02 15:04:05")},
			)
			if err != nil || foundRecord == nil {
				return c.JSON(http.StatusOK, map[string]any{
					"exists": false,
				})
			}

			expiresAt := foundRecord.GetDateTime("expires_at").Time()
			isExpired := time.Now().After(expiresAt) || foundRecord.GetInt("used_count") >= foundRecord.GetInt("max_uses")

			return c.JSON(http.StatusOK, map[string]any{
				"exists":    !isExpired,
				"id":        foundRecord.Id,
				"code":      foundRecord.GetString("code"),
				"expiresAt": expiresAt.Format("2006-01-02 15:04"),
				"maxUses":   foundRecord.GetInt("max_uses"),
				"usedCount": foundRecord.GetInt("used_count"),
			})
		})

		// POST /api/v1/ledgers/{id}/invitation
		e.Router.POST("/api/v1/ledgers/{id}/invitation", func(c *core.RequestEvent) error {
			user, err := authenticateRequest(app, c)
			if err != nil {
				return c.JSON(http.StatusUnauthorized, map[string]any{
					"code":    401,
					"message": "未认证",
				})
			}

			ledgerId := c.Request.PathValue("id")
			ledger, err := app.FindRecordById("ledgers", ledgerId)
			if err != nil {
				return c.JSON(http.StatusNotFound, map[string]any{
					"code":    404,
					"message": "账本不存在",
				})
			}

			if ledger.GetString("owner") != user.Id {
				return c.JSON(http.StatusForbidden, map[string]any{
					"code":    403,
					"message": "只有账本所有者可以生成邀请码",
				})
			}

			var req struct {
				MaxUses int `json:"max_uses"`
			}
			if err := json.NewDecoder(c.Request.Body).Decode(&req); err != nil {
				req.MaxUses = 10
			}
			if req.MaxUses < 1 || req.MaxUses > 99 {
				req.MaxUses = 10
			}

			existingRecord, _ := app.FindFirstRecordByFilter(
				"invitation_codes",
				"ledger = {:ledger} && used_count < max_uses && expires_at > {:now}",
				map[string]any{"ledger": ledgerId, "now": time.Now().Format("2006-01-02 15:04:05")},
			)
			if existingRecord != nil {
				expiresAt := existingRecord.GetDateTime("expires_at").Time()
				if time.Now().Before(expiresAt) && existingRecord.GetInt("used_count") < existingRecord.GetInt("max_uses") {
					return c.JSON(http.StatusConflict, map[string]any{
						"code":    409,
						"message": "该账本已存在有效邀请码",
					})
				}
			}

			code := generateInvitationCode()

			invitationCollection, err := app.FindCollectionByNameOrId("invitation_codes")
			if err != nil {
				return c.JSON(http.StatusInternalServerError, map[string]any{
					"code":    500,
					"message": "内部服务器错误",
				})
			}

			record := core.NewRecord(invitationCollection)
			record.Set("code", code)
			record.Set("ledger", ledgerId)
			record.Set("created_by", user.Id)
			record.Set("expires_at", time.Now().Add(24*time.Hour).Format("2006-01-02 15:04:05"))
			record.Set("max_uses", req.MaxUses)
			record.Set("used_count", 0)

			if err := app.Save(record); err != nil {
				return c.JSON(http.StatusInternalServerError, map[string]any{
					"code":    500,
					"message": "保存邀请码失败",
				})
			}

			return c.JSON(http.StatusOK, map[string]any{
				"code":       code,
				"expires_at": time.Now().Add(24 * time.Hour).Format("2006-01-02 15:04"),
				"max_uses":   req.MaxUses,
				"used_count": 0,
			})
		})

		// DELETE /api/v1/invitations/{id}
		e.Router.DELETE("/api/v1/invitations/{id}", func(c *core.RequestEvent) error {
			user, err := authenticateRequest(app, c)
			if err != nil {
				return c.JSON(http.StatusUnauthorized, map[string]any{
					"code":    401,
					"message": "未认证",
				})
			}

			invitationId := c.Request.PathValue("id")
			foundRecord, err := app.FindRecordById("invitation_codes", invitationId)
			if err != nil {
				return c.JSON(http.StatusNotFound, map[string]any{
					"code":    404,
					"message": "邀请码不存在",
				})
			}

			ledgerId := foundRecord.GetString("ledger")
			ledger, err := app.FindRecordById("ledgers", ledgerId)
			if err != nil {
				return c.JSON(http.StatusNotFound, map[string]any{
					"code":    404,
					"message": "账本不存在",
				})
			}

			if ledger.GetString("owner") != user.Id {
				return c.JSON(http.StatusForbidden, map[string]any{
					"code":    403,
					"message": "只有账本所有者可以删除邀请码",
				})
			}

			if err := app.Delete(foundRecord); err != nil {
				return c.JSON(http.StatusInternalServerError, map[string]any{
					"code":    500,
					"message": "删除邀请码失败",
				})
			}

			return c.JSON(http.StatusOK, map[string]any{
				"success": true,
				"message": "邀请码已删除",
			})
		})

		// POST /api/v1/invitations/join
		e.Router.POST("/api/v1/invitations/join", func(c *core.RequestEvent) error {
			user, err := authenticateRequest(app, c)
			if err != nil {
				return c.JSON(http.StatusUnauthorized, map[string]any{
					"code":    401,
					"message": "未认证",
				})
			}

			var req struct {
				Code string `json:"code"`
			}
			if err := json.NewDecoder(c.Request.Body).Decode(&req); err != nil || req.Code == "" {
				return c.JSON(http.StatusBadRequest, map[string]any{
					"code":    400,
					"message": "邀请码不能为空",
				})
			}

			foundRecord, err := app.FindFirstRecordByFilter("invitation_codes", "code = {:code}", map[string]any{"code": req.Code})
			if err != nil {
				return c.JSON(http.StatusNotFound, map[string]any{
					"code":    404,
					"message": "邀请码不存在",
				})
			}

			expiresAt := foundRecord.GetDateTime("expires_at").Time()
			if time.Now().After(expiresAt) {
				app.Delete(foundRecord)
				return c.JSON(http.StatusGone, map[string]any{
					"code":    410,
					"message": "邀请码已过期",
				})
			}

			if foundRecord.GetInt("used_count") >= foundRecord.GetInt("max_uses") {
				app.Delete(foundRecord)
				return c.JSON(http.StatusConflict, map[string]any{
					"code":    409,
					"message": "邀请码已被使用完",
				})
			}

			ledgerId := foundRecord.GetString("ledger")

			existingMember, _ := app.FindFirstRecordByFilter(
				"ledger_members",
				"ledger = {:ledger} && user = {:user}",
				map[string]any{"ledger": ledgerId, "user": user.Id},
			)
			if existingMember != nil {
				return c.JSON(http.StatusConflict, map[string]any{
					"code":    409,
					"message": "你已经是该账本的成员",
				})
			}

			var ledgerName string
			err = app.RunInTransaction(func(txApp core.App) error {
				memberCollection, err := txApp.FindCollectionByNameOrId("ledger_members")
				if err != nil {
					return err
				}
				newMember := core.NewRecord(memberCollection)
				newMember.Set("ledger", ledgerId)
				newMember.Set("user", user.Id)
				newMember.Set("role", "member")
				if err := txApp.Save(newMember); err != nil {
					return err
				}

				newUsedCount := foundRecord.GetInt("used_count") + 1
				foundRecord.Set("used_count", newUsedCount)
				if newUsedCount >= foundRecord.GetInt("max_uses") {
					return txApp.Delete(foundRecord)
				}
				return txApp.Save(foundRecord)
			})
			if err != nil {
				return c.JSON(http.StatusInternalServerError, map[string]any{
					"code":    500,
					"message": "加入账本失败",
				})
			}

			ledger, _ := app.FindRecordById("ledgers", ledgerId)
			if ledger != nil {
				ledgerName = ledger.GetString("name")
			}

			return c.JSON(http.StatusOK, map[string]any{
				"success":    true,
				"ledgerId":   ledgerId,
				"ledgerName": ledgerName,
				"message":    "成功加入账本",
			})
		})

		return e.Next()
	})

	// 添加邀请码API路由
	app.OnServe().BindFunc(func(e *core.ServeEvent) error {
		// GET /api/invitations/by-ledger/{ledger_id} - 获取账本当前有效邀请码
		e.Router.GET("/api/invitations/by-ledger/{ledger_id}", func(c *core.RequestEvent) error {
			ledgerId := c.Request.PathValue("ledger_id")

			authRecord := c.Auth
			if authRecord == nil {
				return c.JSON(http.StatusUnauthorized, map[string]any{
					"code":    401,
					"message": "未认证",
				})
			}

			// 验证账本是否存在且用户是所有者
			ledger, err := app.FindRecordById("ledgers", ledgerId)
			if err != nil {
				return c.JSON(http.StatusNotFound, map[string]any{
					"code":    404,
					"message": "账本不存在",
				})
			}

			if ledger.GetString("owner") != authRecord.Id {
				return c.JSON(http.StatusForbidden, map[string]any{
					"code":    403,
					"message": "只有账本所有者可以查看邀请码",
				})
			}

			// 查找该账本的有效邀请码（未过期且未用完）
			foundRecord, err := app.FindFirstRecordByFilter(
				"invitation_codes",
				"ledger = {:ledger} && used_count < max_uses && expires_at > {:now}",
				map[string]any{"ledger": ledgerId, "now": time.Now().Format("2006-01-02 15:04:05")},
			)
			if err != nil || foundRecord == nil {
				// 不存在有效邀请码
				return c.JSON(http.StatusOK, map[string]any{
					"exists": false,
				})
			}

			expiresAt := foundRecord.GetDateTime("expires_at").Time()
			isExpired := time.Now().After(expiresAt) || foundRecord.GetInt("used_count") >= foundRecord.GetInt("max_uses")

			return c.JSON(http.StatusOK, map[string]any{
				"exists":    !isExpired,
				"id":        foundRecord.Id,
				"code":      foundRecord.GetString("code"),
				"expiresAt": expiresAt.Format("2006-01-02 15:04"),
				"maxUses":   foundRecord.GetInt("max_uses"),
				"usedCount": foundRecord.GetInt("used_count"),
			})
		})

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
			ledger, err := app.FindRecordById("ledgers", req.LedgerId)
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

			// 检查是否已存在有效邀请码
			existingRecord, _ := app.FindFirstRecordByFilter(
				"invitation_codes",
				"ledger = {:ledger} && used_count < max_uses && expires_at > {:now}",
				map[string]any{"ledger": req.LedgerId, "now": time.Now().Format("2006-01-02 15:04:05")},
			)
			if existingRecord != nil {
				// 检查是否已过期或已用完
				expiresAt := existingRecord.GetDateTime("expires_at").Time()
				if time.Now().Before(expiresAt) && existingRecord.GetInt("used_count") < existingRecord.GetInt("max_uses") {
					return c.JSON(http.StatusConflict, map[string]any{
						"code":    409,
						"message": "该账本已存在有效邀请码",
						"data": map[string]any{
							"id":        existingRecord.Id,
							"code":      existingRecord.GetString("code"),
							"expiresAt": expiresAt.Format("2006-01-02 15:04"),
							"maxUses":   existingRecord.GetInt("max_uses"),
							"usedCount": existingRecord.GetInt("used_count"),
						},
					})
				}
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
				"used_count": 0,
			})
		})

		// GET /api/invitations/by-code/{code}
		e.Router.GET("/api/invitations/by-code/{code}", func(c *core.RequestEvent) error {
			code := c.Request.PathValue("code") // 使用 Go 1.22+ 路由参数获取方式

			// 验证邀请码格式
			matched, _ := regexp.MatchString("^[A-Z]{3}-\\d{6}$", code)
			if !matched {
				return c.JSON(http.StatusBadRequest, map[string]any{
					"code":    400,
					"message": "邀请码格式不正确",
				})
			}

			// 使用新 API 查询特定的邀请码
			foundRecord, err := app.FindFirstRecordByFilter("invitation_codes", "code = {:code}", map[string]any{"code": code})
			if err != nil {
				return c.JSON(http.StatusNotFound, map[string]any{
					"code":    404,
					"message": "邀请码不存在",
				})
			}

			// 检查邀请码是否过期 (PocketBase 内部通常存为 UTC)
			expiresAt := foundRecord.GetDateTime("expires_at").Time()
			if time.Now().After(expiresAt) {
				// 清理过期邀请码
				app.Delete(foundRecord)
				return c.JSON(http.StatusGone, map[string]any{
					"code":    410,
					"message": "邀请码已过期",
				})
			}

			// 检查邀请码是否已用完
			maxUses := foundRecord.GetInt("max_uses")
			usedCount := foundRecord.GetInt("used_count")
			if usedCount >= maxUses {
				// 清理已用完邀请码
				app.Delete(foundRecord)
				return c.JSON(http.StatusConflict, map[string]any{
					"code":    409,
					"message": "邀请码已被使用完",
				})
			}

			// 获取关联信息
			ledgerId := foundRecord.GetString("ledger")
			ledger, err := app.FindRecordById("ledgers", ledgerId)
			if err != nil {
				return c.JSON(http.StatusNotFound, map[string]any{
					"code":    404,
					"message": "账本不存在",
				})
			}

			createdById := foundRecord.GetString("created_by")
			creator, err := app.FindRecordById("users", createdById)
			if err != nil {
				return c.JSON(http.StatusInternalServerError, map[string]any{
					"code":    500,
					"message": "内部服务器错误",
				})
			}

			return c.JSON(http.StatusOK, map[string]any{
				"ledgerId":   ledgerId,
				"ledgerName": ledger.GetString("name"),
				"createdBy":  creator.GetString("name"),
				"expiresAt":  expiresAt.Format("2006-01-02 15:04:05"),
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

			authRecord := c.Auth
			if authRecord == nil {
				return c.JSON(http.StatusUnauthorized, map[string]any{
					"code":    401,
					"message": "未认证",
				})
			}

			// 查询邀请码
			foundRecord, err := app.FindFirstRecordByFilter("invitation_codes", "code = {:code}", map[string]any{"code": req.Code})
			if err != nil {
				return c.JSON(http.StatusNotFound, map[string]any{
					"code":    404,
					"message": "邀请码不存在",
				})
			}

			// 检查状态
			expiresAt := foundRecord.GetDateTime("expires_at").Time()
			if time.Now().After(expiresAt) {
				// 清理过期邀请码
				app.Delete(foundRecord)
				return c.JSON(http.StatusGone, map[string]any{
					"code":    410,
					"message": "邀请码已过期",
				})
			}

			if foundRecord.GetInt("used_count") >= foundRecord.GetInt("max_uses") {
				// 清理已用完邀请码
				app.Delete(foundRecord)
				return c.JSON(http.StatusConflict, map[string]any{
					"code":    409,
					"message": "邀请码已被使用完",
				})
			}

			ledgerId := foundRecord.GetString("ledger")

			existingMember, _ := app.FindFirstRecordByFilter(
				"ledger_members",
				"ledger = {:ledger} && user = {:user}",
				map[string]any{"ledger": ledgerId, "user": authRecord.Id},
			)

			if existingMember != nil {
				return c.JSON(http.StatusConflict, map[string]any{
					"code":    409,
					"message": "你已经是该账本的成员",
				})
			}

			var ledgerName string
			err = app.RunInTransaction(func(txApp core.App) error {
				memberCollection, err := txApp.FindCollectionByNameOrId("ledger_members")
				if err != nil {
					return err
				}
				newMember := core.NewRecord(memberCollection)
				newMember.Set("ledger", ledgerId)
				newMember.Set("user", authRecord.Id)
				newMember.Set("role", "member")
				if err := txApp.Save(newMember); err != nil {
					return err
				}

				newUsedCount := foundRecord.GetInt("used_count") + 1
				foundRecord.Set("used_count", newUsedCount)
				if newUsedCount >= foundRecord.GetInt("max_uses") {
					return txApp.Delete(foundRecord)
				}
				return txApp.Save(foundRecord)
			})
			if err != nil {
				return c.JSON(http.StatusInternalServerError, map[string]any{
					"code":    500,
					"message": "加入账本失败",
				})
			}

			ledger, _ := app.FindRecordById("ledgers", ledgerId)
			if ledger != nil {
				ledgerName = ledger.GetString("name")
			}

			return c.JSON(http.StatusOK, map[string]any{
				"success":    true,
				"ledgerId":   ledgerId,
				"ledgerName": ledgerName,
				"message":    "成功加入账本",
			})
		})

		// DELETE /api/invitations/{id} - 删除邀请码
		e.Router.DELETE("/api/invitations/{id}", func(c *core.RequestEvent) error {
			invitationId := c.Request.PathValue("id")

			authRecord := c.Auth
			if authRecord == nil {
				return c.JSON(http.StatusUnauthorized, map[string]any{
					"code":    401,
					"message": "未认证",
				})
			}

			// 查找邀请码
			foundRecord, err := app.FindRecordById("invitation_codes", invitationId)
			if err != nil {
				return c.JSON(http.StatusNotFound, map[string]any{
					"code":    404,
					"message": "邀请码不存在",
				})
			}

			ledgerId := foundRecord.GetString("ledger")

			// 验证账本是否存在且用户是所有者
			ledger, err := app.FindRecordById("ledgers", ledgerId)
			if err != nil {
				return c.JSON(http.StatusNotFound, map[string]any{
					"code":    404,
					"message": "账本不存在",
				})
			}

			if ledger.GetString("owner") != authRecord.Id {
				return c.JSON(http.StatusForbidden, map[string]any{
					"code":    403,
					"message": "只有账本所有者可以删除邀请码",
				})
			}

			// 删除邀请码
			if err := app.Delete(foundRecord); err != nil {
				return c.JSON(http.StatusInternalServerError, map[string]any{
					"code":    500,
					"message": "删除邀请码失败",
				})
			}

			return c.JSON(http.StatusOK, map[string]any{
				"success": true,
				"message": "邀请码已删除",
			})
		})

		// GET /api/ledgers/{ledger_id}/stats - 获取账本统计
		e.Router.GET("/api/ledgers/{ledger_id}/stats", func(c *core.RequestEvent) error {
			ledgerId := c.Request.PathValue("ledger_id")

			authRecord := c.Auth
			if authRecord == nil {
				return c.JSON(http.StatusUnauthorized, map[string]any{
					"code":    401,
					"message": "未认证",
				})
			}

			// 验证账本是否存在
			ledger, err := app.FindRecordById("ledgers", ledgerId)
			if err != nil {
				return c.JSON(http.StatusNotFound, map[string]any{
					"code":    404,
					"message": "账本不存在",
				})
			}

			// 验证权限：账本所有者或成员
			isOwner := ledger.GetString("owner") == authRecord.Id
			var memberRecord *core.Record
			if !isOwner {
				memberRecord, err = app.FindFirstRecordByFilter(
					"ledger_members",
					"ledger = {:ledger} && user = {:user}",
					map[string]any{"ledger": ledgerId, "user": authRecord.Id},
				)
				if err != nil || memberRecord == nil {
					return c.JSON(http.StatusForbidden, map[string]any{
						"code":    403,
						"message": "无权访问该账本",
					})
				}
			}

			// 获取查询参数
			month := c.Request.URL.Query().Get("month")
			var dateFilter string
			if month != "" {
				// 验证 month 格式: YYYY-MM
				matched, _ := regexp.MatchString("^\\d{4}-\\d{2}$", month)
				if !matched {
					return c.JSON(http.StatusBadRequest, map[string]any{
						"code":    400,
						"message": "月份格式不正确，应为 YYYY-MM",
					})
				}

				// 解析月份
				_, err := time.Parse("2006-01", month)
				if err != nil {
					return c.JSON(http.StatusBadRequest, map[string]any{
						"code":    400,
						"message": "无效的月份",
					})
				}

				// 计算下个月
				parsedTime, _ := time.Parse("2006-01", month)
				nextMonth := parsedTime.AddDate(0, 1, 0).Format("2006-01")
				dateFilter = fmt.Sprintf(" && date >= \"%s-01\" && date < \"%s-01\"", month, nextMonth)
			}

			// 获取账本的所有成员
			members, err := app.FindRecordsByFilter(
				"ledger_members",
				"ledger = {:ledger}",
				"", // sort
				0,  // limit
				0,  // offset
				map[string]any{"ledger": ledgerId},
			)
			if err != nil {
				return c.JSON(http.StatusInternalServerError, map[string]any{
					"code":    500,
					"message": "获取成员失败",
				})
			}

			// 构建成员映射
			memberMap := make(map[string]*struct {
				userId       string
				name         string
				email        string
				avatar       string
				totalExpense int
				totalBenefit int
				totalIncome  int
				incomeShare  int
			})

			for _, member := range members {
				userId := member.GetString("user")
				user, err := app.FindRecordById("users", userId)
				if err != nil {
					continue
				}
				memberMap[userId] = &struct {
					userId       string
					name         string
					email        string
					avatar       string
					totalExpense int
					totalBenefit int
					totalIncome  int
					incomeShare  int
				}{
					userId:       userId,
					name:         user.GetString("name"),
					email:        user.GetString("email"),
					avatar:       user.GetString("avatar"),
					totalExpense: 0,
					totalBenefit: 0,
					totalIncome:  0,
					incomeShare:  0,
				}
			}

			// 获取交易记录
			transactions, err := app.FindRecordsByFilter(
				"transactions",
				fmt.Sprintf("ledger = \"%s\"%s", ledgerId, dateFilter),
				"",
				0,
				0,
				nil,
			)
			if err != nil {
				return c.JSON(http.StatusInternalServerError, map[string]any{
					"code":    500,
					"message": "获取交易失败",
				})
			}

			// 计算支出、收入和受益
			totalExpense := 0
			totalBenefit := 0
			totalIncome := 0
			totalIncomeShare := 0
			monthlyExpense := 0
			monthlyIncome := 0
			last7DaysExpense := 0
			last7DaysIncome := 0

			now := time.Now()
			monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
			nextMonthStart := monthStart.AddDate(0, 1, 0)
			last7DaysStart := now.AddDate(0, 0, -6)

			for _, tx := range transactions {
				amount := tx.GetInt("amount")
				payer := tx.GetString("payer")
				txType := tx.GetString("type")
				beneficiary := tx.GetString("beneficiary")
				direction := tx.GetString("direction")
				txDate := tx.GetDateTime("date").Time()

				if direction == "INCOME" {
					// 收入统计
					if member, ok := memberMap[payer]; ok {
						member.totalIncome += amount
						totalIncome += amount

						if !txDate.Before(monthStart) && txDate.Before(nextMonthStart) {
							monthlyIncome += amount
						}

						if !txDate.Before(last7DaysStart) && !txDate.After(now) {
							last7DaysIncome += amount
						}
					}

					// 收入分配
					if txType == "AA" {
						if len(members) > 0 {
							base := amount / len(members)
							remainder := amount % len(members)

							sort.Slice(members, func(i, j int) bool {
								return members[i].GetString("user") < members[j].GetString("user")
							})

							for i, member := range members {
								userId := member.GetString("user")
								share := base
								if i < remainder {
									share += 1
								}
								if m, ok := memberMap[userId]; ok {
									m.incomeShare += share
									totalIncomeShare += share
								}
							}
						}
					} else if txType == "SINGLE" && beneficiary != "" {
						if m, ok := memberMap[beneficiary]; ok {
							m.incomeShare += amount
							totalIncomeShare += amount
						}
					}
				} else {
					// 支出统计
					if member, ok := memberMap[payer]; ok {
						member.totalExpense += amount
						totalExpense += amount

						if !txDate.Before(monthStart) && txDate.Before(nextMonthStart) {
							monthlyExpense += amount
						}

						if !txDate.Before(last7DaysStart) && !txDate.After(now) {
							last7DaysExpense += amount
						}
					}

					// 受益统计
					if txType == "AA" {
						if len(members) > 0 {
							base := amount / len(members)
							remainder := amount % len(members)

							sort.Slice(members, func(i, j int) bool {
								return members[i].GetString("user") < members[j].GetString("user")
							})

							for i, member := range members {
								userId := member.GetString("user")
								benefit := base
								if i < remainder {
									benefit += 1
								}
								if m, ok := memberMap[userId]; ok {
									m.totalBenefit += benefit
									totalBenefit += benefit
								}
							}
						}
					} else if txType == "SINGLE" && beneficiary != "" {
						if m, ok := memberMap[beneficiary]; ok {
							m.totalBenefit += amount
							totalBenefit += amount
						}
					}
				}
			}

			// 计算结余和百分比
			type MemberStat struct {
				UserId       string `json:"userId"`
				Name         string `json:"name"`
				Email        string `json:"email"`
				Avatar       string `json:"avatar"`
				TotalExpense int    `json:"totalExpense"`
				TotalBenefit int    `json:"totalBenefit"`
				TotalIncome  int    `json:"totalIncome"`
				IncomeShare  int    `json:"incomeShare"`
				Balance      int    `json:"balance"`
				Percentage   int    `json:"percentage"`
			}

			var memberStats []MemberStat
			maxBalance := 0

			for _, member := range memberMap {
				balance := member.totalExpense - member.totalIncome - member.totalBenefit + member.incomeShare
				balanceAbs := balance
				if balanceAbs < 0 {
					balanceAbs = -balanceAbs
				}
				if balanceAbs > maxBalance {
					maxBalance = balanceAbs
				}

				stat := MemberStat{
					UserId:       member.userId,
					Name:         member.name,
					Email:        member.email,
					Avatar:       member.avatar,
					TotalExpense: member.totalExpense,
					TotalBenefit: member.totalBenefit,
					TotalIncome:  member.totalIncome,
					IncomeShare:  member.incomeShare,
					Balance:      balance,
				}
				memberStats = append(memberStats, stat)
			}

			// 计算百分比
			for i := range memberStats {
				balanceAbs := memberStats[i].Balance
				if balanceAbs < 0 {
					balanceAbs = -balanceAbs
				}
				if maxBalance > 0 {
					memberStats[i].Percentage = int(float64(balanceAbs) / float64(maxBalance) * 100)
				} else {
					memberStats[i].Percentage = 0
				}
			}

			// 按结余绝对值降序排列
			sort.Slice(memberStats, func(i, j int) bool {
				balanceI := memberStats[i].Balance
				balanceJ := memberStats[j].Balance
				if balanceI < 0 {
					balanceI = -balanceI
				}
				if balanceJ < 0 {
					balanceJ = -balanceJ
				}
				return balanceI > balanceJ
			})

			return c.JSON(http.StatusOK, map[string]any{
				"totalExpense":     totalExpense,
				"totalBenefit":     totalBenefit,
				"totalIncome":      totalIncome,
				"totalIncomeShare": totalIncomeShare,
				"monthlyExpense":   monthlyExpense,
				"monthlyIncome":    monthlyIncome,
				"last7DaysExpense": last7DaysExpense,
				"last7DaysIncome":  last7DaysIncome,
				"memberStats":      memberStats,
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
