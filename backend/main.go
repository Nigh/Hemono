package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
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

			// 清理已过期或已用完的邀请码
			expiresAt := foundRecord.GetDateTime("expires_at").Time()
			if time.Now().After(expiresAt) || foundRecord.GetInt("used_count") >= foundRecord.GetInt("max_uses") {
				app.Delete(foundRecord)
				return c.JSON(http.StatusOK, map[string]any{
					"exists": false,
				})
			}

			return c.JSON(http.StatusOK, map[string]any{
				"exists":    true,
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

			// 写入成员并更新计数（推荐在事务中执行，这里简化处理）
			memberCollection, _ := app.FindCollectionByNameOrId("ledger_members")
			newMember := core.NewRecord(memberCollection)
			newMember.Set("ledger", ledgerId)
			newMember.Set("user", authRecord.Id)
			newMember.Set("role", "member")

			if err := app.Save(newMember); err != nil {
				return c.JSON(http.StatusInternalServerError, map[string]any{
					"code":    500,
					"message": "加入账本失败",
				})
			}

			newUsedCount := foundRecord.GetInt("used_count") + 1
			foundRecord.Set("used_count", newUsedCount)

			// 检查是否已用完，如果是则删除邀请码
			if newUsedCount >= foundRecord.GetInt("max_uses") {
				app.Delete(foundRecord)
			} else {
				app.Save(foundRecord)
			}

			ledger, _ := app.FindRecordById("ledgers", ledgerId)

			return c.JSON(http.StatusOK, map[string]any{
				"success":    true,
				"ledgerId":   ledgerId,
				"ledgerName": ledger.GetString("name"),
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
				}{
					userId:       userId,
					name:         user.GetString("name"),
					email:        user.GetString("email"),
					avatar:       user.GetString("avatar"),
					totalExpense: 0,
					totalBenefit: 0,
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

			// 计算支出和受益
			totalExpense := 0
			totalBenefit := 0
			monthlyExpense := 0
			last7DaysExpense := 0

			now := time.Now()
			monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
			nextMonthStart := monthStart.AddDate(0, 1, 0)
			last7DaysStart := now.AddDate(0, 0, -6)

			for _, tx := range transactions {
				amount := tx.GetInt("amount")
				payer := tx.GetString("payer")
				txType := tx.GetString("type")
				beneficiary := tx.GetString("beneficiary")
				txDate := tx.GetDateTime("date").Time()

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
					// AA：平均分配给所有成员
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
					// SINGLE：分配给受益人
					if m, ok := memberMap[beneficiary]; ok {
						m.totalBenefit += amount
						totalBenefit += amount
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
				Balance      int    `json:"balance"`
				Percentage   int    `json:"percentage"`
			}

			var memberStats []MemberStat
			maxBalance := 0

			for _, member := range memberMap {
				balance := member.totalExpense - member.totalBenefit
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
				"monthlyExpense":   monthlyExpense,
				"last7DaysExpense": last7DaysExpense,
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
