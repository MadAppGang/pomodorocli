package ui

import (
	"bufio"
	"embed"
	"fmt"
	"strings"
)

//go:embed fonts/*.flf
var fontFS embed.FS

// FigletFont represents a parsed Figlet font
type FigletFont struct {
	Name         string
	Height       int
	Hardblank    rune
	CharPatterns map[rune][]string
}

// FontManager handles the available fonts and current font selection
type FontManager struct {
	Fonts       map[string]*FigletFont
	CurrentFont string
	FontNames   []string
}

// NewFontManager creates a new font manager and loads the embedded fonts
func NewFontManager() (*FontManager, error) {
	manager := &FontManager{
		Fonts:     make(map[string]*FigletFont),
		FontNames: []string{},
	}

	// Load embedded fonts
	fontFiles, err := fontFS.ReadDir("fonts")
	if err != nil {
		return nil, fmt.Errorf("failed to read embedded fonts: %w", err)
	}

	for _, fontFile := range fontFiles {
		if !fontFile.IsDir() && strings.HasSuffix(fontFile.Name(), ".flf") {
			fontName := strings.TrimSuffix(fontFile.Name(), ".flf")
			fontPath := "fonts/" + fontFile.Name()

			fontData, err := fontFS.ReadFile(fontPath)
			if err != nil {
				fmt.Printf("Error reading font file %s: %v\n", fontPath, err)
				continue // Skip this font if it can't be read
			}

			font, err := parseFigletFont(fontName, string(fontData))
			if err != nil {
				fmt.Printf("Error parsing font %s: %v\n", fontName, err)
				continue // Skip this font if it can't be parsed
			}

			// Store with normalized key for consistent lookup
			normalizedName := strings.ReplaceAll(fontName, " ", "_")
			manager.Fonts[normalizedName] = font
			manager.FontNames = append(manager.FontNames, normalizedName)
		}
	}

	// Set DOS_Rebel as the default font if available, otherwise use the first font
	if len(manager.FontNames) > 0 {
		defaultFont := "DOS_Rebel"
		if _, exists := manager.Fonts[defaultFont]; exists {
			manager.CurrentFont = defaultFont
		} else {
			manager.CurrentFont = manager.FontNames[0]
		}
	}

	return manager, nil
}

// NextFont switches to the next available font
func (fm *FontManager) NextFont() {
	if len(fm.FontNames) <= 1 {
		return // Nothing to switch to
	}

	// Find the current index
	currentIndex := -1
	for i, name := range fm.FontNames {
		if name == fm.CurrentFont {
			currentIndex = i
			break
		}
	}

	// Switch to the next font
	nextIndex := (currentIndex + 1) % len(fm.FontNames)
	fm.CurrentFont = fm.FontNames[nextIndex]
}

// GetCurrentFont returns the currently selected font
func (fm *FontManager) GetCurrentFont() *FigletFont {
	if font, exists := fm.Fonts[fm.CurrentFont]; exists {
		return font
	}

	// If the current font doesn't exist (shouldn't happen), use the first one
	if len(fm.FontNames) > 0 {
		fm.CurrentFont = fm.FontNames[0]
		return fm.Fonts[fm.CurrentFont]
	}

	return nil // No fonts available
}

// RenderDigit returns the pattern for a specific digit in the current font
func (fm *FontManager) RenderDigit(digit rune) []string {
	font := fm.GetCurrentFont()
	if font == nil {
		return []string{} // No font available
	}

	pattern, exists := font.CharPatterns[digit]
	if !exists {
		// If the digit isn't in the font, return empty lines
		emptyPattern := make([]string, font.Height)
		for i := range emptyPattern {
			emptyPattern[i] = strings.Repeat(" ", 10) // Default width
		}
		return emptyPattern
	}

	return pattern
}

// parseFigletFont parses a Figlet font file
func parseFigletFont(name string, data string) (*FigletFont, error) {
	scanner := bufio.NewScanner(strings.NewReader(data))
	if !scanner.Scan() {
		return nil, fmt.Errorf("empty font file")
	}

	// Parse the header line
	header := scanner.Text()
	if !strings.HasPrefix(header, "flf2") {
		return nil, fmt.Errorf("not a valid Figlet font")
	}

	parts := strings.Split(header, " ")
	if len(parts) < 5 {
		return nil, fmt.Errorf("invalid font header format")
	}

	// The header format is: flf2a$ height baseLine maxLength ...
	height := 0
	_, err := fmt.Sscanf(parts[1], "%d", &height)
	if err != nil || height <= 0 {
		return nil, fmt.Errorf("invalid font height")
	}

	hardblank := ' '
	if len(parts[0]) > 4 {
		hardblank = rune(parts[0][4])
	}

	font := &FigletFont{
		Name:         name,
		Height:       height,
		Hardblank:    hardblank,
		CharPatterns: make(map[rune][]string),
	}

	// Skip comment lines
	commentLines := 0
	if len(parts) >= 6 {
		_, err := fmt.Sscanf(parts[5], "%d", &commentLines)
		if err == nil {
			for i := 0; i < commentLines && scanner.Scan(); i++ {
				// Skip comment line
			}
		}
	}

	// Parse the character patterns
	// First character is ASCII 32 (space)
	currentChar := rune(32)
	charPattern := make([]string, 0, height)

	// Read all lines
	for scanner.Scan() {
		line := scanner.Text()

		// Hardblank replacement
		line = strings.ReplaceAll(line, string(hardblank), " ")

		// Process end markers for character definition
		if strings.HasSuffix(line, "@@") {
			// Complete character definition with @@
			line = strings.TrimSuffix(line, "@@")

			// Replace placeholder $ characters with spaces
			line = strings.ReplaceAll(line, "$", " ")

			charPattern = append(charPattern, line)
			font.CharPatterns[currentChar] = charPattern
			charPattern = make([]string, 0, height)
			currentChar++
			continue
		} else if strings.HasSuffix(line, "@") {
			// Line ends with @ marker
			line = strings.TrimSuffix(line, "@")

			// Replace placeholder $ characters with spaces
			line = strings.ReplaceAll(line, "$", " ")

			charPattern = append(charPattern, line)

			// If we've collected all lines for this character, store it and move to next
			if len(charPattern) >= height {
				font.CharPatterns[currentChar] = charPattern
				charPattern = make([]string, 0, height)
				currentChar++
			}
			continue
		}

		// Regular line (shouldn't normally happen in well-formed FLF files)
		// Replace any $ placeholders and add to pattern
		line = strings.ReplaceAll(line, "$", " ")
		charPattern = append(charPattern, line)
	}

	// Ensure we have at least the digits 0-9 and colon for the timer
	requiredChars := "0123456789:"
	for _, char := range requiredChars {
		if _, exists := font.CharPatterns[char]; !exists {
			return nil, fmt.Errorf("font missing required character: %c", char)
		}
	}

	return font, nil
}

// RenderTimeString renders a time string (e.g. "25:00") using the current font
func (fm *FontManager) RenderTimeString(timeStr string) string {
	font := fm.GetCurrentFont()
	if font == nil {
		return timeStr // Fallback to the original string
	}

	// Initialize an array for each line of the result
	lines := make([]string, font.Height)

	// Add each character pattern
	for _, char := range timeStr {
		pattern := fm.RenderDigit(char)

		// Append each line of this character to the corresponding result line
		for i := 0; i < font.Height && i < len(pattern); i++ {
			lines[i] += pattern[i]
		}
	}

	// Join the lines with newlines
	return strings.Join(lines, "\n")
}
