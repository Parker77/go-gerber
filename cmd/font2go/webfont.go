package main

import (
	"log"
	"regexp"
	"strconv"
)

// FontData represents the SVG webfont data.
type FontData struct {
	Font *Font `xml:"defs>font"`
}

// Font represents the <font> XML block of the webfont data.
type Font struct {
	ID           string        `xml:"id,attr"`
	HorizAdvX    int           `xml:"horiz-adv-x,attr"`
	FontFace     *FontFace     `xml:"font-face"`
	MissingGlyph *MissingGlyph `xml:"missing-glyph"`
	Glyphs       []*Glyph      `xml:"glyph"`
}

// FontFace represents the <font-face> XML block of the webfont data.
type FontFace struct {
	UnitsPerEm int `xml:"units-per-em,attr"`
	Ascent     int `xml:"ascent,attr"`
	Descent    int `xml:"descent,attr"`
}

// MissingGlyph represents the <missing-glyph> XML block of the webfont data.
type MissingGlyph struct {
	HorizAdvX int `xml:"horiz-adv-x,attr"`
}

// Glyph represents a <glyph> XML block of the webfont data.
type Glyph struct {
	HorizAdvX int     `xml:"horiz-adv-x,attr"`
	Unicode   *string `xml:"unicode,attr,omitempty"`
	D         *string `xml:"d,attr,omitempty"`
	DOrig     *string `xml:"d-orig,attr,omitempty"`
	GerberLP  *string `xml:"gerber-lp,attr,omitempty"`

	// D is parsed into a sequence of PathSteps:
	PathSteps []*PathStep
}

// PathStep represents a single path step.
//
// There are 20 possible commands, broken up into 6 types,
// with each command having an "absolute" (upper case) and
// a "relative" (lower case) version.
//
// MoveTo: M, m
// LineTo: L, l, H, h, V, v
// Cubic Bézier Curve: C, c, S, s
// Quadratic Bézier Curve: Q, q, T, t
// Elliptical Arc Curve: A, a
// ClosePath: Z, z
type PathStep struct {
	Command    string
	Parameters []float64
}

var (
	cmdRE   = regexp.MustCompile(`(?i)^([mlhvcsqta])(?:\s*(-?\d+\.?\d*)[,\s+]?)+`)
	closeRE = regexp.MustCompile(`(?i)^(z)\s*`)
	numRE   = regexp.MustCompile(`^\s*(-?\d+\.?\d*)[,\s+]?`)
)

// ParsePath parses a Glyph path.
func (g *Glyph) ParsePath() {
	if g == nil || g.D == nil {
		return
	}
	d := *g.D
	if g.DOrig != nil && *g.DOrig != "" {
		// log.Printf("Warning: ignoring DOrig for glyph %+q", *g.Unicode)
		log.Printf("Warning: using DOrig for glyph %+q", *g.Unicode)
		d = *g.DOrig
	}

	var numZs int
	for len(d) > 0 {
		m := closeRE.FindStringSubmatch(d)
		if len(m) == 2 {
			g.PathSteps = append(g.PathSteps, &PathStep{Command: m[1]})
			d = d[len(m[0]):]
			numZs++
			continue
		}

		m = cmdRE.FindStringSubmatch(d)
		if len(m) >= 3 {
			g.PathSteps = append(g.PathSteps, &PathStep{
				Command:    m[1],
				Parameters: parseParams(m[0][1:]),
			})
			d = d[len(m[0]):]
			continue
		}

		log.Fatalf("Unknown path command: %q", d)
	}

	if numZs > 1 && (g.GerberLP == nil || len(*g.GerberLP) != numZs) {
		if g.GerberLP == nil {
			log.Printf("Warning: glyph=%+q, numZs=%v, g.GerberLP=<nil>", *g.Unicode, numZs)
		} else {
			log.Printf("Warning: glyph=%+q, numZs=%v, g.GerberLP=%q", *g.Unicode, numZs, *g.GerberLP)
		}
	}
}

func atof(s string) float64 {
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		log.Fatalf("unable to parse %q as float64", s)
	}
	return v
}

func parseParams(d string) (result []float64) {
	for len(d) > 0 {
		m := numRE.FindStringSubmatch(d)
		if len(m) == 2 {
			result = append(result, atof(m[1]))
			d = d[len(m[0]):]
			continue
		}
		log.Fatalf("parseParams: unable to parse %q", d)
	}
	return result
}
