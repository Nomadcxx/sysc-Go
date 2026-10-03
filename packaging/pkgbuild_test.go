package packaging_test

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// v1.0.3Sum is the SHA-256 of the GitHub tag archive that PKGBUILD fetches:
// https://github.com/Nomadcxx/sysc-Go/archive/v1.0.3.tar.gz
// Computed 2026-10-02. v1.0.2 is 664c308fcc910ab7f573ef577e00f2ebd2cae92ae5da80a33bdd20481d9f2af4.
const (
	wantPkgVer = "1.0.3"
	v103Sum    = "d30e1246b157970be690f45403b978ae4f2f1c16044a66e864eb6e4b2e560bdd"
)

func TestPKGBUILDPinsFetchedSource(t *testing.T) {
	pkgbuild := readRepoFile(t, "PKGBUILD")
	srcinfo := readRepoFile(t, ".SRCINFO")

	ver := mustMatch(t, pkgbuild, `(?m)^pkgver=([0-9.]+)$`)
	sum := mustMatch(t, pkgbuild, `(?m)^sha256sums=\('([0-9a-f]{64})'\)$`)
	if ver != wantPkgVer {
		t.Fatalf("PKGBUILD pkgver = %s, want %s", ver, wantPkgVer)
	}
	if sum != v103Sum {
		t.Fatalf("PKGBUILD sha256sums = %s, want v%s sum %s", sum, wantPkgVer, v103Sum)
	}
	if !strings.Contains(pkgbuild, "archive/v${pkgver}.tar.gz") {
		t.Fatal("PKGBUILD source URL does not follow v${pkgver}")
	}

	srcVer := mustMatch(t, srcinfo, `(?m)^\tpkgver = ([0-9.]+)$`)
	srcSum := mustMatch(t, srcinfo, `(?m)^\tsha256sums = ([0-9a-f]{64})$`)
	source := mustMatch(t, srcinfo, `(?m)^\tsource = (.+)$`)
	if srcVer != ver || srcSum != sum {
		t.Fatalf(".SRCINFO pkgver/sha256sums = %s / %s, PKGBUILD has %s / %s", srcVer, srcSum, ver, sum)
	}
	wantSource := "syscgo-" + ver + ".tar.gz::https://github.com/Nomadcxx/sysc-Go/archive/v" + ver + ".tar.gz"
	if source != wantSource {
		t.Fatalf(".SRCINFO source = %s, want %s", source, wantSource)
	}
}

func readRepoFile(t *testing.T, name string) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("..", name))
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func mustMatch(t *testing.T, text, pattern string) string {
	t.Helper()
	m := regexp.MustCompile(pattern).FindStringSubmatch(text)
	if m == nil {
		t.Fatalf("pattern %s did not match", pattern)
	}
	return m[1]
}
