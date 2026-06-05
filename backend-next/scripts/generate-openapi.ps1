$root = Split-Path -Parent $PSScriptRoot
$oapi = "$(go env GOPATH)\bin\oapi-codegen.exe"
if (-not (Test-Path $oapi)) {
    $oapi = "oapi-codegen"
    if (-not (Get-Command $oapi -ErrorAction SilentlyContinue)) {
        Write-Error "oapi-codegen not found in PATH or GOPATH/bin"
        exit 1
    }
}

$null = New-Item -ItemType Directory -Force -Path "$root\internal\generated\openapi"

& $oapi -generate types,chi-server -package openapi `
    "$root\openapi\openapi-draft.yaml" `
    | Out-File -FilePath "$root\internal\generated\openapi\types.gen.go" -Encoding utf8

Write-Host "Codegen complete: internal/generated/openapi/types.gen.go"
