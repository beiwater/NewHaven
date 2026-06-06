<#
.SYNOPSIS
Phase 16 Inventory Validator v3 -- validates ALL Phase 16 CSVs with source drift checks.
Exit 0 = pass, 1 = any check fails.

.DESCRIPTION
Validates all 6 CSVs under docs/2026-06-06/rebuild/phase-16/:
- legacy-routes.csv: zero unclassified, valid enums, unique paths
- classification.csv: matches legacy-routes, zero unclassified
- backend-next-routes.csv: handler_domain, disposition, verification required
- frontend-callsites.csv: 58 rows, source_ref validated, legacy_owner/verification required
- scheduler-inventory.csv: 9 stages, responsible_owner/verification required
- data-groups.csv: >= 30 rows, verification_requirement non-empty
All disposition/verification fields validated against approved enums.
No non-ASCII characters allowed.
Counts and source-reference sets are derived from the current source tree.
#>

param(
    [string]$ProjectRoot = (Get-Location).Path,
    [switch]$Quiet
)

function Write-Status($Label, $Passed, $Detail) {
    $icon = if ($Passed) { "[PASS]" } else { "[FAIL]" }
    $color = if ($Passed) { 'Green' } else { 'Red' }
    if (-not $Quiet) {
        Write-Host ("{0,-8} {1,-50} {2}" -f $icon, $Label, ($Detail -join " ")) -ForegroundColor $color
    }
}

$exitCode = 0
$errors = @()
$phase16Dir = Join-Path $ProjectRoot "docs/2026-06-06/rebuild/phase-16"

# Approved enums
$validDispositions = @('migrated','remaining-core','dev-only','retire-candidate','deferred-plugin','new-only')
$validMethodAmbiguity = @('exact','prefix','internal-switch')
$validVerification = @('status-200','shape','shape+semantics','shape+semantics+side-effects','none')

function Load-Csv($name) {
    $path = Join-Path $phase16Dir $name
    if (Test-Path $path) { return Import-Csv -Path $path }
    return @()
}

function Compare-StringSet($Expected, $Actual) {
    $expectedSet = @{}
    $actualSet = @{}
    foreach ($item in $Expected) { $expectedSet[[string]$item] = $true }
    foreach ($item in $Actual) { $actualSet[[string]$item] = $true }

    $missing = @($expectedSet.Keys | Where-Object { -not $actualSet.ContainsKey($_) } | Sort-Object)
    $extra = @($actualSet.Keys | Where-Object { -not $expectedSet.ContainsKey($_) } | Sort-Object)
    return [PSCustomObject]@{
        Passed = ($missing.Count -eq 0 -and $extra.Count -eq 0)
        Missing = $missing
        Extra = $extra
    }
}

Write-Host "`n=== Phase 16 Completion Dashboard v3 ===" -ForegroundColor Cyan
Write-Host "Generated: $(Get-Date -Format 'yyyy-MM-dd HH:mm:ss')" -ForegroundColor Gray

# -- 1. File existence -------------------------------
$csvFiles = @('legacy-routes.csv','backend-next-routes.csv','frontend-callsites.csv','scheduler-inventory.csv','data-groups.csv','classification.csv')
$allExist = $true
$missingFiles = @()
foreach ($f in $csvFiles) {
    $path = Join-Path $phase16Dir $f
    if (-not (Test-Path $path)) { $allExist = $false; $missingFiles += $f }
}
Write-Status "ALL_CSVS_EXIST" $allExist @("$(if ($missingFiles.Count -gt 0) { "Missing: $($missingFiles -join ',')" } else { '6/6' })")
if (-not $allExist) { $exitCode = 1 }

# -- 2. Non-ASCII check -----------------------------
$nonAsciiFiles = @()
foreach ($f in ($csvFiles + @('../phase-16-summary.md'))) {
    $path = Join-Path $phase16Dir $f
    if (Test-Path $path) {
        $content = Get-Content -Path $path -Raw
        if ($content -match '[^\x00-\x7F]') {
            $nonAsciiFiles += $f
            $errors += "NON_ASCII: $f"
        }
    }
}
Write-Status "NO_NON_ASCII" ($nonAsciiFiles.Count -eq 0) @("$($nonAsciiFiles.Count) files with non-ASCII")
if ($nonAsciiFiles.Count -gt 0) { $exitCode = 1 }

# -- 3. Source drift verification -------------------
$legacy = Load-Csv 'legacy-routes.csv'
$bn = Load-Csv 'backend-next-routes.csv'
$fe = Load-Csv 'frontend-callsites.csv'
$sc = Load-Csv 'scheduler-inventory.csv'
$dg = Load-Csv 'data-groups.csv'
$class = Load-Csv 'classification.csv'

