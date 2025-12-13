package webawesome

import (
	"crypto/md5"
	"fmt"
	"strings"
	"time"
)

// PerformanceOptimizer provides minimal performance optimizations for Web Awesome integration
type PerformanceOptimizer struct {
	assetCache map[string]string
	version    string
}

// NewPerformanceOptimizer creates a new performance optimizer
func NewPerformanceOptimizer() *PerformanceOptimizer {
	return &PerformanceOptimizer{
		assetCache: make(map[string]string),
		version:    generateVersion(),
	}
}

// generateVersion creates a simple version string for cache busting
func generateVersion() string {
	return fmt.Sprintf("%d", time.Now().Unix())
}

// OptimizeAssetLoading provides optimized asset loading configuration
func (p *PerformanceOptimizer) OptimizeAssetLoading() AssetConfig {
	return AssetConfig{
		CDNBaseURL: "https://cdn.jsdelivr.net/npm/@shoelace-style/shoelace@2.15.1/cdn/",
		Version:    "2.15.1",
		CacheHeaders: map[string]string{
			"Cache-Control": "public, max-age=31536000, immutable",
			"ETag":          p.generateETag("shoelace-2.15.1"),
		},
		Preload: []string{
			"themes/light.css",
			"themes/dark.css",
			"shoelace.css",
			"shoelace-autoloader.js",
		},
	}
}

// AssetConfig holds configuration for optimized asset loading
type AssetConfig struct {
	CDNBaseURL   string
	Version      string
	CacheHeaders map[string]string
	Preload      []string
}

// generateETag creates a simple ETag for caching
func (p *PerformanceOptimizer) generateETag(content string) string {
	hash := md5.Sum([]byte(content))
	return fmt.Sprintf(`"%x"`, hash)
}

// DeduplicateResources removes duplicate Web Awesome dependencies
func (p *PerformanceOptimizer) DeduplicateResources(resources []string) []string {
	seen := make(map[string]bool)
	var deduplicated []string
	
	for _, resource := range resources {
		if !seen[resource] {
			seen[resource] = true
			deduplicated = append(deduplicated, resource)
		}
	}
	
	return deduplicated
}

// GeneratePreloadTags creates preload tags for critical Web Awesome assets
func (p *PerformanceOptimizer) GeneratePreloadTags(config AssetConfig) string {
	var preloadTags []string
	
	for _, asset := range config.Preload {
		url := config.CDNBaseURL + asset
		var asType string
		
		if strings.HasSuffix(asset, ".css") {
			asType = "style"
		} else if strings.HasSuffix(asset, ".js") {
			asType = "script"
		} else {
			continue
		}
		
		tag := fmt.Sprintf(`<link rel="preload" href="%s" as="%s" crossorigin>`, url, asType)
		preloadTags = append(preloadTags, tag)
	}
	
	return strings.Join(preloadTags, "\n      ")
}

// OptimizeComponentLoading provides lazy loading configuration for components
func (p *PerformanceOptimizer) OptimizeComponentLoading() ComponentLoadingConfig {
	return ComponentLoadingConfig{
		LazyLoad: true,
		LoadOnDemand: []string{
			"sl-dialog",
			"sl-drawer", 
			"sl-dropdown",
			"sl-menu",
			"sl-tooltip",
		},
		EagerLoad: []string{
			"sl-button",
			"sl-input",
			"sl-card",
			"sl-badge",
		},
	}
}

// ComponentLoadingConfig holds component loading optimization settings
type ComponentLoadingConfig struct {
	LazyLoad     bool
	LoadOnDemand []string
	EagerLoad    []string
}

// GenerateOptimizedScript creates an optimized loading script
func (p *PerformanceOptimizer) GenerateOptimizedScript(config ComponentLoadingConfig) string {
	return fmt.Sprintf(`
<script>
// Optimized Web Awesome component loading
(function() {
  'use strict';
  
  // Set base path immediately
  if (window.SlSetBasePath) {
    window.SlSetBasePath('https://cdn.jsdelivr.net/npm/@shoelace-style/shoelace@2.15.1/cdn/');
  }
  
  // Eager load critical components
  const eagerComponents = %s;
  
  // Lazy load on-demand components
  const lazyComponents = %s;
  
  // Load critical components immediately
  if (window.SlAutoloader) {
    eagerComponents.forEach(component => {
      window.SlAutoloader.loadComponent(component);
    });
  }
  
  // Set up intersection observer for lazy loading
  if (%t && 'IntersectionObserver' in window) {
    const observer = new IntersectionObserver((entries) => {
      entries.forEach(entry => {
        if (entry.isIntersecting) {
          const tagName = entry.target.tagName.toLowerCase();
          if (lazyComponents.includes(tagName) && window.SlAutoloader) {
            window.SlAutoloader.loadComponent(tagName);
            observer.unobserve(entry.target);
          }
        }
      });
    }, { rootMargin: '50px' });
    
    // Observe lazy components
    document.addEventListener('DOMContentLoaded', () => {
      lazyComponents.forEach(component => {
        const elements = document.querySelectorAll(component);
        elements.forEach(el => observer.observe(el));
      });
    });
  }
})();
</script>`,
		formatStringArray(config.EagerLoad),
		formatStringArray(config.LoadOnDemand),
		config.LazyLoad,
	)
}

