<#
.SYNOPSIS
Phase 17 Formula and Static Data Governance Validator.
Exit 0 = pass, 1 = any check fails.
#>

param(
    [string]$ProjectRoot = (Get-Location).Path,
    [switch]$Quiet
)

$exitCode = 0
$failedChecks = @()
$totalPassed = 0
$totalFailed = 0
$bnDir = Join-Path $ProjectRoot "backend-next"

function Write-Status($Label, $Passed, $Detail) {
    if ($Passed) { $script:totalPassed++ } else { $script:totalFailed++ }
    $icon = if ($Passed) { "[PASS]" } else { "[FAIL]" }
    $color = if ($Passed) { 'Green' } else { 'Red' }
    if (-not $Quiet) {
        Write-Host ("{0,-8} {1,-55} {2}" -f $icon, $Label, ($Detail -join " ")) -ForegroundColor $color
    }
}

function Write-SubStatus($Label, $Passed, $Detail) {
    $icon = if ($Passed) { "  [PASS]" } else { "  [FAIL]" }
    $color = if ($Passed) { 'Green' } else { 'Red' }
    if (-not $Quiet) {
        Write-Host ("{0,-8} {1,-55} {2}" -f $icon, $Label, ($Detail -join " ")) -ForegroundColor $color
    }
}

function Run-Test($Dir, $Package, $TestName) {
    Push-Location $Dir
    $result = & go test $Package -run $TestName -count=1 2>&1
    $ok = $LASTEXITCODE -eq 0
    Pop-Location
    return $ok
}

function Count-String($Pattern, $Path) {
    $matches = Select-String -Path $Path -Pattern $Pattern -SimpleMatch -ErrorAction SilentlyContinue
    return ($matches | Measure-Object).Count
}

function Test-FileExists($Path) {
    return (Test-Path $Path -PathType Leaf)
}

function Get-CanonicalSHA256($Path) {
    $text = [IO.File]::ReadAllText($Path)
    $canonical = $text.Replace("`r`n", "`n").Replace("`r", "`n")
    $bytes = [Text.UTF8Encoding]::new($false).GetBytes($canonical)
    $sha = [Security.Cryptography.SHA256]::Create()
    try {
        return (($sha.ComputeHash($bytes) | ForEach-Object { $_.ToString('x2') }) -join '')
    } finally {
        $sha.Dispose()
    }
}

Write-Host "`n=== Phase 17 Governance Validator ===" -ForegroundColor Cyan

# ---------------------------------------------------------------
# Check 1: Formula package builds
# ---------------------------------------------------------------
Write-Host "`n--- Check 1: Formula package builds ---" -ForegroundColor Yellow
Push-Location $bnDir
$buildResult = & go build ./internal/formula/ 2>&1
$buildOk = $LASTEXITCODE -eq 0
Pop-Location
Write-Status "FORMULA_BUILD" $buildOk @()
if (-not $buildOk) { $exitCode = 1; $failedChecks += "FORMULA_BUILD_FAILED" }

# ---------------------------------------------------------------
# Check 2: Formula tests pass
# ---------------------------------------------------------------
Write-Host "`n--- Check 2: Formula tests ---" -ForegroundColor Yellow
$formulaOk = Run-Test $bnDir "./internal/formula/" "Test"
Write-Status "FORMULA_TESTS" $formulaOk @()
if (-not $formulaOk) { $exitCode = 1; $failedChecks += "FORMULA_TESTS_FAILED" }

# ---------------------------------------------------------------
# Check 3: Catalog tests pass
# ---------------------------------------------------------------
Write-Host "`n--- Check 3: Catalog tests ---" -ForegroundColor Yellow
$catalogOk = Run-Test $bnDir "./internal/catalog/" "Test"
Write-Status "CATALOG_TESTS" $catalogOk @()
if (-not $catalogOk) { $exitCode = 1; $failedChecks += "CATALOG_TESTS_FAILED" }

# ---------------------------------------------------------------
# Check 4: Service formula-wiring tests pass (3 focused tests)
# ---------------------------------------------------------------
Write-Host "`n--- Check 4: Service formula-wiring tests ---" -ForegroundColor Yellow
$allServiceOk = $true

