<script lang="ts">
	import { pb, currentUser } from '$lib/pb';

	let activeSection = $state('auth');
	let copiedEndpoint = $state('');

	function getBaseUrl(): string {
		if (typeof window !== 'undefined') return window.location.origin;
		return 'http://localhost:8090';
	}

	function getAuthHeader(): string {
		return pb.authStore.token || '<YOUR_TOKEN>';
	}

	function copyToClipboard(text: string) {
		navigator.clipboard.writeText(text).then(() => {
			copiedEndpoint = text;
			setTimeout(() => (copiedEndpoint = ''), 2000);
		});
	}

	function curlExample(method: string, path: string, body?: string): string {
		let cmd = `curl -X ${method} '${getBaseUrl()}${path}' \\\n  -H 'Authorization: Bearer ${getAuthHeader()}'`;
		if (body) {
			cmd += ` \\\n  -H 'Content-Type: application/json' \\\n  -d '${body}'`;
		}
		return cmd;
	}

	interface ApiEndpoint {
		method: string;
		path: string;
		description: string;
		auth: boolean;
		permission: string;
		requestBody?: string;
		responseExample?: string;
		curlExample?: string;
	}

	interface ApiSection {
		id: string;
		title: string;
		icon: string;
		description: string;
		endpoints: ApiEndpoint[];
	}

	const sections: ApiSection[] = [
		{
			id: 'auth',
			title: '认证',
			icon: '🔐',
			description: '用户认证相关接口',
			endpoints: [
				{
					method: 'POST',
					path: '/api/users/auth-via-oauth2',
					description: '通过 GitHub OAuth2 登录',
					auth: false,
					permission: '公开',
					requestBody: '{"provider": "github", "code": "...", "codeVerifier": "..."}',
					responseExample:
						'{"token": "eyJ...", "record": {"id": "...", "email": "...", "name": "...", "avatar": "..."}}'
				},
				{
					method: 'POST',
					path: '/api/users/auth-refresh',
					description: '刷新认证 Token',
					auth: true,
					permission: '已认证用户',
					responseExample: '{"token": "eyJ...", "record": {...}}'
				}
			]
		},
		{
			id: 'users',
			title: '用户',
			icon: '👤',
			description: 'PocketBase 内置用户集合 CRUD',
			endpoints: [
				{
					method: 'GET',
					path: '/api/users',
					description: '获取用户列表（仅自己的信息）',
					auth: true,
					permission: '仅自己'
				},
				{
					method: 'GET',
					path: '/api/users/{id}',
					description: '获取单个用户信息',
					auth: true,
					permission: '仅自己'
				},
				{
					method: 'PATCH',
					path: '/api/users/{id}',
					description: '更新用户信息',
					auth: true,
					permission: '仅自己',
					requestBody: '{"name": "新名字", "avatar": "文件"}'
				},
				{
					method: 'DELETE',
					path: '/api/users/{id}',
					description: '删除用户',
					auth: true,
					permission: '仅自己'
				}
			]
		},
		{
			id: 'ledgers',
			title: '账本',
			icon: '📒',
			description: '账本的增删改查操作',
			endpoints: [
				{
					method: 'GET',
					path: '/api/ledgers',
					description: '获取账本列表（创建的 + 作为成员加入的）',
					auth: true,
					permission: '所有者或成员',
					responseExample:
						'{"items": [{"id": "...", "name": "合租账本", "owner": "...", "created": "..."}], "totalPages": 1}'
				},
				{
					method: 'POST',
					path: '/api/ledgers',
					description: '创建新账本（自动将创建者添加为 admin 成员）',
					auth: true,
					permission: '已认证用户',
					requestBody: '{"name": "新账本名称", "owner": "<当前用户ID>"}',
					responseExample: '{"id": "...", "name": "新账本名称", "owner": "..."}'
				},
				{
					method: 'GET',
					path: '/api/ledgers/{id}',
					description: '获取单个账本详情',
					auth: true,
					permission: '所有者或成员'
				},
				{
					method: 'PATCH',
					path: '/api/ledgers/{id}',
					description: '更新账本信息',
					auth: true,
					permission: '仅所有者',
					requestBody: '{"name": "更新后的名称"}'
				},
				{
					method: 'DELETE',
					path: '/api/ledgers/{id}',
					description: '删除账本',
					auth: true,
					permission: '仅所有者'
				}
			]
		},
		{
			id: 'ledger_members',
			title: '账本成员',
			icon: '👥',
			description: '账本成员管理',
			endpoints: [
				{
					method: 'GET',
					path: '/api/ledger_members',
					description: '获取账本成员列表',
					auth: true,
					permission: '所有者或成员',
					responseExample:
						'{"items": [{"id": "...", "ledger": "...", "user": "...", "role": "admin", "expand": {"user": {"name": "...", "avatar": "..."}}}], "totalPages": 1}'
				},
				{
					method: 'POST',
					path: '/api/ledger_members',
					description: '添加成员（通常通过邀请码接口完成）',
					auth: true,
					permission: '无（通过邀请码流程）',
					requestBody: '{"ledger": "<账本ID>", "user": "<用户ID>", "role": "member"}'
				}
			]
		},
		{
			id: 'transactions',
			title: '交易记录',
			icon: '💰',
			description: '记账交易的增删改查',
			endpoints: [
				{
					method: 'GET',
					path: '/api/transactions',
					description: '获取交易记录列表（支持 filter/sort 分页）',
					auth: true,
					permission: '所有者或成员',
					responseExample:
						'{"items": [{"id": "...", "ledger": "...", "payer": "...", "amount": 15000, "type": "AA", "direction": "EXPENSE", "beneficiary": null, "note": "晚餐", "date": "2024-01-15 00:00:00", "expand": {"payer": {"name": "..."}}}], "totalPages": 1}'
				},
				{
					method: 'POST',
					path: '/api/transactions',
					description: '创建交易记录（金额单位：分）',
					auth: true,
					permission: '所有者或成员',
					requestBody:
						'{"ledger": "<账本ID>", "payer": "<付款人ID>", "amount": 15000, "type": "AA", "direction": "EXPENSE", "beneficiary": null, "note": "晚餐AA", "date": "2024-01-15"}',
					responseExample:
						'{"id": "...", "ledger": "...", "payer": "...", "amount": 15000, "type": "AA", "direction": "EXPENSE", "note": "晚餐AA"}'
				},
				{
					method: 'GET',
					path: '/api/transactions/{id}',
					description: '获取单个交易记录',
					auth: true,
					permission: '所有者或成员'
				},
				{
					method: 'PATCH',
					path: '/api/transactions/{id}',
					description: '更新交易记录',
					auth: true,
					permission: '仅付款人',
					requestBody: '{"amount": 20000, "note": "更新备注"}'
				},
				{
					method: 'DELETE',
					path: '/api/transactions/{id}',
					description: '删除交易记录',
					auth: true,
					permission: '仅付款人'
				}
			]
		},
		{
			id: 'invitations',
			title: '邀请码',
			icon: '🎟️',
			description: '邀请码管理（自定义 API 路由）',
			endpoints: [
				{
					method: 'GET',
					path: '/api/invitations/by-ledger/{ledger_id}',
					description: '获取账本当前有效邀请码',
					auth: true,
					permission: '仅账本所有者',
					responseExample:
						'{"exists": true, "id": "...", "code": "ABC-123456", "expiresAt": "2024-01-16 12:00", "maxUses": 10, "usedCount": 3}'
				},
				{
					method: 'POST',
					path: '/api/invitations/generate',
					description: '生成新邀请码（24小时有效，1-99次使用）',
					auth: true,
					permission: '仅账本所有者',
					requestBody: '{"ledger_id": "<账本ID>", "max_uses": 10}',
					responseExample:
						'{"code": "ABC-123456", "expires_at": "2024-01-16 12:00", "max_uses": 10, "used_count": 0}'
				},
				{
					method: 'GET',
					path: '/api/invitations/by-code/{code}',
					description: '通过邀请码查询账本信息（公开接口）',
					auth: false,
					permission: '公开',
					responseExample:
						'{"ledgerId": "...", "ledgerName": "合租账本", "createdBy": "张三", "expiresAt": "2024-01-16 12:00:00", "maxUses": 10, "usedCount": 3}'
				},
				{
					method: 'POST',
					path: '/api/invitations/join',
					description: '通过邀请码加入账本',
					auth: true,
					permission: '已认证用户',
					requestBody: '{"code": "ABC-123456"}',
					responseExample:
						'{"success": true, "ledgerId": "...", "ledgerName": "合租账本", "message": "成功加入账本"}'
				},
				{
					method: 'DELETE',
					path: '/api/invitations/{id}',
					description: '删除邀请码',
					auth: true,
					permission: '仅账本所有者',
					responseExample: '{"success": true, "message": "邀请码已删除"}'
				}
			]
		},
		{
			id: 'stats',
			title: '统计',
			icon: '📊',
			description: '账本统计数据（自定义 API 路由）',
			endpoints: [
				{
					method: 'GET',
					path: '/api/ledgers/{ledger_id}/stats',
					description: '获取账本统计数据（支持 ?month=YYYY-MM 参数筛选）',
					auth: true,
					permission: '所有者或成员',
					responseExample:
						'{"totalExpense": 150000, "totalBenefit": 50000, "totalIncome": 80000, "totalIncomeShare": 40000, "monthlyExpense": 30000, "monthlyIncome": 15000, "last7DaysExpense": 12000, "last7DaysIncome": 5000, "memberStats": [{"userId": "...", "name": "张三", "email": "...", "avatar": "...", "totalExpense": 80000, "totalBenefit": 25000, "totalIncome": 40000, "incomeShare": 20000, "balance": 45000, "percentage": 100}]}'
				}
			]
		}
	];

	function getMethodColor(method: string): string {
		switch (method) {
			case 'GET':
				return 'badge-info';
			case 'POST':
				return 'badge-success';
			case 'PATCH':
				return 'badge-warning';
			case 'DELETE':
				return 'badge-error';
			default:
				return 'badge-ghost';
		}
	}

	function scrollToSection(id: string) {
		activeSection = id;
		document.getElementById(`section-${id}`)?.scrollIntoView({ behavior: 'smooth' });
	}
