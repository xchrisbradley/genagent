package core

import (
	"github.com/fatih/color"
)

// ColorScheme defines the color configuration for different message types
type ColorScheme struct {
	InfoColor    *color.Color
	DebugColor   *color.Color
	ErrorColor   *color.Color
	SuccessColor *color.Color
	WarningColor *color.Color
}

// DefaultColorScheme returns the default color configuration
func DefaultColorScheme() *ColorScheme {
	return &ColorScheme{
		InfoColor:    color.New(color.FgCyan),
		DebugColor:   color.New(color.FgMagenta),
		ErrorColor:   color.New(color.FgRed),
		SuccessColor: color.New(color.FgGreen),
		WarningColor: color.New(color.FgYellow),
	}
}

// ColorPrintf prints a formatted message with the specified color
func (cs *ColorScheme) ColorPrintf(c *color.Color, format string, a ...interface{}) {
	c.Printf(format+"\n", a...)
}

// Info prints an info message in cyan
func (cs *ColorScheme) Info(format string, a ...interface{}) {
	cs.ColorPrintf(cs.InfoColor, format, a...)
}

// Debug prints a debug message in magenta
func (cs *ColorScheme) Debug(format string, a ...interface{}) {
	cs.ColorPrintf(cs.DebugColor, format, a...)
}

// Error prints an error message in red
func (cs *ColorScheme) Error(format string, a ...interface{}) {
	cs.ColorPrintf(cs.ErrorColor, format, a...)
}

// Success prints a success message in green
func (cs *ColorScheme) Success(format string, a ...interface{}) {
	cs.ColorPrintf(cs.SuccessColor, format, a...)
}

// Warning prints a warning message in yellow
func (cs *ColorScheme) Warning(format string, a ...interface{}) {
	cs.ColorPrintf(cs.WarningColor, format, a...)
}