$prodTestOk = Run-Test $bnDir "./internal/app/production/" "TestStartProduction_DurationMatchesFormula"
Write-SubStatus "PROD_DURATION_TEST" $prodTestOk @()
if (-not $prodTestOk) { $exitCode = 1; $allServiceOk = $false; $failedChecks += "PROD_DURATION_TEST_FAILED" }

$marketTestOk = Run-Test $bnDir "./internal/app/market/" "TestTakeOrder_FeeMatchesFormula"
Write-SubStatus "MARKET_FEE_TEST" $marketTestOk @()
if (-not $marketTestOk) { $exitCode = 1; $allServiceOk = $false; $failedChecks += "MARKET_FEE_TEST_FAILED" }

$financeTestOk = Run-Test $bnDir "./internal/app/finance/" "TestBondInterest_MatchesFormula"
Write-SubStatus "FINANCE_BOND_TEST" $financeTestOk @()
if (-not $financeTestOk) { $exitCode = 1; $allServiceOk = $false; $failedChecks += "FINANCE_BOND_TEST_FAILED" }

Write-Status "SERVICE_TESTS" $allServiceOk @()

# ---------------------------------------------------------------
# Check 5: No deferred-plugin formula names in formula/ directory
# ---------------------------------------------------------------
Write-Host "`n--- Check 5: Deferred-plugin formula names absent ---" -ForegroundColor Yellow
$forbiddenNames = @('AdminOverheadWithCOO','CTOProductionMultiplier','UnitsSoldPerHour','PeriodBondInterest')
$foundForbidden = 0
foreach ($name in $forbiddenNames) {
    $fmtDir = Join-Path $bnDir "internal/formula"
    $matches = Get-ChildItem "$fmtDir/*.go" -Recurse | Select-String -Pattern $name -SimpleMatch -ErrorAction SilentlyContinue
    if ($matches) { $foundForbidden++; Write-SubStatus "FORBIDDEN:$name" $false @() }
}
$noForbiddenOk = ($foundForbidden -eq 0)
Write-Status "NO_DEFERRED_FORMULAS" $noForbiddenOk @("$foundForbidden forbidden names found")
if (-not $noForbiddenOk) { $exitCode = 1; $failedChecks += "FORBIDDEN_FORMULA_FOUND" }

# ---------------------------------------------------------------
# Check 6: Static-data manifest has required paths
# ---------------------------------------------------------------
Write-Host "`n--- Check 6: Static-data manifest required paths ---" -ForegroundColor Yellow
$staticManifestPath = Join-Path $bnDir "internal/catalog/static-data-manifest.json"
$staticPathsOk = $true
$requiredStaticPaths = @(
    "decompiled/data/resources.json",
    "decompiled/data/buildings.json",
    "decompiled/data/economy_model.json",
    "decompiled/data/resource_lookups.json",
    "backend/configs/game.json"
)

if (Test-FileExists $staticManifestPath) {
    $staticManifest = Get-Content $staticManifestPath -Raw | ConvertFrom-Json
    $staticManifestPaths = $staticManifest.files.path
    if ($staticManifestPaths.Count -ne $requiredStaticPaths.Count) {
        Write-SubStatus "STATIC_PATH_COUNT" $false @("got $($staticManifestPaths.Count), want $($requiredStaticPaths.Count)")
        $staticPathsOk = $false
    }
    foreach ($rp in $requiredStaticPaths) {
        if ($staticManifestPaths -notcontains $rp) {
            Write-SubStatus "MISSING:$rp" $false @()
            $staticPathsOk = $false
        }
    }
    if (-not $staticPathsOk) {
        $exitCode = 1
        $failedChecks += "STATIC_MANIFEST_PATHS_INCOMPLETE"
    }
} else {
    $staticPathsOk = $false
    $exitCode = 1
    $failedChecks += "STATIC_MANIFEST_NOT_FOUND"
}
Write-Status "STATIC_MANIFEST_PATHS" $staticPathsOk @()

