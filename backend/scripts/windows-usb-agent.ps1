param(
  [Parameter(Mandatory=$true)][string]$BackendBaseUrl,
  [Parameter(Mandatory=$true)][string]$AgentToken,
  [int]$PollSeconds = 5
)

$ErrorActionPreference = "Stop"

function Invoke-JsonPost {
  param(
    [string]$Url,
    [object]$Body,
    [hashtable]$Headers
  )

  $json = $Body | ConvertTo-Json -Depth 8
  return Invoke-RestMethod -Method Post -Uri $Url -Headers $Headers -ContentType "application/json" -Body $json
}

function Set-USBStorageBlock {
  param([bool]$Block)

  $path = "HKLM:\SYSTEM\CurrentControlSet\Services\USBSTOR"
  $value = if ($Block) { 4 } else { 3 }
  Set-ItemProperty -Path $path -Name "Start" -Value $value -Type DWord

  if ($Block) {
    return "USB storage blocked (registry set to 4). Re-plug USB devices to apply immediately."
  }
  return "USB storage unblocked (registry set to 3)."
}

$BackendBaseUrl = $BackendBaseUrl.TrimEnd("/")
$headers = @{ "X-Agent-Token" = $AgentToken }

Write-Host "USB agent started. Polling $BackendBaseUrl every $PollSeconds seconds"

while ($true) {
  try {
    $nextUrl = "$BackendBaseUrl/api/v1/public/agent/commands/next"
    $resp = Invoke-JsonPost -Url $nextUrl -Body @{ agent_token = $AgentToken } -Headers $headers
    $item = $resp.item

    if ($null -ne $item) {
      $success = $true
      $resultText = ""

      try {
        switch ($item.command_type) {
          "usb.block" {
            $resultText = Set-USBStorageBlock -Block $true
          }
          "usb.unblock" {
            $resultText = Set-USBStorageBlock -Block $false
          }
          default {
            throw "Unsupported command type: $($item.command_type)"
          }
        }
      } catch {
        $success = $false
        $resultText = $_.Exception.Message
      }

      $resultUrl = "$BackendBaseUrl/api/v1/public/agent/commands/$($item.id)/result"
      Invoke-JsonPost -Url $resultUrl -Body @{
        agent_token = $AgentToken
        success = $success
        result_text = $resultText
      } -Headers $headers | Out-Null

      if ($success) {
        Write-Host "[$([DateTime]::Now)] Command $($item.id) applied: $resultText"
      } else {
        Write-Warning "[$([DateTime]::Now)] Command $($item.id) failed: $resultText"
      }
    }
  } catch {
    Write-Warning "[$([DateTime]::Now)] Agent polling error: $($_.Exception.Message)"
  }

  Start-Sleep -Seconds $PollSeconds
}
