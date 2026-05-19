# Windows USB Agent Setup (Free Testing)

## 1) Get Agent Token from Admin UI
- Login as admin.
- Open Employee Laptops table.
- Click Copy Token for the laptop you want to control.

## 2) Run Agent as Administrator (PowerShell)
```powershell
Set-ExecutionPolicy -Scope Process Bypass -Force
cd C:\path\to\firewall\backend\scripts
.\windows-usb-agent.ps1 -BackendBaseUrl "https://your-app.onrender.com" -AgentToken "PASTE_TOKEN_HERE" -PollSeconds 5
```

## 3) Queue USB Commands
- In admin dashboard laptop table:
  - Block USB queues `usb.block`
  - Unblock USB queues `usb.unblock`

## 4) Notes
- USB storage block/unblock uses `HKLM\SYSTEM\CurrentControlSet\Services\USBSTOR`.
- Value `4` blocks USB storage, value `3` unblocks.
- Re-plug USB device after command for immediate effect.
- Script must run as Administrator.
