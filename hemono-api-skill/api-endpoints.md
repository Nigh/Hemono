# API Endpoints

All endpoints require `Authorization: Bearer <token>` header unless noted otherwise.

---

## Ledgers

### GET /api/v1/ledgers

List all ledgers the authenticated user owns or is a member of.

**Auth**: Required

**Response**: `200 OK`
```json
[
  {
    "id": "abc123",
    "name": "旅行账本",
    "owner": "user_id",
    "created": "2026-01-15T10:30:00Z"
  }
]
```

---

### POST /api/v1/ledgers

Create a new ledger. The authenticated user becomes the owner and is automatically added as an `admin` member.

**Auth**: Required

**Request Body**:
```json
{
  "name": "账本名称"
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | Yes | Ledger name, cannot be empty |

**Response**: `200 OK`
```json
{
  "id": "new_ledger_id",
  "name": "账本名称",
  "owner": "user_id",
  "created": "2026-05-31T02:00:00Z"
}
```

**Errors**:
- `400` — 账本名称不能为空

---

### GET /api/v1/ledgers/{id}

Get details of a specific ledger. User must be the owner or a member.

**Auth**: Required

**Path Parameters**:
| Param | Type | Description |
|-------|------|-------------|
| `id` | string | Ledger ID |

**Response**: `200 OK`
```json
{
  "id": "abc123",
  "name": "旅行账本",
  "owner": "user_id",
  "created": "2026-01-15T10:30:00Z"
}
```

**Errors**:
- `404` — 账本不存在
- `403` — 无权访问该账本

---

### DELETE /api/v1/ledgers/{id}

Delete a ledger. **Owner only**. This does not cascade-delete transactions or members.

**Auth**: Required

**Path Parameters**:
| Param | Type | Description |
|-------|------|-------------|
| `id` | string | Ledger ID |

**Response**: `200 OK`
```json
{
  "success": true,
  "message": "账本已删除"
}
```

**Errors**:
- `404` — 账本不存在
- `403` — 只有账本所有者可以删除账本

---

## Members

### GET /api/v1/ledgers/{id}/members

List all members of a ledger. User must be the owner or a member.

**Auth**: Required

**Path Parameters**:
| Param | Type | Description |
|-------|------|-------------|
| `id` | string | Ledger ID |

**Response**: `200 OK`
```json
[
  {
    "id": "user_id",
    "name": "张三",
    "email": "zhangsan@example.com",
    "avatar": "https://avatars.githubusercontent.com/u/123456"
  }
]
```

**Errors**:
- `404` — 账本不存在
- `403` — 无权访问该账本

---

## Transactions

### GET /api/v1/ledgers/{id}/transactions

List all transactions in a ledger, sorted by date descending then created descending. User must be the owner or a member.

**Auth**: Required

**Path Parameters**:
| Param | Type | Description |
|-------|------|-------------|
| `id` | string | Ledger ID |

**Response**: `200 OK`
```json
[
  {
    "id": "tx_id",
    "ledger": "ledger_id",
    "payer": "user_id",
    "payerName": "张三",
    "payerAvatar": "https://...",
    "amount": 12850,
    "type": "AA",
    "direction": "EXPENSE",
    "beneficiary": "user_id",
    "note": "晚餐",
    "date": "2026-05-31",
    "created": "2026-05-31T12:00:00Z"
  }
]
```

**Notes**:
- `payerName` and `payerAvatar` are resolved from the payer user record
- `beneficiary` is only present when `type` is `SINGLE`
- `amount` is in cents (分)

**Errors**:
- `404` — 账本不存在
- `403` — 无权访问该账本

---

### POST /api/v1/ledgers/{id}/transactions

Create a new transaction in a ledger. The authenticated user becomes the `payer`. User must be the owner or a member.

**Auth**: Required

**Path Parameters**:
| Param | Type | Description |
|-------|------|-------------|
| `id` | string | Ledger ID |

**Request Body**:
```json
{
  "amount": 12850,
  "type": "AA",
  "direction": "EXPENSE",
  "beneficiary": "",
  "note": "晚餐",
  "date": "2026-05-31"
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `amount` | int | Yes | Amount in cents (分), must be > 0 |
| `type` | string | Yes | `AA` (split equally) or `SINGLE` (one beneficiary) |
| `direction` | string | Yes | `EXPENSE` or `INCOME` |
| `beneficiary` | string | Conditional | Required when `type` = `SINGLE`. User ID of the beneficiary |
| `note` | string | No | Description/note for the transaction |
| `date` | string | No | Date in `YYYY-MM-DD` format, defaults to today |

**Response**: `200 OK`
```json
{
  "id": "new_tx_id",
  "ledger": "ledger_id",
  "payer": "user_id",
  "amount": 12850,
  "type": "AA",
  "direction": "EXPENSE",
  "note": "晚餐",
  "date": "2026-05-31"
}
```

**Validation Rules**:
- `amount` must be > 0
- `type` must be `AA` or `SINGLE`
- `direction` must be `EXPENSE` or `INCOME`
- If `type` = `SINGLE`, `beneficiary` must not be empty
- `date` must match `YYYY-MM-DD` format if provided

**Errors**:
- `400` — 金额必须大于 0 / 类型必须为 AA 或 SINGLE / 方向必须为 EXPENSE 或 INCOME / SINGLE 类型必须指定受益人 / 日期格式不正确
- `404` — 账本不存在
- `403` — 无权访问该账本

---

### DELETE /api/v1/transactions/{id}

Delete a transaction. **Payer only** (the user who created the transaction).

**Auth**: Required

**Path Parameters**:
| Param | Type | Description |
|-------|------|-------------|
| `id` | string | Transaction ID |

**Response**: `200 OK`
```json
{
  "success": true,
  "message": "交易已删除"
}
```

**Errors**:
- `404` — 交易不存在
- `403` — 只有付款人可以删除交易

---

## Statistics

### GET /api/v1/ledgers/{id}/stats

Get statistical summary for a ledger. User must be the owner or a member.

**Auth**: Required

**Path Parameters**:
| Param | Type | Description |
|-------|------|-------------|
| `id` | string | Ledger ID |

**Query Parameters**:
| Param | Type | Required | Description |
|-------|------|----------|-------------|
| `month` | string | No | Filter by month, format `YYYY-MM` |

**Response**: `200 OK`
```json
{
  "totalExpense": 50000,
  "totalBenefit": 25000,
  "totalIncome": 10000,
  "totalIncomeShare": 5000,
  "monthlyExpense": 20000,
  "monthlyIncome": 5000,
  "last7DaysExpense": 8000,
  "last7DaysIncome": 2000,
  "memberStats": [
    {
      "userId": "user_id",
      "name": "张三",
      "email": "zhangsan@example.com",
      "avatar": "https://...",
      "totalExpense": 30000,
      "totalBenefit": 15000,
      "totalIncome": 5000,
      "incomeShare": 2500,
      "balance": 7500,
      "percentage": 100
    }
  ]
}
```

**All amounts are in cents (分).**

**Balance Calculation**:
```
balance = totalExpense - totalIncome - totalBenefit + incomeShare
```
- Positive balance: this member is owed money
- Negative balance: this member owes money

**Percentage**: Relative to the member with the highest absolute balance (100%).

**Sorting**: Members sorted by absolute balance descending.

**Errors**:
- `400` — 月份格式不正确
- `404` — 账本不存在
- `403` — 无权访问该账本

---

## Invitations

### GET /api/v1/ledgers/{id}/invitation

Get the current active invitation code for a ledger. **Owner only**.

**Auth**: Required

**Path Parameters**:
| Param | Type | Description |
|-------|------|-------------|
| `id` | string | Ledger ID |

**Response** (active invitation exists): `200 OK`
```json
{
  "exists": true,
  "id": "invitation_id",
  "code": "ABC-123456",
  "expiresAt": "2026-06-01 14:30",
  "maxUses": 10,
  "usedCount": 3
}
```

**Response** (no active invitation): `200 OK`
```json
{
  "exists": false
}
```

**Errors**:
- `404` — 账本不存在
- `403` — 只有账本所有者可以查看邀请码

---

### POST /api/v1/ledgers/{id}/invitation

Generate a new invitation code for a ledger. **Owner only**. Returns `409` if an active invitation already exists.

**Auth**: Required

**Path Parameters**:
| Param | Type | Description |
|-------|------|-------------|
| `id` | string | Ledger ID |

**Request Body**:
```json
{
  "max_uses": 10
}
```

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `max_uses` | int | No | 10 | Max number of uses, range 1-99 |

**Response**: `200 OK`
```json
{
  "code": "ABC-123456",
  "expires_at": "2026-06-01 14:30",
  "max_uses": 10,
  "used_count": 0
}
```

**Invitation code format**: `XXX-YYYYYY` (3 uppercase letters, dash, 6 digits)

**Expiry**: 24 hours from creation.

**Errors**:
- `404` — 账本不存在
- `403` — 只有账本所有者可以生成邀请码
- `409` — 该账本已存在有效邀请码

---

### POST /api/v1/invitations/join

Join a ledger using an invitation code.

**Auth**: Required

**Request Body**:
```json
{
  "code": "ABC-123456"
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `code` | string | Yes | Invitation code |

**Response**: `200 OK`
```json
{
  "success": true,
  "ledgerId": "ledger_id",
  "ledgerName": "旅行账本",
  "message": "成功加入账本"
}
```

**Errors**:
- `400` — 邀请码不能为空
- `404` — 邀请码不存在
- `409` — 你已经是该账本的成员 / 邀请码已被使用完
- `410` — 邀请码已过期

---

### DELETE /api/v1/invitations/{id}

Delete an invitation code. **Owner only**.

**Auth**: Required

**Path Parameters**:
| Param | Type | Description |
|-------|------|-------------|
| `id` | string | Invitation code record ID |

**Response**: `200 OK`
```json
{
  "success": true,
  "message": "邀请码已删除"
}
```

**Errors**:
- `404` — 邀请码不存在 / 账本不存在
- `403` — 只有账本所有者可以删除邀请码
