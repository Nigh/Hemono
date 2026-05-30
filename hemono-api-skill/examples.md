# Examples

Complete `curl` examples for all Hemono API v1 operations.

Replace `BASE_URL` and `TOKEN` with your actual values:
```bash
BASE_URL="http://localhost:8090"
TOKEN="hmn_your_token_here"
```

---

## Ledgers

### List all ledgers

```bash
curl -s -H "Authorization: Bearer $TOKEN" \
  "$BASE_URL/api/v1/ledgers" | jq
```

### Create a ledger

```bash
curl -s -X POST "$BASE_URL/api/v1/ledgers" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name": "日本旅行"}' | jq
```

### Get ledger detail

```bash
LEDGER_ID="your_ledger_id"
curl -s -H "Authorization: Bearer $TOKEN" \
  "$BASE_URL/api/v1/ledgers/$LEDGER_ID" | jq
```

### Delete a ledger (owner only)

```bash
LEDGER_ID="your_ledger_id"
curl -s -X DELETE "$BASE_URL/api/v1/ledgers/$LEDGER_ID" \
  -H "Authorization: Bearer $TOKEN" | jq
```

---

## Members

### List ledger members

```bash
LEDGER_ID="your_ledger_id"
curl -s -H "Authorization: Bearer $TOKEN" \
  "$BASE_URL/api/v1/ledgers/$LEDGER_ID/members" | jq
```

---

## Transactions

### List transactions

```bash
LEDGER_ID="your_ledger_id"
curl -s -H "Authorization: Bearer $TOKEN" \
  "$BASE_URL/api/v1/ledgers/$LEDGER_ID/transactions" | jq
```

### Create an AA expense (split equally)

¥256.00 dinner, split among all members:

```bash
LEDGER_ID="your_ledger_id"
curl -s -X POST "$BASE_URL/api/v1/ledgers/$LEDGER_ID/transactions" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "amount": 25600,
    "type": "AA",
    "direction": "EXPENSE",
    "note": "晚餐AA",
    "date": "2026-05-31"
  }' | jq
```

### Create a SINGLE expense (one beneficiary)

¥50.00 taxi for a specific person:

```bash
LEDGER_ID="your_ledger_id"
BENEFICIARY_ID="target_user_id"
curl -s -X POST "$BASE_URL/api/v1/ledgers/$LEDGER_ID/transactions" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{
    \"amount\": 5000,
    \"type\": \"SINGLE\",
    \"direction\": \"EXPENSE\",
    \"beneficiary\": \"$BENEFICIARY_ID\",
    \"note\": \"出租车\",
    \"date\": \"2026-05-31\"
  }" | jq
```

### Create an AA income (shared income)

¥1000.00 refund, split equally:

```bash
LEDGER_ID="your_ledger_id"
curl -s -X POST "$BASE_URL/api/v1/ledgers/$LEDGER_ID/transactions" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "amount": 100000,
    "type": "AA",
    "direction": "INCOME",
    "note": "退款",
    "date": "2026-05-31"
  }' | jq
```

### Delete a transaction (payer only)

```bash
TX_ID="your_transaction_id"
curl -s -X DELETE "$BASE_URL/api/v1/transactions/$TX_ID" \
  -H "Authorization: Bearer $TOKEN" | jq
```

---

## Statistics

### Get all-time stats

```bash
LEDGER_ID="your_ledger_id"
curl -s -H "Authorization: Bearer $TOKEN" \
  "$BASE_URL/api/v1/ledgers/$LEDGER_ID/stats" | jq
```

### Get monthly stats

```bash
LEDGER_ID="your_ledger_id"
curl -s -H "Authorization: Bearer $TOKEN" \
  "$BASE_URL/api/v1/ledgers/$LEDGER_ID/stats?month=2026-05" | jq
```

---

## Invitations

### Get active invitation (owner only)

```bash
LEDGER_ID="your_ledger_id"
curl -s -H "Authorization: Bearer $TOKEN" \
  "$BASE_URL/api/v1/ledgers/$LEDGER_ID/invitation" | jq
```

### Generate invitation code (owner only)

```bash
LEDGER_ID="your_ledger_id"
curl -s -X POST "$BASE_URL/api/v1/ledgers/$LEDGER_ID/invitation" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"max_uses": 5}' | jq
```

### Join a ledger via invitation code

```bash
curl -s -X POST "$BASE_URL/api/v1/invitations/join" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"code": "ABC-123456"}' | jq
```

### Delete an invitation (owner only)

```bash
INVITATION_ID="your_invitation_id"
curl -s -X DELETE "$BASE_URL/api/v1/invitations/$INVITATION_ID" \
  -H "Authorization: Bearer $TOKEN" | jq
```

---

## Python Example

```python
import requests

BASE_URL = "http://localhost:8090"
TOKEN = "hmn_your_token_here"

headers = {
    "Authorization": f"Bearer {TOKEN}",
    "Content-Type": "application/json",
}

# List ledgers
ledgers = requests.get(f"{BASE_URL}/api/v1/ledgers", headers=headers).json()
print(ledgers)

# Create a transaction
ledger_id = ledgers[0]["id"]
tx = requests.post(
    f"{BASE_URL}/api/v1/ledgers/{ledger_id}/transactions",
    headers=headers,
    json={
        "amount": 12850,  # ¥128.50
        "type": "AA",
        "direction": "EXPENSE",
        "note": "午餐",
        "date": "2026-05-31",
    },
).print(tx.json())

# Get stats
stats = requests.get(
    f"{BASE_URL}/api/v1/ledgers/{ledger_id}/stats",
    headers=headers,
    params={"month": "2026-05"},
).json()
print(f"本月支出: ¥{stats['monthlyExpense'] / 100:.2f}")
```

## JavaScript/Node.js Example

```javascript
const BASE_URL = 'http://localhost:8090';
const TOKEN = 'hmn_your_token_here';

const headers = {
  Authorization: `Bearer ${TOKEN}`,
  'Content-Type': 'application/json',
};

// List ledgers
const ledgers = await fetch(`${BASE_URL}/api/v1/ledgers`, { headers }).then((r) => r.json());
console.log(ledgers);

// Create an expense
const ledgerId = ledgers[0].id;
const tx = await fetch(`${BASE_URL}/api/v1/ledgers/${ledgerId}/transactions`, {
  method: 'POST',
  headers,
  body: JSON.stringify({
    amount: 8800, // ¥88.00
    type: 'AA',
    direction: 'EXPENSE',
    note: '打车',
    date: '2026-05-31',
  }),
}).then((r) => r.json());
console.log(tx);
```
