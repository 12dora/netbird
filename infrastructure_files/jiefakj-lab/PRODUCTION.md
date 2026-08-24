# NetBird 生产部署方案(jiefakj)

面向"办公/外网员工经 NetBird 访问保密 VLAN、准入由 EasyAuth 授权、身份由 Authentik/钉钉"的真实业务。本文给出与联调栈一致、但满足生产要求的部署基线。方案已在本机 Docker 栈跑通场景 1/2/3(预创建收养、默认拒绝+文案、撤权即断)。

## 0. 架构定位(三方各司其职)

| 系统 | 角色 | 边界 |
|------|------|------|
| Authentik(`auth.example.com`) | **唯一身份源** | 钉钉 OIDC 登录;NetBird 客户端走设备码/PKCE 认证。JWT `sub` = Authentik 用户 uuid,不承载任何授权语义 |
| EasyAuth(`iam.example.com`) | **唯一授权源** | 门户申请→站内审批→授权组;NetBird 连接器周期/事件对账,主动写 NetBird 管理 API(预创建/绑组/block) |
| NetBird management(fork) | **网络执行点** | 只认 EasyAuth 推来的授权组;JWT 组同步关闭;身份不落本地 |

授权事实只从 EasyAuth 流出,NetBird 侧手工改动会被下一轮对账矫正(映射管理范围内)。

## 1. 镜像与 fork

- 仅改 **management** 服务端,客户端/dashboard/signal/relay 用官方镜像。本 fork 三个补丁:
  - **P1** `POST /api/users` 支持显式 `id` 预创建常规用户(`0c83bc5f`)——消除外接 IdP 的 JIT gap,审批通过即权限就位。
  - **P2** `NB_BLOCKED_USER_MESSAGE` 可配置被拦文案(`e931f089`)——未授权用户看到申请入口。
    覆盖两条路径:peer 注册(`management/server/peer.go`)与反向代理 SSO 拒绝页
    (`management/server/http/handlers/proxy/auth.go`,上游 v0.77.1 新增),
    实现收敛在 `management/server/blockeduser`。仅覆盖账号状态类拒绝,
    签名/查询失败仍返回上游的通用文案,不被改写。
  - **P3** OIDC 发现重试——启动时 IdP `/.well-known/openid-configuration` 暂时不可达则 15s 超时 + 指数退避重试(预算 10 分钟),避免 crash loop。
- **基线**:锁最新稳定 tag(当前 `v0.77.1`),不跟 main。升级 = 合入新 tag 并化解冲突 → 跑 `management/server/http/handlers/users` 与 `management/server` 单测 → 重建镜像 → 本机联调复验场景 1/2/3 → 上生产。
- **构建**:`docker build -f management/Dockerfile.multistage -t registry.example.com/netbird-management:v0.77.1-jiefakj.1 .`,推私有 registry,生产按不可变 tag 引用(勿用 `:local`/`:latest`)。
- **客户端**:v0.77.1 起上游桌面 UI 由 Fyne 改为 Wails v3 + React,并自带简体中文
  (`client/ui/i18n/locales/zh-CN/common.json`,与 en 453/453 键对齐,已在语言选择器注册)。
  本 fork 早期自维护的 Fyne 中文补丁已随之删除,桌面包直接用上游 release 产物;
  本仓库的 `publish-desktop-packages.yml` 只再发布 Linux CLI deb。
- **AGPL 合规**(`management/` 为 AGPL-3.0):公司内部自用无触发义务;**若将该管理面作为服务提供给公司外部用户,须依 AGPL 公开含补丁的源码**。建议私有 fork 保留完整 LICENSE,补丁面刻意最小以便审计与合规。

## 2. 部署拓扑

```
                 公网
  netbird.example.com  ──►  反代/LB(TLS 终止)──► management :443 + dashboard
  signal.example.com   ──►                        signal(gRPC over TLS)
  relay.example.com    ──►                        relay(rel:// 或 rels://)
  turn.example.com     ──►  coturn(公网 IP, UDP 3478 + 中继端口段)
  auth.example.com     ──►  Authentik(已有)
  iam.example.com      ──►  EasyAuth(已有)

  保密 VLAN(如 10.10.1.0/24)内部署 1–2 台 NetBird routing peer(高可用双 metric),
  宣告网段路由 + masquerade;保密业务本身不装客户端。
```

- 客户端要能到达 management(443)、signal、relay、turn。TURN 需公网可达的 UDP(NAT 穿透兜底)。
- routing peer 放在保密 VLAN 内,是办公/外网用户进入保密网段的唯一受控入口;双 peer 不同 metric 做故障切换。

## 3. management 配置基线(management.json)

