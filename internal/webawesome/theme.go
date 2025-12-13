package webawesome

// Theme represents a Web Awesome theme configuration
type Theme struct {
	Name         string            `json:"name" yaml:"name"`
	DisplayName  string            `json:"displayName" yaml:"displayName"`
	Description  string            `json:"description" yaml:"description"`
	Version      string            `json:"version" yaml:"version"`
	Colors       ThemeColors       `json:"colors" yaml:"colors"`
	Typography   ThemeTypography   `json:"typography" yaml:"typography"`
	Spacing      ThemeSpacing      `json:"spacing" yaml:"spacing"`
	BorderRadius ThemeBorderRadius `json:"borderRadius" yaml:"borderRadius"`
	Shadows      ThemeShadows      `json:"shadows" yaml:"shadows"`
	Custom       map[string]string `json:"custom" yaml:"custom"`
}

// ThemeColors defines the color palette for a theme
type ThemeColors struct {
	Primary     string `json:"primary" yaml:"primary"`
	Secondary   string `json:"secondary" yaml:"secondary"`
	Success     string `json:"success" yaml:"success"`
	Warning     string `json:"warning" yaml:"warning"`
	Danger      string `json:"danger" yaml:"danger"`
	Neutral     string `json:"neutral" yaml:"neutral"`
	Background  string `json:"background" yaml:"background"`
	Surface     string `json:"surface" yaml:"surface"`
	OnPrimary   string `json:"onPrimary" yaml:"onPrimary"`
	OnSecondary string `json:"onSecondary" yaml:"onSecondary"`
	OnSurface   string `json:"onSurface" yaml:"onSurface"`
}

// ThemeTypography defines typography settings for a theme
type ThemeTypography struct {
	FontFamily string `json:"fontFamily" yaml:"fontFamily"`
	FontSize   string `json:"fontSize" yaml:"fontSize"`
	LineHeight string `json:"lineHeight" yaml:"lineHeight"`
	FontWeight string `json:"fontWeight" yaml:"fontWeight"`
}

// ThemeSpacing defines spacing values for a theme
type ThemeSpacing struct {
	XSmall string `json:"xSmall" yaml:"xSmall"`
	Small  string `json:"small" yaml:"small"`
	Medium string `json:"medium" yaml:"medium"`
	Large  string `json:"large" yaml:"large"`
	XLarge string `json:"xLarge" yaml:"xLarge"`
}

// ThemeBorderRadius defines border radius values for a theme
type ThemeBorderRadius struct {
	Small  string `json:"small" yaml:"small"`
	Medium string `json:"medium" yaml:"medium"`
	Large  string `json:"large" yaml:"large"`
	XLarge string `json:"xLarge" yaml:"xLarge"`
	Pill   string `json:"pill" yaml:"pill"`
}

// ThemeShadows defines shadow values for a theme
type ThemeShadows struct {
	XSmall string `json:"xSmall" yaml:"xSmall"`
	Small  string `json:"small" yaml:"small"`
	Medium string `json:"medium" yaml:"medium"`
	Large  string `json:"large" yaml:"large"`
	XLarge string `json:"xLarge" yaml:"xLarge"`
}

// DefaultTheme returns a default Web Awesome theme configuration
func DefaultTheme() *Theme {
	return &Theme{
		Name:        "default",
		DisplayName: "Default Theme",
		Description: "Default Web Awesome theme for Gothic Forge",
		Version:     "1.0.0",
		Colors: ThemeColors{
			Primary:     "hsl(198.6 88.7% 48.4%)",
			Secondary:   "hsl(240 4.9% 83.9%)",
			Success:     "hsl(142.1 76.2% 36.3%)",
			Warning:     "hsl(47.9 95.8% 53.1%)",
			Danger:      "hsl(0 84.2% 60.2%)",
			Neutral:     "hsl(240 5.9% 10%)",
			Background:  "hsl(0 0% 100%)",
			Surface:     "hsl(0 0% 100%)",
			OnPrimary:   "hsl(0 0% 100%)",
			OnSecondary: "hsl(240 5.9% 10%)",
			OnSurface:   "hsl(240 5.9% 10%)",
		},
		Typography: ThemeTypography{
			FontFamily: "-apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Helvetica, Arial, sans-serif, 'Apple Color Emoji', 'Segoe UI Emoji', 'Segoe UI Symbol'",
			FontSize:   "1rem",
			LineHeight: "1.6",
			FontWeight: "400",
		},
		Spacing: ThemeSpacing{
			XSmall: "0.125rem",
			Small:  "0.25rem",
			Medium: "0.5rem",
			Large:  "1rem",
			XLarge: "1.5rem",
		},
		BorderRadius: ThemeBorderRadius{
			Small:  "0.1875rem",
			Medium: "0.25rem",
			Large:  "0.5rem",
			XLarge: "1rem",
			Pill:   "9999px",
		},
		Shadows: ThemeShadows{
			XSmall: "0 1px 2px hsl(240 3.8% 46.1% / 6%)",
			Small:  "0 1px 2px hsl(240 3.8% 46.1% / 12%)",
			Medium: "0 2px 4px hsl(240 3.8% 46.1% / 12%)",
			Large:  "0 2px 8px hsl(240 3.8% 46.1% / 12%)",
			XLarge: "0 4px 16px hsl(240 3.8% 46.1% / 12%)",
		},
		Custom: make(map[string]string),
	}
}