# Greeter palettes in sysc-Go

The accepted greeter design uses sysc-terminal for secondary-output backgrounds.
Its catalog and palette getters come from sysc-Go. Add Ayu, Amber, Blue, Purple,
Green and Orange here, using the sysc-shell dark tokens already bundled with
sysc-greet. Keep existing themes, aliases, defaults and registry order.

Extend the existing switches in animations/palettes.go and append registry
metadata in animations/registry.go. Preserve each getter's slot meanings:
backgrounds first in matrix/fire, smoke and canvas last in burn, and the existing
seven-slot contracts for skull, cracktro and logo. No new API or dependency.

A registry-only change would advertise themes that still render fallback colors.
Copying palettes into sysc-terminal would duplicate the authoritative library.
Adding the colors to the existing getters avoids both problems.

## Implementation and proof

First run animations/greeter_palettes_test.go against the old library: it must
reject the missing catalog entries and fallback colors. Extend the getters and
registry, then check exact tokens, slot counts, case-insensitive lookup and
independent returned slices. Run the capped sysc-Go suite and vet.

Add TestGreeterThemeCatalogAndEffects in sysc-terminal's internal/effect tests.
Run it against the old dependency to prove it rejects all six themes. Build and
run against this library in a temporary Go workspace; exercise all effects.
On the native dual-output desktop, select each theme in fullscreen sysc-greet
--test and check helper IPC, focus, delayed screenshots and exit cleanup.

Publish the reviewed sysc-Go branch only with owner permission. Then pin that
immutable revision in sysc-terminal and repeat the focused checks without a
workspace override. No release tag or installed binary changes belong to this
palette patch. Task status stays in sysc-greet's beads issue 52.
