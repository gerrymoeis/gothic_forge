package webawesome

import (
	"encoding/json"
	"fmt"
	"log"
	"time"
)

// PerformanceMonitor provides simple performance monitoring for Web Awesome components
type PerformanceMonitor struct {
	metrics map[string]interface{}
	enabled bool
}

// NewPerformanceMonitor creates a new performance monitor
func NewPerformanceMonitor(enabled bool) *PerformanceMonitor {
	return &PerformanceMonitor{
		metrics: make(map[string]interface{}),
		enabled: enabled,
	}
}

// TrackComponentLoad tracks component loading performance
func (pm *PerformanceMonitor) TrackComponentLoad(componentName string, duration time.Duration) {
	if !pm.enabled {
		return
	}
	
	key := fmt.Sprintf("component_load_%s", componentName)
	pm.metrics[key] = map[string]interface{}{
		"duration_ms": float64(duration.Nanoseconds()) / 1e6,
		"timestamp":   time.Now().Unix(),
	}
	
	// Log in development
	log.Printf("Component %s loaded in %.2fms", componentName, float64(duration.Nanoseconds())/1e6)
}

// TrackThemeSwitch tracks theme switching performance
func (pm *PerformanceMonitor) TrackThemeSwitch(fromTheme, toTheme string, duration time.Duration) {
	if !pm.enabled {
		return
	}
	
	pm.metrics["theme_switch"] = map[string]interface{}{
		"from":        fromTheme,
		"to":          toTheme,
		"duration_ms": float64(duration.Nanoseconds()) / 1e6,
		"timestamp":   time.Now().Unix(),
	}
}

// TrackPageLoad tracks overall page load performance
func (pm *PerformanceMonitor) TrackPageLoad(duration time.Duration) {
	if !pm.enabled {
		return
	}
	
	pm.metrics["page_load"] = map[string]interface{}{
		"duration_ms": float64(duration.Nanoseconds()) / 1e6,
		"timestamp":   time.Now().Unix(),
	}
}

// GetMetrics returns current performance metrics
func (pm *PerformanceMonitor) GetMetrics() map[string]interface{} {
	return pm.metrics
}