</script>

<div class="inset-0 bg-base-200 fixed z-[60] overflow-auto">
	<div class="max-w-4xl p-6 space-y-6 mx-auto">
		<div class="flex items-center justify-between">
			<div>
				<h1 class="text-2xl font-bold">API 文档</h1>
				<p class="text-base-content/60 text-sm mt-1">
					荷物账本接口参考 — 基于 PocketBase REST API + 自定义路由
				</p>
			</div>
			<a href="/" class="btn btn-ghost btn-sm">
				<svg
					xmlns="http://www.w3.org/2000/svg"
					class="h-4 w-4"
					fill="none"
					viewBox="0 0 24 24"
					stroke="currentColor"
				>
					<path
						stroke-linecap="round"
						stroke-linejoin="round"
						stroke-width="2"
						d="M15 19l-7-7 7-7"
					/>
				</svg>
				返回
			</a>
		</div>

		<div class="card bg-base-100 border-base-300 border">
			<div class="card-body p-4">
				<h2 class="card-title text-sm">认证方式</h2>
				<p class="text-base-content/70 text-xs">所有需要认证的接口须在请求头中携带 Token：</p>
				<div class="mockup-code text-xs mt-2">
					<pre><code>Authorization: Bearer &lt;TOKEN&gt;</code></pre>
				</div>
				{#if $currentUser}
					<div class="mt-2">
						<p class="text-xs text-base-content/60 mb-1">当前 Token：</p>
						<div class="gap-2 flex items-center">
							<code class="text-xs bg-base-200 px-2 py-1 rounded block max-w-full truncate"
								>{pb.authStore.token}</code
							>
							<button
								class="btn btn-ghost btn-xs"
								onclick={() => copyToClipboard(pb.authStore.token || '')}
							>
								{copiedEndpoint === pb.authStore.token ? '已复制' : '复制'}
							</button>
						</div>
					</div>
				{/if}
			</div>
		</div>

		<div class="gap-6 flex">
			<nav class="lg:block w-48 top-20 sticky hidden shrink-0 self-start">
				<ul class="menu menu-sm space-y-1">
					{#each sections as section (section.id)}
						<li>
							<button
								class="gap-2 flex items-center {activeSection === section.id ? 'active' : ''}"
								onclick={() => scrollToSection(section.id)}
							>
								<span>{section.icon}</span>
								<span>{section.title}</span>
							</button>
						</li>
					{/each}
				</ul>
			</nav>

			<div class="space-y-8 min-w-0 flex-1">
				{#each sections as section (section.id)}
					<div id="section-{section.id}" class="scroll-mt-20">
						<div class="gap-2 mb-4 flex items-center">
							<span class="text-xl">{section.icon}</span>
							<h2 class="text-lg font-bold">{section.title}</h2>
						</div>
						<p class="text-base-content/60 text-sm mb-4">{section.description}</p>

						<div class="space-y-3">
							{#each section.endpoints as endpoint (endpoint.path + endpoint.method)}
								<div
									class="card bg-base-100 border-base-300 hover:border-primary/30 border transition-colors"
								>
									<div class="card-body p-4">
										<div class="gap-3 flex items-start">
											<span
												class="badge badge-sm {getMethodColor(
													endpoint.method
												)} font-mono font-bold shrink-0"
											>
												{endpoint.method}
											</span>
											<div class="min-w-0 flex-1">
												<div class="gap-2 flex flex-wrap items-center">
													<code class="text-sm font-mono bg-base-200 px-2 py-0.5 rounded break-all"
														>{endpoint.path}</code
													>
													<span class="badge badge-xs badge-outline">{endpoint.permission}</span>
												</div>
												<p class="text-sm text-base-content/70 mt-1">{endpoint.description}</p>

												{#if endpoint.requestBody || endpoint.responseExample || endpoint.curlExample}
													<div class="collapse-arrow mt-3 bg-base-200/50 collapse">
														<input type="checkbox" />
														<div
															class="collapse-title text-xs font-medium text-base-content/60 py-2 min-h-0"
														>
															查看详情
														</div>
														<div class="collapse-content px-4">
															{#if endpoint.requestBody}
																<div class="mb-3">
																	<p class="text-xs font-semibold text-base-content/60 mb-1">
																		请求体 (JSON)：
																	</p>
																	<div class="mockup-code text-xs">
																		<pre><code>{endpoint.requestBody}</code></pre>
																	</div>
																</div>
															{/if}
															{#if endpoint.responseExample}
																<div class="mb-3">
																	<p class="text-xs font-semibold text-base-content/60 mb-1">
																		响应示例：
																	</p>
																	<div class="mockup-code text-xs">
																		<pre><code>{endpoint.responseExample}</code></pre>
																	</div>
																</div>
															{/if}
															<div>
																<p class="text-xs font-semibold text-base-content/60 mb-1">
																	cURL 示例：
																</p>
																<div class="mockup-code text-xs">
																	<pre><code
																			>{endpoint.method === 'GET' || endpoint.method === 'DELETE'
																				? curlExample(endpoint.method, endpoint.path)
																				: curlExample(
																						endpoint.method,
																						endpoint.path,
																						endpoint.requestBody || '{}'
																					)}</code
																		></pre>
																</div>
															</div>
														</div>
													</div>
												{/if}
											</div>
										</div>
									</div>
								</div>
							{/each}
						</div>
					</div>
				{/each}

				<div class="divider"></div>

				<div class="card bg-base-200 border-base-300 border">
					<div class="card-body p-4">
						<h3 class="card-title text-sm">数据模型说明</h3>
						<div class="space-y-3 text-xs text-base-content/70">
							<div>
								<p class="font-semibold text-base-content">金额单位</p>
								<p>
									所有金额字段以 <code class="bg-base-300 px-1 rounded">分（cents）</code>
									为单位整数存储。显示时除以 100 并使用
									<code class="bg-base-300 px-1 rounded">.toFixed(2)</code> 格式化。
								</p>
							</div>
							<div>
								<p class="font-semibold text-base-content">交易类型 (type)</p>
								<p>
									<code class="bg-base-300 px-1 rounded">AA</code> — 全员均摊 |
									<code class="bg-base-300 px-1 rounded">SINGLE</code> — 指定受益人
								</p>
							</div>
							<div>
								<p class="font-semibold text-base-content">交易方向 (direction)</p>
								<p>
									<code class="bg-base-300 px-1 rounded">EXPENSE</code> — 支出 |
									<code class="bg-base-300 px-1 rounded">INCOME</code> — 收入
								</p>
							</div>
							<div>
								<p class="font-semibold text-base-content">成员角色 (role)</p>
								<p>
									<code class="bg-base-300 px-1 rounded">admin</code> — 管理员（账本创建者） |
									<code class="bg-base-300 px-1 rounded">member</code> — 普通成员
								</p>
							</div>
							<div>
								<p class="font-semibold text-base-content">邀请码格式</p>
								<p>
									<code class="bg-base-300 px-1 rounded">ABC-123456</code> — 3 位大写字母 + 连字符 + 6
									位数字，24 小时有效
								</p>
							</div>
							<div>
								<p class="font-semibold text-base-content">PocketBase 筛选语法</p>
								<p>
									列表接口支持 <code class="bg-base-300 px-1 rounded">filter</code>、<code
										class="bg-base-300 px-1 rounded">sort</code
									>、<code class="bg-base-300 px-1 rounded">page</code>、<code
										class="bg-base-300 px-1 rounded">perPage</code
									> 查询参数。
								</p>
								<p>
									例：<code class="bg-base-300 px-1 rounded"
										>?filter=ledger="&#123;ledgerId&#125;"&sort=-date,-created</code
									>
								</p>
							</div>
						</div>
					</div>
				</div>

				<div class="card bg-base-200 border-base-300 border">
					<div class="card-body p-4">
						<h3 class="card-title text-sm">错误响应格式</h3>
						<div class="mockup-code text-xs mt-2">
							<pre><code>&#123;"code": 401, "message": "未认证"&#125;</code></pre>
						</div>
						<div class="mt-3 text-xs text-base-content/70 space-y-1">
							<p><code class="bg-base-300 px-1 rounded">400</code> 请求参数错误</p>
							<p><code class="bg-base-300 px-1 rounded">401</code> 未认证 / Token 无效</p>
							<p><code class="bg-base-300 px-1 rounded">403</code> 权限不足</p>
							<p><code class="bg-base-300 px-1 rounded">404</code> 资源不存在</p>
							<p><code class="bg-base-300 px-1 rounded">409</code> 冲突（已存在 / 已用完）</p>
							<p><code class="bg-base-300 px-1 rounded">410</code> 已过期</p>
							<p><code class="bg-base-300 px-1 rounded">500</code> 服务器内部错误</p>
						</div>
					</div>
				</div>
			</div>
		</div>
	</div>
</div>
