# 邀请码功能部署指南

## 功能概述

已实现的"加入账本"功能包括：
- 账本拥有者生成邀请码
- 用户通过邀请码加入账本
- 邀请码24小时有效期，一次性使用
- 完整的错误处理和用户体验

## 待完成步骤

### 1. 创建数据库集合

需要在 PocketBase 管理界面中手动创建 `invitation_codes` 集合：

**集合名称：** `invitation_codes`

**字段列表：**
1. **code** (Text) - 邀请码
   - 最小/最大长度：10
   - 正则：`^[A-Z]{3}-\d{6}$`
   - 必填，唯一

2. **ledger** (Relation) - 关联账本
   - 关联集合：ledgers
   - 最大选择数：1
   - 级联删除：是
   - 必填

3. **created_by** (Relation) - 创建者
   - 关联集合：users
   - 最大选择数：1
   - 级联删除：否
   - 必填

4. **expires_at** (Date) - 过期时间
   - 必填

5. **used** (Bool) - 是否已使用
   - 默认值：false

6. **used_at** (Date) - 使用时间
   - 可选

7. **used_by** (Relation) - 使用者
   - 关联集合：users
   - 最大选择数：1
   - 级联删除：否
   - 可选

8. **created** (Date) - 创建时间（系统字段）

9. **updated** (Date) - 更新时间（系统字段）

**权限设置：**
- **Create Rule:** `created_by = @request.auth.id`
- **View Rule:** 空（允许所有人查看，用于验证邀请码）
- **Update Rule:** `created_by = @request.auth.id`
- **Delete Rule:** `created_by = @request.auth.id`

### 2. 部署后端

```bash
cd /home/engine/project/backend
go build
# 运行服务器
./hemono
```

### 3. 启动前端

```bash
cd /home/engine/project/frontend
npm install
npm run dev
```

## 功能测试

### 测试生成邀请码
1. 登录用户A，创建账本
2. 点击账本的"邀请成员"按钮
3. 生成邀请码

### 测试加入账本
1. 登录用户B
2. 点击头像菜单的"📥 加入账本"
3. 输入邀请码
4. 确认加入

## 文件结构

```
frontend/src/
├── lib/
│   └── components/
│       ├── JoinLedgerModal.svelte     # 加入账本模态框
│       └── GenerateInviteModal.svelte # 生成邀请码模态框
└── routes/
    ├── +layout.svelte                 # 布局文件（已更新）
    └── +page.svelte                   # 主页（已更新）

backend/
└── main.go                           # 主程序（已更新API）
```

## API 接口

### 生成邀请码
- **POST** `/api/invitations/generate`
- **Body:** `{"ledger_id": "账本ID"}`
- **权限：** 账本拥有者

### 验证邀请码
- **GET** `/api/invitations/by-code/{code}`
- **返回：** 账本信息

### 加入账本
- **POST** `/api/invitations/join`
- **Body:** `{"code": "邀请码"}`
- **权限：** 已登录用户

## 注意事项

1. 确保 PocketBase 服务器正在运行
2. 确保数据库连接正常
3. 邀请码在生成24小时后自动过期
4. 每个邀请码只能使用一次
5. 用户不能重复加入同一账本