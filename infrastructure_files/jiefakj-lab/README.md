# jiefakj NetBird 本机联调环境

用 Docker 复现"办公/外网用户经 NetBird 访问保密 VLAN"的完整链路,验证与 EasyAuth 授权连接器、Authentik 身份源的三方联调。**仅用于本机联调,不是生产部署。**

## 组件与拓扑

```
Authentik(外部, auth.example.com)         身份源:OIDC + 设备码/PKCE, JWT sub = 用户 uuid
EasyAuth(外部, iam.example.com)           授权源:审批→授权组→NetBird 连接器周期/事件对账推送
                                          经 host.docker.internal:33073 调本栈管理 API
─ 本 compose 栈 ────────────────────────────────────────────────
management(netbird-management-jiefakj:local, 含 P1/P2/P3 补丁) :33073
signal / relay / coturn / dashboard(:8090)
router-peer   vlan1(10.10.1.10) + nb-core   宣告 10.10.1.0/24 路由, masquerade
secret-server vlan1(10.10.1.20)             保密业务(nginx), 仅在 vlan1
office-peer   vlan2(10.10.2.11) + nb-core   张三:已审批, 经隧道访问保密网段
unauth-peer   vlan2(10.10.2.12) + nb-core   李四:未申请, 默认拒绝
```

`vlan1`/`vlan2` 都是 `internal` 网络,office-peer 与 secret-server **无 L3 直连**,唯一通路是 WireGuard 隧道 → router-peer → secret-server。

## 前置

- 镜像 `netbird-management-jiefakj:local` 已构建:
  `docker build -f management/Dockerfile.multistage -t netbird-management-jiefakj:local .`(在仓库根)
- Authentik 已建 `netbird` OIDC 应用(public client、device_code grant、sub_mode=user_uuid、brand 绑定设备码流程)。
- `management.json` 里的 IdP 端点指向 `https://auth.example.com/application/o/netbird/`。
- 真实域名与 IP 写在 `.env`(不入库); `sh render-management-json.sh` 从 `management.json.example` 按 `AUTH_DOMAIN` 生成 `management.json`。

## 启动

```bash
cd infrastructure_files/jiefakj-lab
cp .env.example .env                            # 填三密钥(openssl rand -hex 24 / -hex 16 / -base64 32)与真实域名/IP
cp turnserver.conf.example turnserver.conf      # 把口令换成 .env 的 TURN_PASSWORD
sh render-management-json.sh                    # 按 .env 的 AUTH_DOMAIN 生成 management.json
docker compose up -d
```

management 首启会下载 GeoIP 库,约 30–60s 后 `:33073` 可用。

## Bootstrap NetBird 账户

社区版无内嵌管理员,首个 SSO 登录者成为 owner。用 `ak_device_login.py` 无头完成设备码授权(等价于用户在浏览器登录):

```bash
# 1) 触发设备码, 拿 user_code(见管理端或客户端输出)
# 2) 以某 Authentik 用户完成授权
python3 ak_device_login.py <username> <password> <user_code>
```

之后用该 owner 的 access_token 建 service user + PAT,把 PAT 配到 EasyAuth 连接器实例;删除默认 All↔All 策略,按授权组建网络策略(本栈已用 `DisableDefaultPolicy: true`)。

## 场景矩阵(联调验收)

| 场景 | 操作 | 预期 |
|------|------|------|
| 1 预创建+收养+访问 | EasyAuth 审批张三 → 连接器预创建 NetBird 用户(绑 vpn-secret 组)→ 张三 `netbird up` SSO | 首登原样收养, peer 入组, `curl http://10.10.1.20/` 得保密页 |
| 2 默认拒绝+文案 | 李四未申请, 直接 `netbird up` SSO | JIT 建 Blocked+PendingApproval, 注册被拒, 客户端显示 `NB_BLOCKED_USER_MESSAGE` |
| 3 撤权即断 | EasyAuth 撤销张三 → 事件对账 | 移组+block, 网络图即时重推, 张三访问保密网段超时 |

`ak_device_login.py` 走公网 `auth.example.com`,与真实用户路径一致,仅用标准库。
