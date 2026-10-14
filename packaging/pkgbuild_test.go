package packaging_test

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

func TestPKGBUILDPinsFetchedSource(t *testing.T) {
	pkgbuild := readRepoFile(t, "PKGBUILD")
	srcinfo := readRepoFile(t, ".SRCINFO")

	ver := mustMatch(t, pkgbuild, `(?m)^pkgver=([0-9.]+)$`)
	sum := mustMatch(t, pkgbuild, `(?m)^sha256sums=\('([0-9a-f]{64})'\)$`)
	goVersion := mustMatch(t, readRepoFile(t, "go.mod"), `(?m)^go ([0-9.]+)$`)
	goDependency := mustMatch(t, pkgbuild, `(?m)^makedepends=\('go>=([0-9.]+)'\)$`)
	if goDependency != goVersion {
		t.Fatalf("PKGBUILD Go dependency = %s, want go.mod requirement %s", goDependency, goVersion)
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