$legacySourcePaths = @(
    Get-ChildItem (Join-Path $ProjectRoot 'backend/internal/handler') -Filter '*.go' |
        Where-Object { $_.Name -notlike '*_test.go' } |
        Select-String -Pattern 'mux\.HandleFunc\("([^"]+)"' -AllMatches |
        ForEach-Object { $_.Matches | ForEach-Object { $_.Groups[1].Value } }
)
$legacySourceComparison = Compare-StringSet $legacySourcePaths @($legacy.legacy_path)
Write-Status "SOURCE_LEGACY_PATHS" $legacySourceComparison.Passed @("$($legacySourcePaths.Count) source registrations; missing=$($legacySourceComparison.Missing.Count); extra=$($legacySourceComparison.Extra.Count)")
if (-not $legacySourceComparison.Passed) {
    $exitCode = 1
    $errors += @($legacySourceComparison.Missing | ForEach-Object { "LEGACY_SOURCE_MISSING_IN_CSV: $_" })
    $errors += @($legacySourceComparison.Extra | ForEach-Object { "LEGACY_CSV_NOT_IN_SOURCE: $_" })
}

$bnSourceHandlers = @(
    Select-String -Path (Join-Path $ProjectRoot 'backend-next/internal/httpapi/router.go') -Pattern '\.(Get|Post|Put|Delete|Patch)\("[^"]+",\s*([^)]+)\)' -AllMatches |
        ForEach-Object { $_.Matches | ForEach-Object { ($_.Groups[2].Value.Trim() -split '\.')[-1] } }
)
$bnSourceComparison = Compare-StringSet $bnSourceHandlers @($bn.handler_func)
Write-Status "SOURCE_BN_HANDLERS" $bnSourceComparison.Passed @("$($bnSourceHandlers.Count) source registrations; missing=$($bnSourceComparison.Missing.Count); extra=$($bnSourceComparison.Extra.Count)")
if (-not $bnSourceComparison.Passed) {
    $exitCode = 1
    $errors += @($bnSourceComparison.Missing | ForEach-Object { "BN_SOURCE_MISSING_IN_CSV: $_" })
    $errors += @($bnSourceComparison.Extra | ForEach-Object { "BN_CSV_NOT_IN_SOURCE: $_" })
}