// GetMetricsJSON returns metrics as JSON string
func (pm *PerformanceMonitor) GetMetricsJSON() (string, error) {
	data, err := json.MarshalIndent(pm.metrics, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// GeneratePerformanceReport creates a simple performance report
func (pm *PerformanceMonitor) GeneratePerformanceReport() string {
	if !pm.enabled || len(pm.metrics) == 0 {
		return "Performance monitoring disabled or no metrics available."
	}
	
	report := "# Web Awesome Performance Report\n\n"
	
	// Page load metrics
	if pageLoad, exists := pm.metrics["page_load"]; exists {
		if data, ok := pageLoad.(map[string]interface{}); ok {
			if duration, ok := data["duration_ms"].(float64); ok {
				report += fmt.Sprintf("**Page Load Time**: %.2fms\n", duration)
			}
		}
	}
	
	// Theme switch metrics
	if themeSwitch, exists := pm.metrics["theme_switch"]; exists {
		if data, ok := themeSwitch.(map[string]interface{}); ok {
			if duration, ok := data["duration_ms"].(float64); ok {
				report += fmt.Sprintf("**Theme Switch Time**: %.2fms\n", duration)
			}
		}
	}
	
	// Component load metrics
	report += "\n## Component Load Times\n\n"
	for key, value := range pm.metrics {
		if len(key) > 15 && key[:15] == "component_load_" {
			componentName := key[15:]
			if data, ok := value.(map[string]interface{}); ok {
				if duration, ok := data["duration_ms"].(float64); ok {
					report += fmt.Sprintf("- **%s**: %.2fms\n", componentName, duration)
				}
			}
		}
	}
	
	return report
}

// ResourceOptimizer provides simple resource optimization suggestions
type ResourceOptimizer struct {
	usageStats map[string]int
}

// NewResourceOptimizer creates a new resource optimizer
func NewResourceOptimizer() *ResourceOptimizer {
	return &ResourceOptimizer{
		usageStats: make(map[string]int),
	}
}

// TrackResourceUsage tracks resource usage
func (ro *ResourceOptimizer) TrackResourceUsage(resourceType, resourceName string) {
	key := fmt.Sprintf("%s:%s", resourceType, resourceName)
	ro.usageStats[key]++
}

// GetOptimizationSuggestions returns simple optimization suggestions
func (ro *ResourceOptimizer) GetOptimizationSuggestions() []string {
	var suggestions []string
	
	if len(ro.usageStats) == 0 {
		return []string{"No resource usage data available."}
	}
	
	// Analyze component usage
	componentUsage := make(map[string]int)
	for key, count := range ro.usageStats {
		if len(key) > 10 && key[:10] == "component:" {
			componentName := key[10:]
			componentUsage[componentName] = count
		}
	}
	
	// Generate suggestions based on usage patterns
	for component, count := range componentUsage {
		if count > 10 {
			suggestions = append(suggestions, 
				fmt.Sprintf("Component '%s' is heavily used (%d times). Consider eager loading for better performance.", component, count))
		} else if count == 1 {
			suggestions = append(suggestions, 
				fmt.Sprintf("Component '%s' is used only once. Consider lazy loading to reduce initial bundle size.", component))
		}
	}
	
	if len(suggestions) == 0 {
		suggestions = append(suggestions, "Resource usage appears optimal. No specific optimizations needed.")
	}
	
	return suggestions
}

// CacheOptimizer provides simple cache optimization
type CacheOptimizer struct {
	cacheHits   int
	cacheMisses int
}

// NewCacheOptimizer creates a new cache optimizer
func NewCacheOptimizer() *CacheOptimizer {
	return &CacheOptimizer{}
}

// RecordCacheHit records a cache hit
func (co *CacheOptimizer) RecordCacheHit() {
	co.cacheHits++
}

// RecordCacheMiss records a cache miss
func (co *CacheOptimizer) RecordCacheMiss() {
	co.cacheMisses++
}

// GetCacheEfficiency returns cache efficiency percentage
func (co *CacheOptimizer) GetCacheEfficiency() float64 {
	total := co.cacheHits + co.cacheMisses
	if total == 0 {
		return 0
	}
	return float64(co.cacheHits) / float64(total) * 100
}

// GetCacheReport returns a simple cache report
func (co *CacheOptimizer) GetCacheReport() string {
	efficiency := co.GetCacheEfficiency()
	total := co.cacheHits + co.cacheMisses
	
	report := fmt.Sprintf("# Cache Performance Report\n\n")
	report += fmt.Sprintf("**Total Requests**: %d\n", total)
	report += fmt.Sprintf("**Cache Hits**: %d\n", co.cacheHits)
	report += fmt.Sprintf("**Cache Misses**: %d\n", co.cacheMisses)
	report += fmt.Sprintf("**Cache Efficiency**: %.1f%%\n\n", efficiency)
	
	if efficiency < 70 {
		report += "**Recommendation**: Cache efficiency is below 70%. Consider increasing cache duration or improving cache strategy.\n"
	} else if efficiency > 90 {
		report += "**Status**: Excellent cache performance!\n"
	} else {
		report += "**Status**: Good cache performance.\n"
	}
	
	return report
}

// PerformanceConfig holds performance optimization configuration
type PerformanceConfig struct {
	EnableMonitoring bool          `json:"enable_monitoring"`
	EnableCaching    bool          `json:"enable_caching"`
	CacheDuration    time.Duration `json:"cache_duration"`
	LazyLoading      bool          `json:"lazy_loading"`
	Preloading       bool          `json:"preloading"`
}

// DefaultPerformanceConfig returns default performance configuration
func DefaultPerformanceConfig() PerformanceConfig {
	return PerformanceConfig{
		EnableMonitoring: true,
		EnableCaching:    true,
		CacheDuration:    24 * time.Hour,
		LazyLoading:      true,
		Preloading:       true,
	}
}

// PerformanceManager manages all performance optimizations
type PerformanceManager struct {
	config    PerformanceConfig
	monitor   *PerformanceMonitor
	optimizer *ResourceOptimizer
	cache     *CacheOptimizer
}

// NewPerformanceManager creates a new performance manager
func NewPerformanceManager(config PerformanceConfig) *PerformanceManager {
	return &PerformanceManager{
		config:    config,
		monitor:   NewPerformanceMonitor(config.EnableMonitoring),
		optimizer: NewResourceOptimizer(),
		cache:     NewCacheOptimizer(),
	}
}

// GetMonitor returns the performance monitor
func (pm *PerformanceManager) GetMonitor() *PerformanceMonitor {
	return pm.monitor
}

// GetOptimizer returns the resource optimizer
func (pm *PerformanceManager) GetOptimizer() *ResourceOptimizer {
	return pm.optimizer
}

// GetCache returns the cache optimizer
func (pm *PerformanceManager) GetCache() *CacheOptimizer {
	return pm.cache
}

// GenerateFullReport generates a comprehensive performance report
func (pm *PerformanceManager) GenerateFullReport() string {
	report := "# Gothic Forge Web Awesome Performance Report\n\n"
	report += "Generated at: " + time.Now().Format(time.RFC3339) + "\n\n"
	
	// Performance metrics
	report += pm.monitor.GeneratePerformanceReport() + "\n\n"
	
	// Cache report
	report += pm.cache.GetCacheReport() + "\n"
	
	// Optimization suggestions
	suggestions := pm.optimizer.GetOptimizationSuggestions()
	report += "## Optimization Suggestions\n\n"
	for i, suggestion := range suggestions {
		report += fmt.Sprintf("%d. %s\n", i+1, suggestion)
	}
	
	return report
}