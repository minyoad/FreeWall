## 产品介绍

FreeGFW 是一个基于 Sing-box 和 Xray 核心构建的高性能代理服务管理系统。它提供了现代化的 Web 界面，用于轻松部署、管理和监控各类代理协议服务。

### 主要功能

- 高性能核心：基于 Sing-box 和 Xray，支持最新的代理协议
- 多协议支持：VLESS (Reality/Vision)、VMess、Shadowsocks、Hysteria2 等
- 一键 WARP 集成：内置 Cloudflare WARP 配置
- 用户管理：多用户系统，独立配置
- 流量监控：实时监控服务器和用户流量
- 自动 HTTPS：集成 Let's Encrypt 自动申请和续期 SSL 证书

### 配置 HTTPS（推荐）

如果宿主机 80 端口已被占用（如 1Panel OpenResty），FreeGFW 内置 Let's Encrypt 无法使用，可以通过 1Panel 反向代理解决：

1. **1Panel → 网站 → 创建网站**，选择**反向代理**
2. 域名填入你的域名，目标地址 `http://127.0.0.1:8080`
3. 开启**自动 HTTPS**，1Panel 会自动申请和续期 SSL 证书
4. 通过 `https://你的域名` 访问 FreeGFW，不再使用内置 Let's Encrypt
