## Introduction

FreeGFW is a high-performance proxy service management system built on Sing-box and Xray core. It provides a modern web interface for easily deploying, managing, and monitoring various proxy protocol services.

### Features

- High Performance Core: Built on Sing-box and Xray, supporting the latest proxy protocols
- Multi-Protocol Support: VLESS (Reality/Vision), VMess, Shadowsocks, Hysteria2, etc.
- Built-in WARP: One-click Cloudflare WARP integration
- User Management: Multi-user system with independent configurations
- Traffic Monitoring: Real-time monitoring of server and user traffic
- Auto HTTPS: Integrated Let's Encrypt for automatic SSL certificate management

### HTTPS Configuration (Recommended)

If port 80 on the host is already occupied (e.g., by 1Panel OpenResty), FreeGFW's built-in Let's Encrypt feature cannot obtain certificates. Instead, use 1Panel's reverse proxy:

1. **1Panel → Websites → Create Website**, select **Reverse Proxy**
2. Enter your domain, set target URL to `http://127.0.0.1:8080`
3. Enable **Auto HTTPS**, 1Panel will automatically obtain and renew SSL certificates
4. Access FreeGFW via `https://your-domain`, without using the built-in Let's Encrypt
