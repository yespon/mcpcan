
<div align="center">
<img width="1872" height="932" alt="image" src="https://github.com/user-attachments/assets/2502f1af-7f0c-400e-a683-e62bb785d42d" />

</div>

<div align="center">

# MCP CAN
一款开源集中管理MCP服务的平台</br>
MCPCAN采用容器技术实现MCP服务的灵活部署，有效解决潜在的系统配置冲突。它支持多协议兼容与转换，可实现不同MCP服务架构之间的无缝集成。此外，它还提供可视化监控、安全认证和一站式部署功能。</br>

  <img src="https://img.shields.io/badge/Vue.js-3.2.47-4FC08D?style=for-the-badge&logo=vue.js&logoColor=ffffff" alt="Vue.js"/>
  <img src="https://img.shields.io/badge/TypeScript-5.0-blue?style=for-the-badge&logo=typescript" alt="TypeScript"/>
  <img src="https://img.shields.io/badge/MySQL-8.0-blue?style=for-the-badge&logo=mysql" alt="MySQL"/>
  <img src="https://img.shields.io/badge/Kubernetes-1.28-326ce5?style=for-the-badge&logo=kubernetes" alt="Kubernetes"/>
  <img src="https://img.shields.io/badge/License-GPL--3.0-blue?style=for-the-badge" alt="GPL-3.0"/>
<p align="center">
    <a href="https://demo.mcpcan.com" target="_blank">
        <img alt="Static Badge" src="https://img.shields.io/badge/Product-F04438"></a>
    <a href="https://dify.ai/pricing" target="_blank">
        <img alt="Static Badge" src="https://img.shields.io/badge/free-pricing?logo=free&color=%20%23155EEF&label=pricing&labelColor=%20%23528bff"></a>
    <a href="https://discord.gg/EegGj7G7Bz" target="_blank">
        <img src="https://img.shields.io/discord/1428637640856571995?logo=discord&labelColor=%20%235462eb&logoColor=%20%23f5f5f5&color=%20%235462eb"
            alt="chat on Discord"></a>
    <a href="https://twitter.com/intent/follow?screen_name=MCPCAN" target="_blank">
        <img src="https://img.shields.io/twitter/follow/MCPCAN?logo=X&color=%20%23f5f5f5"
            alt="follow on X(Twitter)"></a>
</p>

<p align="center">
   <a href="./README.md">English</a> | <strong>中文版</strong> <br>
</p>
</div>

## 🎉介绍
MCPCan 是一个专注于高效管理 MCP（模型上下文协议）服务的开源平台，通过现代化的 Web 界面，为 DevOps 和开发团队提供全面的 MCP 服务生命周期管理功能。
MCPCan 支持多协议兼容和转换，能够实现不同 MCP 服务架构之间的无缝集成，同时提供可视化监控、安全认证和一站式部署功能。<br/>

## ✨ 核心功能

- 🎯 统一管理：集中管理所有 MCP 服务实例及配置项
- 🔄 协议转换：支持多种 MCP 协议间无缝互转
- 📊 实时监控：提供详尽的服务状态与性能监控数据
- 🔐 安全认证：内置身份认证与权限管理体系
- 🚀 一站式部署：MCP 服务快速发布、配置与分发
- 📈 可扩展性：基于 Kubernetes 的云原生架构

## ✨演示和官网
为了获得最佳演示体验，请尝试直接  <a href="https://demo.mcpcan.com">DemoSite : demo.mcpcan.com</a> .<br>
[MP4]<br>
要查看我们的官方网站地址，只需点击 <a href="https://www.mcpcan.com">MainSite : www.mcpcan.com</a>.即可
## 👨‍🚀快速开始