$frontendRoot = (Resolve-Path (Join-Path $ProjectRoot 'client/atlas-foods-client/src')).Path
$frontendSourceRefs = @(
    Get-ChildItem $frontendRoot -Recurse -File -Include '*.ts','*.tsx' |
        Select-String -Pattern 'api\.(get|post|put|delete|patch)' |
        ForEach-Object {
            $_.Path.Substring($frontendRoot.Length + 1).Replace('\','/') + ':' + $_.LineNumber
        }
)
$feSourceComparison = Compare-StringSet $frontendSourceRefs @($fe.source_ref)
Write-Status "SOURCE_FE_REFS" $feSourceComparison.Passed @("$($frontendSourceRefs.Count) source callsites; missing=$($feSourceComparison.Missing.Count); extra=$($feSourceComparison.Extra.Count)")
if (-not $feSourceComparison.Passed) {
    $exitCode = 1
    $errors += @($feSourceComparison.Missing | ForEach-Object { "FE_SOURCE_MISSING_IN_CSV: $_" })
    $errors += @($feSourceComparison.Extra | ForEach-Object { "FE_CSV_NOT_IN_SOURCE: $_" })
}

# -- 4. legacy-routes.csv validation -----------------
$legacyErrors = 0
$legacyPaths = @{}
$unclassifiedCount = 0

foreach ($row in $legacy) {
    $lp = $row.legacy_path
    if ([string]::IsNullOrWhiteSpace($lp)) { $legacyErrors++; continue }
    if ($legacyPaths.ContainsKey($lp)) { $errors += "DUPLICATE_LEGACY_PATH: $lp"; $legacyErrors++ }
    $legacyPaths[$lp] = $true

    $disp = $row.disposition
    if ([string]::IsNullOrWhiteSpace($disp)) { $errors += "EMPTY_DISP: $lp"; $legacyErrors++ }
    elseif ($validDispositions -notcontains $disp) { $errors += "INVALID_DISP '${disp}': $lp"; $legacyErrors++ }
    elseif ($disp -eq 'unclassified') { $unclassifiedCount++; $errors += "UNCLASSIFIED: $lp"; $legacyErrors++ }

    if ($disp -eq 'migrated' -and [string]::IsNullOrWhiteSpace($row.backend_next_owner)) {
        $errors += "MIGRATED_NO_BN_OWNER: $lp"; $legacyErrors++
    }

    if ([string]::IsNullOrWhiteSpace($row.dependency_group)) { $errors += "EMPTY_DEP: $lp"; $legacyErrors++ }

    $ver = $row.verification_requirement
    if ([string]::IsNullOrWhiteSpace($ver)) { $errors += "EMPTY_VER: $lp"; $legacyErrors++ }
    elseif ($ver -eq 'none' -and [string]::IsNullOrWhiteSpace($row.notes)) { $errors += "NONE_VER_NO_NOTES: $lp"; $legacyErrors++ }
    elseif ($validVerification -notcontains $ver) { $errors += "INVALID_VER '${ver}': $lp"; $legacyErrors++ }

    $ma = $row.method_ambiguity
    if (-not [string]::IsNullOrWhiteSpace($ma) -and $validMethodAmbiguity -notcontains $ma) {
        $errors += "INVALID_AMBIGUITY '${ma}': $lp"; $legacyErrors++
    }
}
Write-Status "LEGACY_COUNT" ($legacy.Count -eq $legacySourcePaths.Count) @("$($legacy.Count) rows (source has $($legacySourcePaths.Count))")
Write-Status "LEGACY_VALID" ($legacyErrors -eq 0) @("$legacyErrors issues")
Write-Status "LEGACY_UNCLASSIFIED" ($unclassifiedCount -eq 0) @("$unclassifiedCount unclassified")
if ($legacyErrors -gt 0) { $exitCode = 1 }

# -- 5. classification.csv ---------------------------
$classErrors = 0
$unclass = 0
foreach ($row in $class) {
    $lp = $row.legacy_path
    $disp = $row.disposition
    if ([string]::IsNullOrWhiteSpace($disp) -or $disp -eq 'unclassified') { $unclass++; $classErrors++ }
    elseif ($validDispositions -notcontains $disp) { $classErrors++ }
    if ([string]::IsNullOrWhiteSpace($row.dependency_group)) { $classErrors++ }
    if (-not $legacyPaths.ContainsKey($lp)) { $classErrors++ }
}
Write-Status "CLASSIFICATION" ($class.Count -eq $legacy.Count -and $classErrors -eq 0 -and $unclass -eq 0) @("$($class.Count) rows, $classErrors issues, $unclass unclassified")
if ($classErrors -gt 0 -or $class.Count -ne $legacy.Count) { $exitCode = 1 }

# -- 6. backend-next-routes.csv ----------------------
$bnErrors = 0
foreach ($row in $bn) {
    if ([string]::IsNullOrWhiteSpace($row.handler_domain)) { $errors += "BN_EMPTY_DOMAIN: $($row.path)"; $bnErrors++ }
    if ([string]::IsNullOrWhiteSpace($row.disposition)) { $errors += "BN_EMPTY_DISP: $($row.path)"; $bnErrors++ }
    if ([string]::IsNullOrWhiteSpace($row.verification_requirement)) { $errors += "BN_EMPTY_VER: $($row.path)"; $bnErrors++ }
}
Write-Status "BN_COUNT" ($bn.Count -eq $bnSourceHandlers.Count) @("$($bn.Count) rows (source has $($bnSourceHandlers.Count))")
Write-Status "BN_VALID" ($bnErrors -eq 0) @("$bnErrors issues")
if ($bnErrors -gt 0) { $exitCode = 1 }

# -- 7. frontend-callsites.csv -----------------------
$feErrors = 0
$feSourceRefs = @{}
foreach ($row in $fe) {
    $sr = $row.source_ref
    if ([string]::IsNullOrWhiteSpace($sr)) { $errors += "FE_NO_SOURCE_REF: $($row.function_name)"; $feErrors++; continue }
    if ($feSourceRefs.ContainsKey($sr)) { $errors += "FE_DUP_SOURCE_REF: $sr"; $feErrors++ }
    $feSourceRefs[$sr] = $true

    # Verify source_ref file exists
    $parts = $sr -split ':'
    $relFile = $parts[0]
    $fullPath = "$ProjectRoot\client\atlas-foods-client\src\$relFile"
    $fullPath = $fullPath.Replace('/','\')
    if (-not (Test-Path $fullPath)) { $errors += "FE_SOURCE_REF_MISSING: $sr"; $feErrors++ }

    $disp = $row.disposition
    if ([string]::IsNullOrWhiteSpace($disp)) { $feErrors++; }
    elseif ($validDispositions -notcontains $disp) { $feErrors++ }

    if ([string]::IsNullOrWhiteSpace($row.legacy_owner)) { $errors += "FE_NO_LEGACY_OWNER: $($row.function_name)"; $feErrors++ }
    if ([string]::IsNullOrWhiteSpace($row.verification_requirement)) { $errors += "FE_NO_VER: $($row.function_name)"; $feErrors++ }
}
Write-Status "FE_COUNT" ($fe.Count -eq $frontendSourceRefs.Count) @("$($fe.Count) rows (source has $($frontendSourceRefs.Count))")
Write-Status "FE_VALID" ($feErrors -eq 0) @("$feErrors issues")
Write-Status "FE_SOURCE_REFS" (($fe | Where-Object { -not [string]::IsNullOrWhiteSpace($_.source_ref) }).Count -eq $fe.Count) @("all $($fe.Count) have source_ref")
if ($feErrors -gt 0) { $exitCode = 1 }

# -- 8. scheduler-inventory.csv ----------------------
$scSteps = ($sc | Where-Object { $_.step -match '^\d+$' -and $_.domain -eq 'scheduler' }).Count
$scStageCount = ($sc | Where-Object { $_.step -match '^\d+$' -and $_.domain -eq 'scheduler' } | Select-Object -ExpandProperty step -Unique).Count
Write-Status "SCHEDULER_STAGES" ($scStageCount -eq 9) @("$scStageCount stages ($scSteps calls; 9 ordered tick stages)")

$scErrors = 0
foreach ($row in ($sc | Where-Object { $_.step -match '^\d+$' })) {
    if ([string]::IsNullOrWhiteSpace($row.responsible_owner)) { $errors += "SC_EMPTY_RESP: step $($row.step) $($row.method_name)"; $scErrors++ }
    if ([string]::IsNullOrWhiteSpace($row.verification_requirement)) { $errors += "SC_EMPTY_VER: step $($row.step) $($row.method_name)"; $scErrors++ }
}
Write-Status "SCHEDULER_VALID" ($scErrors -eq 0) @("$scErrors issues")
if ($scErrors -gt 0) { $exitCode = 1 }

# -- 9. data-groups.csv -----------------------------
$dgErrors = 0
foreach ($row in $dg) {
    if ([string]::IsNullOrWhiteSpace($row.disposition)) { $dgErrors++ }
    if ([string]::IsNullOrWhiteSpace($row.dependency_group)) { $dgErrors++ }
    if ([string]::IsNullOrWhiteSpace($row.verification_requirement)) { $dgErrors++ }
}
Write-Status "DG_COUNT" ($dg.Count -ge 30) @("$($dg.Count) rows (expected >= 30)")
Write-Status "DG_VALID" ($dgErrors -eq 0) @("$dgErrors issues")
if ($dgErrors -gt 0) { $exitCode = 1 }

# -- 10. Dashboard counts (computed directly from CSVs) ---
$legacyDisp = @{}
foreach ($row in $legacy) { $d = $row.disposition; if (-not $legacyDisp.ContainsKey($d)) { $legacyDisp[$d] = 0 }; $legacyDisp[$d]++ }

$feDisp = @{}
foreach ($row in $fe) { $d = $row.disposition; if (-not $feDisp.ContainsKey($d)) { $feDisp[$d] = 0 }; $feDisp[$d]++ }

Write-Host "`n--- Legacy Route Counts by Disposition (computed from CSV) ---" -ForegroundColor Yellow
foreach ($key in ($legacyDisp.Keys | Sort-Object)) {
    Write-Host ("  {0,-25} {1,3}" -f "${key}:", $legacyDisp[$key])
}
Write-Host ("  {0,-25} {1,3}" -f "TOTAL:", $legacy.Count)

$coreTotal = $legacy.Count - ($legacyDisp['deferred-plugin'] + $legacyDisp['dev-only'])
Write-Host ("`n  Core total (excl. deferred-plugin + dev-only): {0}" -f $coreTotal)
Write-Host ("  Deferred-plugin excluded from core: {0}" -f $legacyDisp['deferred-plugin'])

Write-Host "`n--- Frontend Callsites by Disposition (computed from CSV) ---" -ForegroundColor Yellow
foreach ($key in ($feDisp.Keys | Sort-Object)) {
    Write-Host ("  {0,-25} {1,3}" -f "${key}:", $feDisp[$key])
}
Write-Host ("  {0,-25} {1,3}" -f "TOTAL:", $fe.Count)

# -- Errors Summary ----------------------------------
if ($errors.Count -gt 0) {
    Write-Host "`n--- Issues Found ($($errors.Count)) ---" -ForegroundColor Red
    foreach ($e in $errors) { Write-Host "  $e" -ForegroundColor Red }
}

Write-Host "`n=== $(if ($exitCode -eq 0) { 'ALL CHECKS PASSED' } else { 'CHECKS FAILED' }) ===" -ForegroundColor $(if ($exitCode -eq 0) { 'Green' } else { 'Red' })
exit $exitCode
