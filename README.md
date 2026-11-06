# Leave

A leave-slip simulator with an Android app, a web admin panel and a Go backend. Built for learning and demonstrating mobile interactions, API development and admin workflows.

**English** · [简体中文](README_zh_CN.md)

[Repository](https://github.com/X-LMH/Leave) · [GitHub Actions](https://github.com/X-LMH/Leave/actions) · [Run from source](#run-from-source) · [Documentation](#documentation)

## Features

- Account registration, login, profile editing and password changes.
- Leave applications, leave-slip details and leave history.
- Light and dark themes, appearance settings, feedback and app updates.
- Admin tools for student accounts, leave records, classes, dormitories, leave types, feedback, version records and overview statistics.
- Server-side login sessions and request rate limiting backed by Redis.

Development is ongoing. Some features still use mock or placeholder data, and integration progress varies by module.

## Tech stack

| Module | Technologies |
| --- | --- |
| Android app | uni-app x, UTS, uvue, HBuilderX |
| Web admin | Vue 3, TypeScript, Vite, Naive UI, UnoCSS, Pinia |
| Backend | Go, Gin, GORM, Zap |
| Storage | MySQL for business data; Redis for sessions and rate limiting |

The admin panel is based on Soybean Admin. Android is the app's primary target platform.

## Run from source

You need Go, MySQL, Redis, Node.js, pnpm and HBuilderX. See [go.mod](backend/go.mod) and [package.json](admin/package.json) for toolchain requirements. Running the app also requires an Android device or a configured Android runtime.

Run each module in its own terminal. Start the commands below from the repository root.

### Backend

Copy the configuration templates with PowerShell:

```powershell
Copy-Item .env.example .env
Copy-Item backend/config/config.example.yaml backend/config/config.local.yaml
```

Configure the following before starting:

- Set `MYSQL_PASSWORD` in the root `.env`. Set `REDIS_PASSWORD` if Redis authentication is enabled; otherwise leave it empty.
- Set the MySQL and Redis connection details in `backend/config/config.local.yaml`. Both services must be available.
- Use an environment-specific `app.namespace`. The session TTLs can use the template values.
- Set `app.package_dir` to your APK directory, and configure `storage.root_dir` and `storage.base_url` for uploads. Serve uploaded files through a static server such as Nginx, using the matching directory and URL prefix.

Initialize a separate development database using the [database guide](backend/database/README.md). New databases use the current table definitions and required seed data; existing databases use incremental migrations. No default admin account or password is provided.

```powershell
cd backend
go mod download
go run ./cmd/server
```

The template uses port `8080` and the `/api/v1` prefix. Check the service from another terminal:

```powershell
Invoke-RestMethod http://127.0.0.1:8080/api/v1/health
```

App requests also check the client version. Prepare published version data matching `frontend/manifest.json` before app integration; see the [release guide](deploy/README.md) for version data and APK preparation.

### Web admin

```powershell
cd admin
Copy-Item .env.example .env
pnpm install --frozen-lockfile
pnpm dev
```

Set `VITE_SERVICE_BASE_URL` in `admin/.env` to your backend API. The template uses `http://localhost:8080/api/v1`. Open the development URL printed in the terminal.

See the [admin guide](admin/README.md) for environment configuration and additional commands.

### Android app

1. Open `frontend/` in HBuilderX.
2. Check the development API and avatar URLs in `frontend/utils/api.uts`. Running uses the development configuration; release packaging uses the production configuration.
3. Connect and authorize an Android device, then run the app from HBuilderX.

For USB debugging, the development API defaults to `http://127.0.0.1:8080/api/v1`. With ADB available and the device authorized, forward the backend port:

```powershell
adb reverse tcp:8080 tcp:8080
```

The default development avatar URL uses a static server on port `8081`. If you use it, start that server and also run `adb reverse tcp:8081 tcp:8081`. The backend's `storage.base_url` must be reachable from the device. A reachable LAN address is another option; update the development URLs accordingly.

The app uses HBuilderX's uni-app x toolchain and has no npm/pnpm build command.

### Optional development helpers

From the repository root, `make phone` checks the device and local services, then sets up USB forwarding. It requires Make, Windows PowerShell and ADB; for multiple devices, pass `DEVICE_SERIAL=<serial>`.

`make air` starts SSH tunnels to remote MySQL/Redis, then launches Air for hot reload. It requires Make, Air, OpenSSH, a verified host key and the ignored local `.private/dev-ssh.json` configuration. Set local MySQL/Redis ports to `13306`/`16379` respectively. Use it only with the intended remote environment and access scope; local-only development can use `go run ./cmd/server`.

## Checks

In `backend/`:

```powershell
go test ./...
go build ./...
```

In `admin/`:

```powershell
pnpm typecheck
pnpm exec oxlint
pnpm exec eslint .
pnpm exec oxfmt --check src packages build tests
node --test tests/*.test.mjs
pnpm build
```

These lint and formatting commands only check files. `pnpm lint` and `pnpm fmt` modify files. `pnpm build` uses production mode; use `pnpm build:test` for test mode, with the corresponding `.env.prod.example` or `.env.test.example` template.

Database and Redis integration tests require isolated local test instances; see the [backend conventions](backend/AGENTS.md). A skipped integration test is not a passed test. The Node tests cover some app logic, but app compilation and device interaction checks still require HBuilderX and Android.

## Documentation

- [Authentication and Redis sessions](docs/authentication.md)
- [Admin development](admin/README.md)
- [Database initialization and migrations](backend/database/README.md)
- [Deployment and app releases](deploy/README.md)
- [Development conventions](AGENTS.md)

The detailed guides are currently written in Chinese.