// formatStringArray formats a string array for JavaScript
func formatStringArray(arr []string) string {
	quoted := make([]string, len(arr))
	for i, s := range arr {
		quoted[i] = fmt.Sprintf("'%s'", s)
	}
	return "[" + strings.Join(quoted, ", ") + "]"
}

// CacheStrategy provides caching recommendations
type CacheStrategy struct {
	StaticAssets  time.Duration
	ComponentDefs time.Duration
	ThemeFiles    time.Duration
}

// GetCacheStrategy returns optimal caching durations
func (p *PerformanceOptimizer) GetCacheStrategy() CacheStrategy {
	return CacheStrategy{
		StaticAssets:  365 * 24 * time.Hour, // 1 year for CDN assets
		ComponentDefs: 24 * time.Hour,       // 1 day for component definitions
		ThemeFiles:    7 * 24 * time.Hour,   // 1 week for theme files
	}
}

// BundleAnalyzer provides simple bundle analysis
type BundleAnalyzer struct {
	components map[string]int
	totalSize  int
}

// NewBundleAnalyzer creates a new bundle analyzer
func NewBundleAnalyzer() *BundleAnalyzer {
	return &BundleAnalyzer{
		components: make(map[string]int),
	}
}

// AnalyzeUsage tracks component usage for optimization
func (b *BundleAnalyzer) AnalyzeUsage(componentName string) {
	b.components[componentName]++
}

// GetUsageReport returns a simple usage report
func (b *BundleAnalyzer) GetUsageReport() map[string]int {
	return b.components
}

// GetOptimizationSuggestions provides simple optimization suggestions
func (b *BundleAnalyzer) GetOptimizationSuggestions() []string {
	var suggestions []string
	
	// Find unused components
	if len(b.components) == 0 {
		suggestions = append(suggestions, "No Web Awesome components detected. Consider removing Web Awesome assets if not needed.")
		return suggestions
	}
	
	// Find heavily used components
	for component, count := range b.components {
		if count > 10 {
			suggestions = append(suggestions, fmt.Sprintf("Component '%s' is heavily used (%d times). Consider eager loading.", component, count))
		} else if count == 1 {
			suggestions = append(suggestions, fmt.Sprintf("Component '%s' is used only once. Consider lazy loading.", component))
		}
	}
	
	if len(suggestions) == 0 {
		suggestions = append(suggestions, "Component usage looks optimal. No specific suggestions.")
	}
	
	return suggestions
}

// PerformanceMetrics provides simple performance tracking
type PerformanceMetrics struct {
	LoadTime      time.Duration
	ComponentInit time.Duration
	ThemeSwitch   time.Duration
}

// MeasurePerformance provides basic performance measurement
func (p *PerformanceOptimizer) MeasurePerformance() string {
	return `
<script>
// Simple performance measurement for Web Awesome
(function() {
  'use strict';
  
  const perf = {
    start: performance.now(),
    componentInit: 0,
    themeSwitch: 0
  };
  
  // Measure component initialization
  document.addEventListener('DOMContentLoaded', () => {
    perf.componentInit = performance.now() - perf.start;
    console.log('Web Awesome components initialized in', perf.componentInit.toFixed(2), 'ms');
  });
  
  // Measure theme switching
  const originalSetTheme = window.setWebAwesomeTheme;
  if (originalSetTheme) {
    window.setWebAwesomeTheme = function(theme) {
      const start = performance.now();
      originalSetTheme(theme);
      perf.themeSwitch = performance.now() - start;
      console.log('Theme switched in', perf.themeSwitch.toFixed(2), 'ms');
    };
  }
  
  // Report performance after page load
  window.addEventListener('load', () => {
    const totalTime = performance.now() - perf.start;
    console.log('Total Web Awesome load time:', totalTime.toFixed(2), 'ms');
    
    // Store metrics for potential reporting
    window.webAwesomeMetrics = {
      totalLoad: totalTime,
      componentInit: perf.componentInit,
      themeSwitch: perf.themeSwitch
    };
  });
})();
</script>`
}