# ---------------------------------------------------------------
# Check 7: Static-data manifest hashes match
# ---------------------------------------------------------------
Write-Host "`n--- Check 7: Static-data manifest hashes match ---" -ForegroundColor Yellow
$staticHashOk = $true
if (Test-FileExists $staticManifestPath) {
    $staticManifest = Get-Content $staticManifestPath -Raw | ConvertFrom-Json
    foreach ($entry in $staticManifest.files) {
        $fullPath = Join-Path $ProjectRoot $entry.path
        if (Test-FileExists $fullPath) {
            $actualHash = Get-CanonicalSHA256 $fullPath
            $expectedHash = $entry.sha256.ToLower()
            if ($actualHash -ne $expectedHash) {
                Write-SubStatus "HASH_MISMATCH:$($entry.path)" $false @()
                $staticHashOk = $false
            }
        } else {
            Write-SubStatus "FILE_NOT_FOUND:$($entry.path)" $false @()
            $staticHashOk = $false
        }
    }
    if (-not $staticHashOk) {
        $exitCode = 1
        $failedChecks += "STATIC_MANIFEST_HASH_MISMATCH"
    }
} else {
    $staticHashOk = $false
    $exitCode = 1
    $failedChecks += "STATIC_MANIFEST_NOT_FOUND"
}
Write-Status "STATIC_MANIFEST_HASHES" $staticHashOk @()

# ---------------------------------------------------------------
# Check 8: Legacy source manifest has required paths
# ---------------------------------------------------------------
Write-Host "`n--- Check 8: Legacy source manifest required paths ---" -ForegroundColor Yellow
$legacyManifestPath = Join-Path $bnDir "internal/formula/legacy-source-manifest.json"
$legacyPathsOk = $true
$requiredLegacyPaths = @(
    "backend/internal/formula/market.go",
    "backend/internal/formula/bonds.go",
    "backend/internal/formula/production.go",
    "backend/internal/formula/costs.go",
    "backend/internal/formula/saturation.go"
)

if (Test-FileExists $legacyManifestPath) {
    $legacyManifest = Get-Content $legacyManifestPath -Raw | ConvertFrom-Json
    $legacyManifestPaths = $legacyManifest.files.path
    if ($legacyManifestPaths.Count -ne $requiredLegacyPaths.Count) {
        Write-SubStatus "LEGACY_PATH_COUNT" $false @("got $($legacyManifestPaths.Count), want $($requiredLegacyPaths.Count)")
        $legacyPathsOk = $false
    }
    foreach ($rp in $requiredLegacyPaths) {
        if ($legacyManifestPaths -notcontains $rp) {
            Write-SubStatus "MISSING:$rp" $false @()
            $legacyPathsOk = $false
        }
    }
    if (-not $legacyPathsOk) {
        $exitCode = 1
        $failedChecks += "LEGACY_MANIFEST_PATHS_INCOMPLETE"
    }
} else {
    $legacyPathsOk = $false
    $exitCode = 1
    $failedChecks += "LEGACY_MANIFEST_NOT_FOUND"
}
Write-Status "LEGACY_MANIFEST_PATHS" $legacyPathsOk @()

# ---------------------------------------------------------------
# Check 9: Legacy source manifest hashes match
# ---------------------------------------------------------------
Write-Host "`n--- Check 9: Legacy source manifest hashes match ---" -ForegroundColor Yellow
$legacyHashOk = $true
if (Test-FileExists $legacyManifestPath) {
    $legacyManifest = Get-Content $legacyManifestPath -Raw | ConvertFrom-Json
    foreach ($entry in $legacyManifest.files) {
        $fullPath = Join-Path $ProjectRoot $entry.path
        if (Test-FileExists $fullPath) {
            $actualHash = Get-CanonicalSHA256 $fullPath
            $expectedHash = $entry.sha256.ToLower()
            if ($actualHash -ne $expectedHash) {
                Write-SubStatus "HASH_MISMATCH:$($entry.path)" $false @()
                $legacyHashOk = $false
            }
        } else {
            Write-SubStatus "FILE_NOT_FOUND:$($entry.path)" $false @()
            $legacyHashOk = $false
        }
    }
    if (-not $legacyHashOk) {
        $exitCode = 1
        $failedChecks += "LEGACY_MANIFEST_HASH_MISMATCH"
    }
} else {
    $legacyHashOk = $false
    $exitCode = 1
    $failedChecks += "LEGACY_MANIFEST_NOT_FOUND"
}
Write-Status "LEGACY_MANIFEST_HASHES" $legacyHashOk @()

# ---------------------------------------------------------------
# Check 10: Governed formula calls present in service files
# ---------------------------------------------------------------
Write-Host "`n--- Check 10: Governed formula calls present ---" -ForegroundColor Yellow
$governedOk = $true

