# MCPCAN：A lightweight MCP service management platform built on a containerized architecture.

<p align="Left">
   <strong>English</strong> | <a href="./README_CN.md">中文版</a> 
</p>
<div align="center">
  <img src="https://img.shields.io/badge/Vue.js-3.2.47-4FC08D?style=for-the-badge&logo=vue.js&logoColor=ffffff" alt="Vue.js"/>
  <img src="https://img.shields.io/badge/TypeScript-5.0-blue?style=for-the-badge&logo=typescript" alt="TypeScript"/>
  <img src="https://img.shields.io/badge/MySQL-8.0-blue?style=for-the-badge&logo=mysql" alt="MySQL"/>
  <img src="https://img.shields.io/badge/Kubernetes-1.28-326ce5?style=for-the-badge&logo=kubernetes" alt="Kubernetes"/>
  <img src="https://img.shields.io/badge/License-GPL--3.0-blue?style=for-the-badge" alt="GPL-3.0"/>
</div>
<img width="1836" height="912" alt="image" src="https://github.com/user-attachments/assets/cc706fe0-f53a-464c-b8d7-c336fec9802e" />


## What is MCPCan ?

MCPCan is an open-source platform focused on efficient management of MCP (Model Context Protocol) services, providing DevOps and development teams with comprehensive MCP service lifecycle management capabilities through a modern web interface.

MCPCan supports multi-protocol compatibility and conversion, enabling seamless integration between different MCP service architectures while providing visual monitoring, security authentication, and one-stop deployment capabilities.

<div align="center">
</div>

## ✨ Key Features

- **🎯 Unified Management**: Centralized management of all MCP service instances and configurations
  1. Supports visualized management and control of the entire lifecycle of MCP service instances, covering centralized operations such as creation, startup, stop, destruction, and restart. It provides a visual operation panel and progress tracking, eliminating the need for logging into multiple nodes for distributed maintenance.

  2. Currently, it supports instance categorization based on container status and MCP protocol. Future updates will expand to include environment and cluster, business line, and service type grouping dimensions to achieve fine-grained categorization and efficient management of instances.
- **🔄 Protocol Conversion**: Supports seamless conversion between various MCP protocols
  1. **Direct Connection Mode:** As a lightweight point-to-point conversion solution, it is directly deployed between the devices/systems at both ends of MCP communication, eliminating the need for intermediate forwarding nodes. It achieves real-time conversion of MCP protocol versions and data formats with a minimalist architecture. Its core function is to reduce conversion latency and deployment complexity, adapting to small-scale, low-latency MCP interoperability scenarios.

  2. **Managed Mode:** This mode encapsulates protocol conversion capabilities as a clustered managed service. Its core function is to achieve centralized conversion management of large-scale MCP terminals. Through cluster load balancing, unified rule configuration, and dynamic updates, it efficiently handles the protocol conversion needs of massive numbers of MCP devices, ensuring consistency and scalability for interoperability between different MCP protocol versions.

  3. **Proxy Mode:** This mode integrates protocol conversion and network proxy capabilities. Its core function is to solve MCP communication problems in complex network environments. It achieves cross-network segment and cross-firewall MCP data forwarding and protocol conversion through intermediate proxy nodes, while also adding access authentication, traffic control, and transmission encryption functions, balancing interoperability and security.
- **📊 Real-time Monitoring**: Provides detailed service status and performance monitoring
- **🔐 Security & Authentication**: Built-in identity authentication and permission management system
- **🚀 One-stop Deployment**: Quick release, configuration, and distribution of MCP services
- **📈 Scalability**: Cloud-native architecture based on Kubernetes

## DEMO Site (Under Construction)

MCPCan provides an online demo site where you can experience MCPCan's features and performance.

Under construction...

## Quickstart

For detailed deployment instructions, please refer to our [Deployment Guide](https://kymo-mcp.github.io/mcpcan-deploy/).

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
  --set infrastructure.redis.auth.password=secure-password

# Deploy with domain name
helm install mcpcan mcpcan/mcpcan-deploy \
  --set global.domain=mcp.example.com \
  --set infrastructure.mysql.auth.rootPassword=secure-password \
  --set infrastructure.redis.auth.password=secure-password
```

## Components

MCPCan consists of multiple key components, which collectively form the functional framework of MCPCan, providing users with comprehensive MCP service management capabilities.

| Project                                | Status                                                      | Description                                |
| -------------------------------------- | ----------------------------------------------------------- | ------------------------------------------ |
| [MCPCan-Web](frontend/)                | ![Status](https://img.shields.io/badge/status-active-green) | MCPCan Web UI (Vue.js Frontend)            |
| [MCPCan-Backend](backend/)             | ![Status](https://img.shields.io/badge/status-active-green) | MCPCan Backend Services (Go Microservices) |
| [MCPCan-Gateway](backend/cmd/gateway/) | ![Status](https://img.shields.io/badge/status-active-green) | MCP Gateway Service                        |
| [MCPCan-Market](backend/cmd/market/)   | ![Status](https://img.shields.io/badge/status-active-green) | MCP Service Marketplace                    |
| [MCPCan-Authz](backend/cmd/authz/)     | ![Status](https://img.shields.io/badge/status-active-green) | Authentication and Authorization Service   |

## Technology Stack

### Frontend

- **Framework**: Vue.js 3.5+ (Composition API)
- **Language**: TypeScript
- **Styling**: UnoCSS, SCSS
- **UI Components**: Element Plus
- **State Management**: Pinia
- **Build Tool**: Vite

### Backend

- **Language**: Go 1.24.2+
- **Framework**: Gin, gRPC
- **Database**: MySQL, Redis
- **Container**: Docker, Kubernetes

## Third-party Projects

- [mcpcan-deploy](https://github.com/Kymo-MCP/mcpcan-deploy) - Official Helm charts source repository for MCPCan
- [MCPCan Helm Charts](https://kymo-mcp.github.io/mcpcan-deploy/) - Official Helm charts repository for MCPCan

## Contributing

Welcome to submit PR to contribute. Please refer to [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

Before contributing, please:

1. Read our [Code of Conduct](CODE_OF_CONDUCT.md)
2. Check existing issues and pull requests
3. Follow our coding standards and commit message conventions

## Security

If you discover a security vulnerability, please refer to our [Security Policy](SECURITY.md) for responsible disclosure guidelines.

## License

Copyright (c) 2024-2025 MCPCan Team, All rights reserved.

Licensed under The GNU General Public License version 3 (GPLv3) (the "License"); you may not use this file except in compliance with the License. You may obtain a copy of the License at

https://www.gnu.org/licenses/gpl-3.0.html

Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the specific language governing permissions and limitations under the License.

## Community & Support

- 📖 [Documentation](https://kymo-mcp.github.io/mcpcan-deploy/)
- 💬 [Discord Community](https://discord.com/channels/1428637640856571995/1428637896532820038)
- 🐛 [Issue Tracker](https://github.com/Kymo-MCP/mcpcan/issues)
- 📧 [Mailing List](mailto:opensource@kymo.cn)

## Acknowledgements

- Thanks to the [MCP Protocol](https://modelcontextprotocol.io/) community
- Thanks to all contributors and supporters
- Special thanks to the open-source projects that make MCPCan possible
