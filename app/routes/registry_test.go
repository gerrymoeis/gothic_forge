package routes

import (
	"testing"
	"time"
)

func TestRegisterURLAndMeta_ListURLInfo(t *testing.T) {
	// reset internal map via package variable reinit (allowed within same package tests)
	urlRegistry = map[string]URLInfo{}

	RegisterURL("/about")
	RegisterURLMeta("/pricing", time.Date(2025, 10, 4, 0, 0, 0, 0, time.UTC), "weekly", 0.8)

	infos := ListURLInfo()
	// must include root
	var hasRoot, hasAbout, hasPricing bool
	var pricing URLInfo
	for _, u := range infos {
		switch u.Path {
		case "/":
			hasRoot = true
		case "/about":
			hasAbout = true
		case "/pricing":
			hasPricing = true
			pricing = u
		}
	}
	if !hasRoot { t.Fatalf("expected root to be included by default") }
	if !hasAbout { t.Fatalf("expected /about to be included") }
	if !hasPricing { t.Fatalf("expected /pricing to be included") }
	if pricing.LastMod != "2025-10-04" || pricing.ChangeFreq != "weekly" || pricing.Priority != "0.8" {
		t.Fatalf("pricing metadata mismatch: %+v", pricing)
	}
}