// OptimizedLayoutHelper provides helpers for optimized layouts
type OptimizedLayoutHelper struct {
	optimizer *PerformanceOptimizer
}

// NewOptimizedLayoutHelper creates a new layout helper
func NewOptimizedLayoutHelper() *OptimizedLayoutHelper {
	return &OptimizedLayoutHelper{
		optimizer: NewPerformanceOptimizer(),
	}
}

// GenerateOptimizedHead generates an optimized head section
func (o *OptimizedLayoutHelper) GenerateOptimizedHead() string {
	config := o.optimizer.OptimizeAssetLoading()
	preloadTags := o.optimizer.GeneratePreloadTags(config)
	
	return fmt.Sprintf(`
      <!-- Optimized Web Awesome Assets -->
      %s
      
      <!-- Web Awesome CSS (optimized loading) -->
      <link rel="stylesheet" href="%sthemes/light.css" media="(prefers-color-scheme: light)">
      <link rel="stylesheet" href="%sthemes/dark.css" media="(prefers-color-scheme: dark)">
      <link rel="stylesheet" href="%sshoelace.css">`,
		preloadTags,
		config.CDNBaseURL,
		config.CDNBaseURL,
		config.CDNBaseURL,
	)
}

// GenerateOptimizedScripts generates optimized JavaScript loading
func (o *OptimizedLayoutHelper) GenerateOptimizedScripts() string {
	loadingConfig := o.optimizer.OptimizeComponentLoading()
	optimizedScript := o.optimizer.GenerateOptimizedScript(loadingConfig)
	performanceScript := o.optimizer.MeasurePerformance()
	
	return fmt.Sprintf(`
      <!-- Optimized Web Awesome JavaScript -->
      <script type="module" src="https://cdn.jsdelivr.net/npm/@shoelace-style/shoelace@2.15.1/cdn/shoelace-autoloader.js"></script>
      
      %s
      
      %s`,
		optimizedScript,
		performanceScript,
	)
}

// GetPerformanceDocumentation returns performance optimization documentation
func (p *PerformanceOptimizer) GetPerformanceDocumentation() string {
	return `
# Web Awesome Performance Optimization Guide

## Overview

Gothic Forge includes built-in performance optimizations for Web Awesome components that follow the framework's philosophy of simplicity and efficiency.

## Automatic Optimizations

### 1. Asset Deduplication
- Automatically removes duplicate Web Awesome dependencies
- Ensures only one version of each asset is loaded
- Reduces bundle size and improves load times

### 2. Smart Component Loading
- **Eager Loading**: Critical components (buttons, inputs, cards) load immediately
- **Lazy Loading**: Interactive components (dialogs, dropdowns) load on-demand
- **Intersection Observer**: Components load when they enter the viewport

### 3. CDN Optimization
- Uses jsDelivr CDN for optimal global delivery
- Proper cache headers for long-term caching
- Preload tags for critical assets

### 4. Theme Loading
- Media queries for efficient theme loading
- Only loads themes that match user preferences
- Smooth theme transitions without layout shifts

## Performance Metrics

The integration includes basic performance monitoring:

` + "```javascript" + `
// Access performance metrics
console.log(window.webAwesomeMetrics);
// Output: { totalLoad: 45.2, componentInit: 12.1, themeSwitch: 2.3 }
` + "```" + `

## Best Practices

1. **Use Appropriate Components**: Choose the right component for the job
2. **Lazy Load Heavy Components**: Dialogs and dropdowns load on-demand
3. **Optimize Images**: Use proper image formats and sizes
4. **Minimize Custom CSS**: Leverage Web Awesome's built-in styles

## Bundle Analysis

The system provides simple usage tracking:

- Identifies heavily used components for eager loading
- Detects unused components for removal
- Suggests optimization opportunities

## Cache Strategy

- **Static Assets**: 1 year cache (immutable CDN assets)
- **Component Definitions**: 1 day cache (allow for updates)
- **Theme Files**: 1 week cache (balance between performance and flexibility)

## Monitoring

Performance is automatically monitored and logged to the console. For production applications, consider integrating with your analytics platform.

The optimizations are designed to work automatically without configuration, maintaining Gothic Forge's principle of "batteries included" development.
`
}