$appDir = Join-Path $bnDir "internal/app"
$prodServicePath = Join-Path (Join-Path $appDir "production") "service.go"
$marketServicePath = Join-Path (Join-Path $appDir "market") "service.go"
$financeServicePath = Join-Path (Join-Path $appDir "finance") "service.go"

$pdCount = Count-String "formula.DurationSeconds(" $prodServicePath
$pdOk = ($pdCount -ge 1)
Write-SubStatus "PROD_DURATION_FORMULA (count=$pdCount, need>=1)" $pdOk @()
if (-not $pdOk) { $exitCode = 1; $governedOk = $false; $failedChecks += "PROD_DURATION_NOT_GOVERNED" }

$mfCount = Count-String "formula.ExchangeFee(" $marketServicePath
$mfOk = ($mfCount -ge 2)
Write-SubStatus "MARKET_FEE_FORMULA (count=$mfCount, need>=2)" $mfOk @()
if (-not $mfOk) { $exitCode = 1; $governedOk = $false; $failedChecks += "MARKET_EXCHANGE_FEE_NOT_GOVERNED" }

$fbCount = Count-String "formula.DailyBondInterest(" $financeServicePath
$fbOk = ($fbCount -ge 1)
Write-SubStatus "FINANCE_BOND_FORMULA (count=$fbCount, need>=1)" $fbOk @()
if (-not $fbOk) { $exitCode = 1; $governedOk = $false; $failedChecks += "FINANCE_BOND_INTEREST_NOT_GOVERNED" }

Write-Status "GOVERNED_CALLS" $governedOk @()

# ---------------------------------------------------------------
# Check 11: No old private wrapper in finance/service.go
# ---------------------------------------------------------------
Write-Host "`n--- Check 11: No old private wrapper ---" -ForegroundColor Yellow
$wrapperCount = Count-String "func dailyBondInterest" $financeServicePath
$wrapperOk = ($wrapperCount -eq 0)
Write-Status "NO_OLD_WRAPPER" $wrapperOk @("$wrapperCount occurrences of 'func dailyBondInterest'")
if (-not $wrapperOk) { $exitCode = 1; $failedChecks += "OLD_DAILY_BOND_INTEREST_WRAPPER" }

# ---------------------------------------------------------------
# Check 12: No old critical inline expressions
# ---------------------------------------------------------------
Write-Host "`n--- Check 12: No old inline expressions ---" -ForegroundColor Yellow
$inlinePatterns = @(
    @{ Path = $marketServicePath; Pattern = "math\.Ceil\(float64\(fill\).*\*.*exchangeFeePct" },
    @{ Path = $financeServicePath; Pattern = "math\.Floor\(float64\(amount\).*\*.*faceValue" },
    @{ Path = $prodServicePath; Pattern = "math\.Ceil\(float64\(req\.Quantity\).*\* 3600" }
)
$inlineCount = 0
foreach ($entry in $inlinePatterns) {
    $matches = Select-String -Path $entry.Path -Pattern $entry.Pattern -ErrorAction SilentlyContinue
    $inlineCount += ($matches | Measure-Object).Count
}
$inlineOk = ($inlineCount -eq 0)
Write-Status "NO_INLINE_CRITICAL_FORMULAS" $inlineOk @("$inlineCount old critical expressions found")
if (-not $inlineOk) { $exitCode = 1; $failedChecks += "OLD_INLINE_CRITICAL_FORMULA" }

# ---------------------------------------------------------------
# Summary
# ---------------------------------------------------------------
$totalChecks = $totalPassed + $totalFailed
Write-Host "`n=== Summary ===" -ForegroundColor Cyan
Write-Host "Total checks: $totalChecks" -ForegroundColor Cyan
Write-Host "Passed: $totalPassed" -ForegroundColor Green
if ($totalFailed -gt 0) {
    Write-Host "Failed: $totalFailed" -ForegroundColor Red
}

Write-Host "`n=== $(if ($exitCode -eq 0) { 'ALL CHECKS PASSED - Governance established' } else { 'CHECKS FAILED - Governance drift detected' }) ===" -ForegroundColor $(if ($exitCode -eq 0) { 'Green' } else { 'Red' })
if ($failedChecks.Count -gt 0) {
    Write-Host "`nFailed checks:" -ForegroundColor Red
    foreach ($e in ($failedChecks | Select-Object -Unique)) {
        Write-Host "  $e" -ForegroundColor Red
    }
}
exit $exitCode