| 配置 | 值 | 原因 |
|------|-----|------|
| `IdpManagerConfig.ManagerType` | **`none`** | 授权与目录都由 EasyAuth 推送,NetBird 无需回调 Authentik 管理目录(其 authentik IdP-manager 的 CreateUser 本就是 stub)。省去 service account 口令,减小暴露面。用户 name/email 由 P1 预创建写入、JIT 从 JWT claims 补齐 |
| `HttpConfig.AuthIssuer` | `https://auth.example.com/application/o/netbird/` | OIDC 校验 |
| `HttpConfig.AuthAudience` | `netbird` | = OIDC client_id |
| `HttpConfig.AuthUserIDClaim` | `sub` | `sub` = Authentik uuid = EasyAuth `authentik_user_id` = NetBird 用户 ID,三方零成本对齐 |
| `HttpConfig.OIDCConfigEndpoint` | `.../netbird/.well-known/openid-configuration` | 自动发现端点 |
| `DeviceAuthorizationFlow` / `PKCEAuthorizationFlow` | client_id=`netbird`, scope=`openid profile email offline_access` | 客户端登录;桌面优先 PKCE,回退设备码 |
| `DisableDefaultPolicy` | **`true`** | 不删默认 All↔All 则任何设备全网互通,EasyAuth 管不住(硬前提) |
| `StoreConfig.Engine` | **`postgres`**(生产) | 联调用 sqlite;生产用 Postgres 便于备份/HA。`NETBIRD_STORE_ENGINE_POSTGRES_DSN` 注入 |
| `DataStoreEncryptionKey` | 强随机、密管托管 | 加密敏感列 |
| TURN/Relay Secret | 强随机、公网端点 | 真实 NAT 穿透 |

账户级设置(经管理 API,非 json):`UserApprovalRequired=true`(默认拒绝)、`GroupsPropagationEnabled=true`(组变更回溯已注册设备)、Peer login expiration **12–24h**(撤权兜底:即便 block 前有存活会话,到期强制重认证)。**JWT 组同步保持关闭**(授权单一来源)。

## 4. Authentik 侧(一次性)

- 建 `netbird` OAuth2 Provider:**public client**、`client_id=netbird`、`sub_mode=user_uuid`、`include_claims_in_id_token=true`、scope 映射含 `openid/profile/email/offline_access`;`grant_types` 必须含 `urn:ietf:params:oauth:grant-type:device_code`。
- Application `netbird` 绑定该 provider。
- **Brand 绑定设备码流程**(`flow_device_code`),否则 `/device` 入口 404、设备码登录不可用。
- redirect_uris 覆盖客户端 PKCE 回调(`http://localhost:53000` 等)与 dashboard 回调。

## 5. EasyAuth 侧(一次性 + 日常)

一次性(控制台图形化或脚本):
1. NetBird 建**专用 service user(admin)+ PAT**,PAT 交 EasyAuth 连接器配置(superuser-only、加密落库);PAT 定期轮换。
2. EasyAuth 建 App `netbird` + 授权组(如 `vpn-secret` 保密网段、按需 `vpn-office` 等),标记 requestable,配审批规则。
3. NetBird 侧建对应组 + **网络策略**(仅对映射组放行到保密资源组),删除默认策略。
4. 连接器实例:`api_url=https://netbird.example.com`、`precreate_users=true`、`block_users_without_grant=true`;映射表 `vpn-secret → NetBird 组 ID`;首轮对账绿灯。

日常(全自动):员工门户申请 → 审批通过 → `GrantService` 产生授权事实 → 连接器事件对账(5s 去抖)预创建/绑组 → 员工装客户端 SSO 登录即通;撤权/离职 → 移组+block,网络图即时重推(离职走秒级 block 快路径)。

## 6. 关键运维项

- **备份**:Postgres(management 数据)、`DataStoreEncryptionKey`、各 Secret、PAT 全部纳入密管;定期演练恢复。
- **高可用**:management 单实例 + Postgres 主从起步;routing peer 双机不同 metric。signal/relay 可水平扩。
- **监控**:management/coturn 健康探针;EasyAuth 连接器 `consecutive_failures>=3` 进健康面板;peer 在线数、对账 run 成功率。
- **证书**:management/signal/dashboard 走公司 CA 或 Let's Encrypt;coturn 生产建议 TLS(turns://)。
- **最小权限**:service user 仅管理 API 用途;PAT 轮换;`DataStoreEncryptionKey` 与 DSN 经密管注入,不落盘明文。

## 7. 升级流程(锁定)

1. 新版本发布 → 合入新稳定 tag(`git merge vX.Y.Z`)并化解冲突。补丁面仅限 `management/`,
   冲突通常只出现在 P3 涉及的 `management/cmd/management_test.go`。
   注意:该文件里 `OIDCConfigEndpoint` 必须保持为空字符串 —— P3 让 `LoadMgmtConfig` 在端点非空时
   按 10 分钟预算重试 OIDC 发现,填真实端点会使单测挂起;上游后续新增的同类用例也需一并置空。
2. 跑 `go test ./management/server/http/handlers/users/ ./management/server/ -run 'PreCreate|BlockedUser'`。
3. 重建并推不可变 tag 镜像。
4. 本机联调栈(`infrastructure_files/jiefakj-lab/`)复验场景 1/2/3。
5. 灰度上生产,观察对账与登录。

## 8. 已验证结论(本机联调)

- 场景 1:EasyAuth 审批 → 连接器预创建 NetBird 用户(绑 `vpn-secret`)→ 张三首次 SSO 登录被原样收养 → peer 入组 → 经隧道 `curl http://10.10.1.20/` 得保密页(office-peer 与保密网段无 L3 直连,唯一通路是 WireGuard)。
- 场景 2:李四未申请直接登录 → JIT 建 `Blocked+PendingApproval` → 注册被拒,客户端显示 `无 VPN 权限，请前往 https://iam.example.com 申请`(P2 文案经服务端返回、客户端原样显示,未改客户端)。
- 场景 3:撤销张三授权 → 事件对账 `groups_removed=1, users_blocked=1` → 张三访问保密网段超时(`http_code=000`),隧道断开。