有关详细部署说明，请参阅我们的[部署指南](https://kymo-mcp.github.io/mcpcan-deploy/).

```bash
# Install Helm Chart repository
helm repo add mcpcan https://kymo-mcp.github.io/mcpcan-deploy/

# Update Helm repository
helm repo update mcpcan

# Install latest version
helm install mcpcan mcpcan/mcpcan-deploy

# Deploy with public IP
helm install mcpcan mcpcan/mcpcan-deploy \
  --set global.publicIP=192.168.1.100 \
  --set infrastructure.mysql.auth.rootPassword=secure-password \
  --set infrastructure.redis.auth.password=secure-password \
  --namespace mcpcan --create-namespace --timeout 600s --wait

# Deploy with domain name
helm install mcpcan mcpcan/mcpcan-deploy \
  --set global.domain=mcp.example.com \
  --set infrastructure.mysql.auth.rootPassword=secure-password \
  --set infrastructure.redis.auth.password=secure-password \
  --namespace mcpcan --create-namespace --timeout 600s --wait
```

## 🚀组件

MCP CAN 由多个关键组件构成，这些组件共同构成 MCP CAN 的功能框架，为用户提供全面的 MCP 服务管理功能。

| Project                                | Status                                                      | Description                                |
| -------------------------------------- | ----------------------------------------------------------- | ------------------------------------------ |
| [MCPCan-Web](frontend/)                | ![Status](https://img.shields.io/badge/status-active-green) | MCPCan Web UI (Vue.js Frontend)            |
| [MCPCan-Backend](backend/)             | ![Status](https://img.shields.io/badge/status-active-green) | MCPCan Backend Services (Go Microservices) |
| [MCPCan-Gateway](backend/cmd/gateway/) | ![Status](https://img.shields.io/badge/status-active-green) | MCP Gateway Service                        |
| [MCPCan-Market](backend/cmd/market/)   | ![Status](https://img.shields.io/badge/status-active-green) | MCP Service Marketplace                    |
| [MCPCan-Authz](backend/cmd/authz/)     | ![Status](https://img.shields.io/badge/status-active-green) | Authentication and Authorization Service   |

## 🐧技术栈

### 🐧前端

- **Framework**: Vue.js 3.5+ (Composition API)
- **Language**: TypeScript
- **Styling**: UnoCSS, SCSS
- **UI Components**: Element Plus
- **State Management**: Pinia
- **Build Tool**: Vite

### 🐧后端

- **Language**: Go 1.24.2+
- **Framework**: Gin, gRPC
- **Database**: MySQL, Redis
- **Container**: Docker, Kubernetes

## 🐧第三方项目

- [mcpcan-deploy](https://github.com/Kymo-MCP/mcpcan-deploy) - Official Helm charts source repository for MCPCan
- [MCPCan Helm Charts](https://kymo-mcp.github.io/mcpcan-deploy/) - Official Helm charts repository for MCPCan

## 💝贡献💝

Welcome to submit PR to contribute. Please refer to [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

Before contributing, please:

1. Read our [Code of Conduct](CODE_OF_CONDUCT.md)
2. Check existing issues and pull requests
3. Follow our coding standards and commit message conventions

## 安全

If you discover a security vulnerability, please refer to our [Security Policy](SECURITY.md) for responsible disclosure guidelines.

## 证书

Copyright (c) 2024-2025 MCPCan Team, All rights reserved.

Licensed under The GNU General Public License version 3 (GPLv3) (the "License"); you may not use this file except in compliance with the License. You may obtain a copy of the License at

https://www.gnu.org/licenses/gpl-3.0.html

Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the specific language governing permissions and limitations under the License.

## 社区与支持

- 📖 [Documentation](https://kymo-mcp.github.io/mcpcan-deploy/)
- 💬 [Discord Community](https://discord.com/channels/1428637640856571995/1428637896532820038)
- 🐛 [Issue Tracker](https://github.com/Kymo-MCP/mcpcan/issues)
- 📧 [Mailing List](mailto:opensource@kymo.cn)

## 致谢

- Thanks to the [MCP Protocol](https://modelcontextprotocol.io/) community
- Thanks to all contributors and supporters
- Special thanks to the open-source projects that make MCPCan possible
