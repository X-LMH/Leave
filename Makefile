.PHONY: migrate phone

migrate:
	cd backend && go run ./cmd/migrate -config ./config/config.local.yaml

# Optional: make connect-phone DEVICE_SERIAL=...
phone:
	powershell.exe -NoProfile -ExecutionPolicy Bypass -File deploy/connect-phone.ps1 $(if $(DEVICE_SERIAL),-DeviceSerial "$(DEVICE_SERIAL)",)
