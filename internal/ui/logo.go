package ui

import (
	"fmt"
	"os"
)

const (
	ansiTeal    = "\033[36m"
	ansiTealDim = "\033[2;36m"
	ansiBold    = "\033[1m"
	ansiDim     = "\033[2m"
	ansiReset   = "\033[0m"
	ansiWhite   = "\033[97m"
)

// PrintLogo prints the Faro Helm ASCII logo to stdout.
// If NO_COLOR is set, a plain-text fallback is printed instead.
//
// Mark geometry: dot + two tapering bars (ratio 14 : 30 : 19).
// Each bar rendered as two rows of ━ for visual thickness.
// bar1 = 11 chars × 2 rows; bar2 = 7 chars × 2 rows (~64% of bar1).
// The dot sits at the junction between bar1's rows (matching cy=22 in the SVG).
//
// Output:
//
//	     ━━━━━━━━━━━
//	  ●  ━━━━━━━━━━━  faro helm
//	     ━━━━━━━
//	     ━━━━━━━      team standups · attendance · leave
func PrintLogo() {
	fmt.Println()

	if os.Getenv("NO_COLOR") != "" {
		fmt.Println("  Faro Helm")
		fmt.Println("  team standups · attendance · leave")
		fmt.Println()
		return
	}

	// Mark elements — bar1 = 11 chars, bar2 = 7 chars (ratio 11:7 ≈ 1.57, spec is 30:19 ≈ 1.58)
	dot      := ansiTeal + "●" + ansiReset
	bar1row  := ansiTeal + "━━━━━━━━━━━" + ansiReset       // 11 chars, used for both rows of bar1
	bar2row  := ansiTealDim + "━━━━━━━" + ansiReset         // 7 chars, used for both rows of bar2

	// Wordmark — "faro" bold white, "helm" teal
	faro := ansiBold + ansiWhite + "faro" + ansiReset
	helm := ansiTeal + "helm" + ansiReset

	// Tagline — dim, aligned to same column as wordmark
	// col of wordmark = 5(indent+dot+space) + 11(bar1) + 2(space) = 18
	// col of tagline  = 5(indent) + 7(bar2) + 6(pad) = 18  ✓
	tag := ansiDim + "team standups · attendance · leave" + ansiReset

	//  Layout (dot at junction of bar1's two rows):
	//    row 0:       ━━━━━━━━━━━
	//    row 1:  ●    ━━━━━━━━━━━   faro helm
	//    row 2:       ━━━━━━━
	//    row 3:       ━━━━━━━       team standups · attendance · leave
	fmt.Printf("     %s\n", bar1row)
	fmt.Printf("  %s  %s  %s %s\n", dot, bar1row, faro, helm)
	fmt.Printf("     %s\n", bar2row)
	fmt.Printf("     %s      %s\n", bar2row, tag)

	fmt.Println()
}
