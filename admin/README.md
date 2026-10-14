# 请假管理平台管理后台

本目录是 Leave 项目的 Web 管理后台，基于 SoybeanAdmin、Vue 3、Vite、TypeScript、Pinia、Vue Router、Naive UI 和 UnoCSS 构建。

## 项目结构

```text
admin/
├── build/                 # Vite 构建和插件配置
├── packages/              # 项目内部可复用包
├── public/                # 不参与构建处理的静态资源
├── src/
│   ├── components/        # 通用组件
│   ├── layouts/            # 后台布局
│   ├── router/             # 路由和菜单
│   ├── service/            # API 请求和接口类型
│   ├── store/              # Pinia 状态管理
│   ├── views/              # 业务页面
│   ├── locales/            # 国际化资源
│   └── styles/             # 全局样式
├── .env                   # 开发环境配置
├── .env.prod              # 生产环境配置
├── .env.test              # 测试环境配置（本地文件，不提交）
├── .env.example           # 所有环境共用的完整配置模板
├── package.json
├── pnpm-workspace.yaml
└── vite.config.ts
```

## 开发命令

在本目录执行：

```bash
pnpm install       # 安装依赖
pnpm dev           # 启动开发环境
pnpm dev:test      # 本地测试模式（联调时另行启用开发代理）
pnpm build         # 构建生产版本
pnpm build:test    # 构建测试版本到 dist-test/
pnpm preview       # 预览构建结果
pnpm preview:test  # 预览 dist-test/ 中的测试版本
pnpm typecheck     # TypeScript 类型检查
pnpm lint          # ESLint、oxlint 检查并修复
pnpm fmt           # 使用 oxfmt 格式化
```

## 环境配置

所有环境共用 `.env.example` 一份完整模板，`.env`、`.env.test` 和 `.env.prod` 保持相同字段结构，仅按环境调整配置值。Vite 仍会先加载 `.env`，再用当前模式的文件覆盖同名字段。已联调功能通过现有请求封装访问 Leave 后端，各页面保持当前真实请求或 Mock 的开发阶段。修改 `VITE_SERVICE_BASE_URL` 切换后端，不在业务页面中硬编码地址。

首次使用时，将 `.env.example` 分别复制为 `.env` 和 `.env.test`。测试文件将 `VITE_SERVICE_BASE_URL` 改为 `/api/v1`，`VITE_SERVICE_PROXY_TARGET` 留空，将 `VITE_HTTP_PROXY` 设为 `N`，`VITE_STORAGE_PREFIX` 设为 `leave-admin-test`；标题按需标识测试环境。其他字段与默认环境保持一致。环境文件已由根目录 `.gitignore` 忽略，只提交统一的 `.env.example` 模板。

模板中的字段用于页面路由、图标插件、API 响应处理、开发代理、登录存储及权限菜单。它们仍有代码引用；不使用的 `VITE_APP_DESC` 已移除。`.npmrc` 配置依赖下载源，当前使用 `https://registry.npmmirror.com/`，与应用运行时的 `.env` 无关。

本地测试开发和预览端口为 `9001`；服务器管理端通过 `https://admin-test.leave.mhspace.dpdns.org` 访问，测试后端端口为 `9000`。当前 `.env.test` 面向服务器静态部署：`VITE_SERVICE_BASE_URL=/api/v1`、`VITE_SERVICE_PROXY_TARGET=`、`VITE_HTTP_PROXY=N`。浏览器使用同域 API，由 Nginx 转发到 `http://127.0.0.1:9000`。测试模式使用独立存储前缀，避免复用默认开发环境的登录 Token。

如果需要本地 `pnpm dev:test` 联调，将开发代理目标填为 `http://39.99.33.228:9000/api/v1` 并将 `VITE_HTTP_PROXY` 设为 `Y`；部署构建不需要这些代理配置。

`pnpm build:test` 将测试产物输出到 `dist-test/`，不覆盖生产构建的 `dist/`。构建后的浏览器请求使用同域 `/api/v1`，开发代理地址只用于本地开发。`pnpm preview:test` 在本机 9001 预览静态文件，但不会执行 Vite 开发代理；本地预览仅检查页面，完整 API 联调使用部署后的 Nginx 网站，或按上面的说明启用本地开发代理。

## 宝塔测试站点部署

1. 在 `admin/` 执行 `pnpm build:test`。
2. 使用宝塔文件管理上传 `dist-test/` **里面的内容**到 `/www/wwwroot/Leave_test/admin/releases/v1/`，确保该目录直接包含 `index.html` 和 `assets/`。
3. 首次部署时，在宝塔终端执行以下命令创建相对软链接。若 `current` 已存在，先确认指向再切换，不覆盖未知目录。

   ```bash
   cd /www/wwwroot/Leave_test/admin
   ln -s releases/v1 current
   ```

4. 在宝塔的网站管理中添加 HTML 项目，绑定 `admin-test.leave.mhspace.dpdns.org`，根目录为 `/www/wwwroot/Leave_test/admin/current`，FTP 选择不创建。HTTP 使用 80 端口，在 SSL 页面配置证书并启用 HTTPS。在站点 Nginx 配置的 `server` 块中配置以下内容；已有相同 location 时修改原块，不重复添加。

   ```nginx
   location / {
       try_files $uri $uri/ /index.html;
   }

   location ^~ /api/v1/uploads/avatars/ {
       alias /www/wwwroot/Leave_test/uploads/avatars/;
       autoindex off;
       expires 7d;
   }

   location ^~ /api/v1/ {
       proxy_pass http://127.0.0.1:9000;
       proxy_set_header Host $host;
       proxy_set_header X-Real-IP $remote_addr;
       proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
       proxy_set_header X-Forwarded-Proto $scheme;
   }
   ```

   `proxy_pass` 的地址结尾不加 `/`，保留请求的 `/api/v1/` 前缀。history 路由通过 `try_files` 支持子页面刷新。
5. 放行网站使用的 TCP 80/443，保存配置并通过宝塔重载 Nginx。后端继续由宝塔 Go 项目在 9000 启动；同机反向代理不要求对公网开放 9000。检查后端测试配置的 `storage.base_url` 为 `https://admin-test.leave.mhspace.dpdns.org/api/v1/uploads`，若修改则通过宝塔重启测试后端。
6. 访问 `https://admin-test.leave.mhspace.dpdns.org/`，并检查 `/api/v1/health` 返回成功，然后验证登录、头像和刷新子页面。

服务器监听由 Nginx 配置决定，前端 `.env` 不能自动设置服务器监听端口。服务器只需上传静态产物，无需上传前端 `.env` 或运行 Node/Vite 服务。后续更新上传新的构建产物即可；域名、后端端口和目录不变时无需调整 Nginx。当前采用宝塔手动部署，未改造 CI/CD。

## 与项目其他部分的关系

```text
frontend/   # uni-app x 用户端
admin/      # Web 管理后台
backend/    # Go 后端服务
```

管理后台与用户端是两个独立的前端应用，共享后端接口，但不要把 uni-app x 的页面或组件直接移植到本目录。
