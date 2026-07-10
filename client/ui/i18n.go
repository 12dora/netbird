//go:build !(linux && 386)

package main

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

// zhCNTranslations contains every user-facing English source string used by
// the desktop UI. Keeping the English text as the key makes English the
// zero-config fallback and keeps translations close to the UI code that uses
// them.
var zhCNTranslations = map[string]string{
	"NetBird Error": "NetBird 错误",
	"NetBird Settings": "NetBird 设置",
	"NetBird Profiles": "NetBird 配置文件",
	"NetBird Session Expired": "NetBird 会话已过期",
	"Automatically updating client": "正在自动更新客户端",
	"Networks": "网络",
	"Connection": "连接",
	"Network": "网络",
	"SSH": "SSH",
	"Profile": "配置文件",
	"Management URL": "管理地址",
	"Pre-shared Key": "预共享密钥",
	"Quantum-Resistance": "抗量子",
	"Interface Name": "接口名称",
	"Interface Port": "接口端口",
	"MTU": "MTU",
	"Log File": "日志文件",
	"Network Monitor": "网络监测",
	"Disable DNS": "禁用 DNS",
	"Disable Client Routes": "禁用客户端路由",
	"Disable Server Routes": "禁用服务器路由",
	"Disable IPv6": "禁用 IPv6",
	"Disable LAN Access": "禁用局域网访问",
	"Enable SSH Root Login": "启用 SSH Root 登录",
	"Enable SSH SFTP": "启用 SSH SFTP",
	"Enable SSH Local Port Forwarding": "启用 SSH 本地端口转发",
	"Enable SSH Remote Port Forwarding": "启用 SSH 远程端口转发",
	"Disable SSH Authentication": "禁用 SSH 身份验证",
	"JWT Cache TTL (seconds, 0=disabled)": "JWT 缓存 TTL（秒，0=禁用）",
	"If set to 0, a random free port will be used": "设为 0 时将使用随机空闲端口",
	"Enable Rosenpass permissive mode": "启用 Rosenpass 宽松模式",
	"Restarts NetBird when the network changes": "网络发生变化时重启 NetBird",
	"Keeps system DNS settings unchanged": "保持系统 DNS 设置不变",
	"This peer won't route traffic to other peers": "此节点不会向其他节点转发流量",
	"This peer won't act as router for others": "此节点不会充当其他节点的路由器",
	"Disable IPv6 overlay addressing": "禁用 IPv6 覆盖网络地址",
	"Blocks local network access when used as exit node": "作为出口节点时阻止访问本地网络",
	"Save": "保存",
	"Cancel": "取消",
	"Connected": "已连接",
	"Disconnected": "未连接",
	"Connecting": "正在连接",
	"Connect": "连接",
	"Disconnect": "断开连接",
	"Settings": "设置",
	"Allow SSH": "允许 SSH",
	"Allow SSH connections": "允许 SSH 连接",
	"Connect on Startup": "启动时连接",
	"Connect automatically when the service starts": "服务启动时自动连接",
	"Enable Quantum-Resistance": "启用抗量子保护",
	"Enable post-quantum security via Rosenpass": "通过 Rosenpass 启用后量子安全保护",
	"Block Inbound Connections": "阻止入站连接",
	"Block inbound connections to the local machine and routed networks": "阻止到本机和已路由网络的入站连接",
	"Notifications": "通知",
	"Enable notifications": "启用通知",
	"Advanced Settings": "高级设置",
	"Advanced settings of the application": "应用程序高级设置",
	"Create Debug Bundle": "创建调试包",
	"Create and open debug information bundle": "创建并打开调试信息包",
	"Exit Node": "出口节点",
	"Use exit node %s": "使用出口节点 %s",
	"get client: %v": "获取客户端失败：%v",
	"failed to select network: %v": "选择网络失败：%v",
	"failed to deselect network: %v": "取消选择网络失败：%v",
	"failed to select all networks: %v": "选择所有网络失败：%v",
	"failed to deselect all networks: %v": "取消选择所有网络失败：%v",
	"Open the networks management window": "打开网络管理窗口",
	"About": "关于",
	"Download latest version": "下载最新版本",
	"Quit the client app": "退出客户端应用",
	"Quit": "退出",
	"Install version %s": "安装版本 %s",
	"NetBird (Connected)": "NetBird（已连接）",
	"NetBird (Disconnected)": "NetBird（未连接）",
	"NetBird (Connecting)": "NetBird（正在连接）",
	"Daemon: %s": "守护进程：%s",
	"Daemon version: %s": "守护进程版本：%s",
	"GUI: %s": "图形界面：%s",
	"GUI Version: %s": "图形界面版本：%s",
	"Update available": "有可用更新",
	"A new version %s is ready to install": "新版本 %s 已准备好安装",
	"Your NetBird session has expired.\nPlease re-authenticate to continue using NetBird.": "您的 NetBird 会话已过期。\n请重新验证身份以继续使用 NetBird。",
	"Re-authenticate": "重新验证",
	"Waiting login failed, please create \na debug bundle in the settings and contact support.": "等待登录失败，请在设置中创建\n调试包并联系技术支持。",
	"Re-authentication successful.\nReconnecting": "重新验证成功。\n正在重新连接",
	"Already connected.\nClosing this window.": "已经连接。\n正在关闭此窗口。",
	"Reconnecting failed, please create \na debug bundle in the settings and contact support.": "重新连接失败，请在设置中创建\n调试包并联系技术支持。",
	"Connection successful.\nClosing this window.": "连接成功。\n正在关闭此窗口。",
	"All networks": "所有网络",
	"Overlapping networks": "重叠网络",
	"Exit-node networks": "出口节点网络",
	"Refresh": "刷新",
	"Select all": "全选",
	"Deselect All": "取消全选",
	"      ID": "      ID",
	"Range/Domains": "范围/域名",
	"Resolved IPs": "已解析 IP",
	"Select": "选择",
	"Active": "活动中",
	"Deregister": "注销",
	"Remove": "移除",
	"Switch Profile": "切换配置文件",
	"Are you sure you want to switch to '%s'?": "确定要切换到“%s”吗？",
	"Profile Switched": "配置文件已切换",
	"Profile '%s' switched successfully": "已成功切换到配置文件“%s”",
	"Delete Profile": "删除配置文件",
	"Are you sure you want to delete '%s'?": "确定要删除“%s”吗？",
	"Profile Removed": "配置文件已移除",
	"Profile '%s' removed successfully": "已成功移除配置文件“%s”",
	"Profile Created": "配置文件已创建",
	"Profile '%s' created successfully": "已成功创建配置文件“%s”",
	"New Profile": "新建配置文件",
	"Enter Profile Name": "输入配置文件名称",
	"Name:": "名称：",
	"Create": "创建",
	"profile name cannot be empty": "配置文件名称不能为空",
	"Deregistered": "已注销",
	"Are you sure you want to deregister from '%s'?": "确定要从“%s”注销吗？",
	"Successfully deregistered from '%s'": "已成功从“%s”注销",
	"Manage Profiles": "管理配置文件",
	"Profiles are disabled by daemon": "守护进程已禁用配置文件",
	"configuration updates are disabled by daemon": "守护进程已禁用配置更新",
	"MDM-managed": "由 MDM 管理",
	"configured": "已配置",
	"invalid pre-shared key value": "预共享密钥无效",
	"invalid interface port": "接口端口无效",
	"invalid interface port: out of range 0-65535": "接口端口无效：必须在 0 到 65535 范围内",
	"invalid MTU value": "MTU 值无效",
	"MTU must be between %d and %d bytes": "MTU 必须介于 %d 和 %d 字节之间",
	"invalid SSH JWT Cache TTL value": "SSH JWT 缓存 TTL 值无效",
	"SSH JWT Cache TTL must be between 0 and %d seconds": "SSH JWT 缓存 TTL 必须介于 0 和 %d 秒之间",
	"Failed to switch profile": "切换配置文件失败",
	"Failed to deregister": "注销失败",
	"Deregistered successfully": "已成功注销",
	"failed to select profile": "选择配置文件失败",
	"failed to handle down click": "断开连接失败",
	"failed to remove profile": "移除配置文件失败",
	"failed to create profile": "创建配置文件失败",
	"failed to connect to service": "连接服务失败",
	"failed to get current user": "获取当前用户失败",
	"deregister failed": "注销失败",
	"Profile: %s (User: %s)": "配置文件：%s（用户：%s）",
	"Error": "错误",
	"Success": "成功",
	"Connection to service lost": "与服务的连接已断开",
	"Failed to connect": "连接失败",
	"Failed to disconnect": "断开连接失败",
	"Failed to update SSH settings": "更新 SSH 设置失败",
	"Failed to update auto-connect settings": "更新自动连接设置失败",
	"Failed to update Rosenpass settings": "更新 Rosenpass 设置失败",
	"Failed to update block inbound settings": "更新阻止入站连接设置失败",
	"Failed to update notifications settings": "更新通知设置失败",
	"Updating...": "正在更新…",
	"Updating": "正在更新",
	"Your client version is older than the auto-update version set in Management.\nUpdating client to: %s.": "您的客户端版本低于管理端设置的自动更新版本。\n正在更新客户端至：%s。",
	"Update timed out. Please try again.": "更新超时，请重试。",
	"Update canceled.": "更新已取消。",
	"Update failed: %s": "更新失败：%s",
	"Create a debug bundle to help troubleshoot issues with NetBird": "创建调试包以帮助排查 NetBird 问题",
	"NetBird Debug": "NetBird 调试",
	"Anonymize sensitive information (public IPs, domains, ...)": "匿名化敏感信息（公网 IP、域名等）",
	"Include system information (routes, interfaces, ...)": "包含系统信息（路由、网络接口等）",
	"Include packet capture": "包含数据包捕获",
	"Upload bundle automatically after creation": "创建后自动上传调试包",
	"Debug upload URL:": "调试包上传 URL：",
	"Enter upload URL": "输入上传 URL",
	"Run with trace logs before creating bundle": "创建调试包前以跟踪日志运行",
	"for": "持续",
	"minute": "分钟",
	"minutes": "分钟",
	"Note: NetBird will be brought up and down during collection": "注意：收集期间 NetBird 将断开并重新连接",
	"must be a number ≥ 1": "必须是大于或等于 1 的数字",
	"Error: Upload URL is required when upload is enabled": "错误：启用上传时必须提供上传 URL",
	"Error: Invalid duration: %v": "错误：无效时长：%v",
	"Running in debug mode for %d minutes...": "正在以调试模式运行 %d 分钟…",
	"Creating debug bundle...": "正在创建调试包…",
	"Bundle created successfully": "调试包创建成功",
	"Running with trace logs... %s remaining": "正在以跟踪日志运行…剩余 %s",
	"Collecting debug data...": "正在收集调试数据…",
	"Creating debug bundle with collected logs...": "正在使用收集的日志创建调试包…",
	"Error: %v": "错误：%v",
	"Error creating bundle: %v": "创建调试包时出错：%v",
	"Bundle upload failed:\n%s\n\nYou can still access the bundle locally.": "调试包上传失败：\n%s\n\n您仍可在本地访问该调试包。",
	"Bundle upload failed:\n%s\n\nA local copy was saved at:\n%s": "调试包上传失败：\n%s\n\n本地副本保存于：\n%s",
	"Upload Failed": "上传失败",
	"Open File": "打开文件",
	"Open Folder": "打开文件夹",
	"Open file": "打开文件",
	"Open folder": "打开文件夹",
	"Copy key": "复制密钥",
	"open the local file:\n%s\n\nError: %v": "打开本地文件失败：\n%s\n\n错误：%v",
	"open the local folder:\n%s\n\nError: %v": "打开本地文件夹失败：\n%s\n\n错误：%v",
	"Bundle uploaded successfully!": "调试包上传成功！",
	"Upload key:": "上传密钥：",
	"Local copy saved at:\n%s": "本地副本保存于：\n%s",
	"Upload Successful": "上传成功",
	"OK": "确定",
	"Bundle created locally at:\n%s\n\nYou can upload it manually or open the local file.": "调试包已创建于本地：\n%s\n\n您可以手动上传或打开本地文件。",
	"Bundle created locally at:\n%s\n\nAdministrator privileges may be required to access the file.": "调试包已创建于本地：\n%s\n\n访问该文件可能需要管理员权限。",
	"Debug Bundle Created": "调试包已创建",
	"You can always access NetBird from your %s.": "您始终可以从%s访问 NetBird。",
	"menu bar": "菜单栏",
	"system tray": "系统托盘",
}

func tr(message string) string {
	return translate(systemLocale(), message)
}

func trf(format string, args ...any) string {
	return fmt.Sprintf(tr(format), args...)
}

func translate(locale, message string) string {
	if !isChineseLocale(locale) {
		return message
	}

	if translated, ok := zhCNTranslations[message]; ok {
		return translated
	}
	return message
}

func isChineseLocale(locale string) bool {
	locale = strings.ToLower(strings.ReplaceAll(strings.TrimSpace(locale), "_", "-"))
	return strings.HasPrefix(locale, "zh") || strings.Contains(locale, "hans") || strings.Contains(locale, "hant")
}

func systemLocale() string {
	for _, name := range []string{"LC_ALL", "LC_MESSAGES", "LANGUAGE", "LANG"} {
		if locale := os.Getenv(name); locale != "" {
			return strings.Split(locale, ":")[0]
		}
	}

	var command *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		command = exec.Command("defaults", "read", "-g", "AppleLocale")
	case "windows":
		command = exec.Command("powershell", "-NoProfile", "-Command", "(Get-Culture).Name")
	}
	if command == nil {
		return ""
	}

	output, err := command.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(output))
}
