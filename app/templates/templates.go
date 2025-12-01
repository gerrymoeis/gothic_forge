package templates

import (
    "context"
    "fmt"
    "io"
    "strings"

    templ "github.com/a-h/templ"
    "gothicforge3/internal/env"
)

// writeHTML writes HTML content to the writer and returns any error encountered.
// This helper ensures consistent error handling across template rendering.
func writeHTML(w io.Writer, html string) error {
    if _, err := io.WriteString(w, html); err != nil {
        return fmt.Errorf("write HTML failed: %w", err)
    }
    return nil
}

func Index() templ.Component {
    body := templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
        // HERO
        if err := writeHTML(w, `<section class="mx-auto max-w-7xl px-4 md:px-6 relative">`); err != nil {
            return fmt.Errorf("Index template (hero section start): %w", err)
        }
        if err := writeHTML(w, `<div class="hero min-h-[60vh] text-center hero-orb">`); err != nil {
            return fmt.Errorf("Index template (hero div): %w", err)
        }
        if err := writeHTML(w, `<div class="hero-content flex-col">`); err != nil {
            return fmt.Errorf("Index template (hero content): %w", err)
        }
        if err := writeHTML(w, `<div class="badge badge-outline mb-3 border-white/20 text-white/80">New</div>`); err != nil {
            return fmt.Errorf("Index template (badge): %w", err)
        }
        if err := writeHTML(w, `<h1 class="text-5xl md:text-7xl font-extrabold tracking-tight bg-gradient-to-r from-[#4F46E5] to-[#EC4899] bg-clip-text text-transparent">Gothic Forge v3</h1>`); err != nil {
            return fmt.Errorf("Index template (title): %w", err)
        }
        if err := writeHTML(w, `<p class="mt-4 max-w-2xl mx-auto opacity-80">Lean, batteries-included Go starter with Templ + HTMX + Tailwind + DaisyUI. No Node required for rendering.</p>`); err != nil {
            return fmt.Errorf("Index template (description): %w", err)
        }
        if err := writeHTML(w, `<div class="mt-6 flex gap-3 justify-center"><a href="#counter" class="btn btn-primary">Try the demo</a><a href="https://github.com/gerrymoeis/gothic_forge" target="_blank" rel="noopener" class="btn btn-outline">View source</a></div>`); err != nil {
            return fmt.Errorf("Index template (CTA buttons): %w", err)
        }
        // Auth links (only show Login if OAuth configured)
        oauthEnabled := strings.TrimSpace(env.Get("GITHUB_CLIENT_ID", "")) != "" && strings.TrimSpace(env.Get("GITHUB_CLIENT_SECRET", "")) != ""
        if oauthEnabled {
            if err := writeHTML(w, `<div class="mt-3 text-sm opacity-90">`+
                `<a href="/auth/github/login" class="link link-hover text-primary">Sign in with GitHub</a>`+
                ` <span class="opacity-50">·</span> `+
                `<a href="/auth/logout" class="link link-hover">Logout</a>`+
                `</div>`); err != nil {
                return fmt.Errorf("Index template (auth links): %w", err)
            }
        }
        if err := writeHTML(w, `</div></div>`); err != nil {
            return fmt.Errorf("Index template (hero close divs): %w", err)
        }
        if err := writeHTML(w, `</section>`); err != nil {
            return fmt.Errorf("Index template (hero section close): %w", err)
        }

        // FEATURES
        if err := writeHTML(w, `<section class="mx-auto max-w-7xl px-4 md:px-6 mt-12">`); err != nil {
            return fmt.Errorf("Index template (features section start): %w", err)
        }
        if err := writeHTML(w, `<div class="grid gap-6 md:grid-cols-3">`); err != nil {
            return fmt.Errorf("Index template (features grid): %w", err)
        }
        // Card 1
        if err := writeHTML(w, `<div class="card bg-base-200/60 border border-white/10 rounded-box shadow-xl ring-1 ring-white/10"><div class="card-body"><h3 class="card-title">Type-safe UI</h3><p>Build with Templ and Go — no runtime JS required for rendering.</p></div></div>`); err != nil {
            return fmt.Errorf("Index template (feature card 1): %w", err)
        }
        // Card 2
        if err := writeHTML(w, `<div class="card bg-base-200/60 border border-white/10 rounded-box shadow-xl ring-1 ring-white/10"><div class="card-body"><h3 class="card-title">Progressive interactivity</h3><p>HTMX for hypermedia, Alpine for local state where needed.</p></div></div>`); err != nil {
            return fmt.Errorf("Index template (feature card 2): %w", err)
        }
        // Card 3
        if err := writeHTML(w, `<div class="card bg-base-200/60 border border-white/10 rounded-box shadow-xl ring-1 ring-white/10"><div class="card-body"><h3 class="card-title">Zero Node toolchain</h3><p>Tailwind compiled with gotailwindcss; DaisyUI via CDN.</p></div></div>`); err != nil {
            return fmt.Errorf("Index template (feature card 3): %w", err)
        }
        if err := writeHTML(w, `</div></section>`); err != nil {
            return fmt.Errorf("Index template (features section close): %w", err)
        }

        // STACK TRIBUTE
        if err := writeHTML(w, `<section class="mx-auto max-w-7xl px-4 md:px-6 mt-12">`); err != nil {
            return fmt.Errorf("Index template (stack section start): %w", err)
        }
        if err := writeHTML(w, `<div class="card bg-base-200/60 border border-white/10 rounded-box shadow-xl ring-1 ring-white/10">`); err != nil {
            return fmt.Errorf("Index template (stack card): %w", err)
        }
        if err := writeHTML(w, `<div class="card-body">`); err != nil {
            return fmt.Errorf("Index template (stack card body): %w", err)
        }
        if err := writeHTML(w, `<h3 class="card-title">Core Stack</h3><p class="opacity-80">Type-safe UI and progressive interactivity.</p>`); err != nil {
            return fmt.Errorf("Index template (stack title): %w", err)
        }
        if err := writeHTML(w, `<div class="flex flex-wrap gap-2 mt-2">`); err != nil {
            return fmt.Errorf("Index template (stack badges container): %w", err)
        }
        if err := writeHTML(w, `<div class="badge badge-outline">Go</div>`); err != nil {
            return fmt.Errorf("Index template (stack badge Go): %w", err)
        }
        if err := writeHTML(w, `<div class="badge badge-outline">Templ</div>`); err != nil {
            return fmt.Errorf("Index template (stack badge Templ): %w", err)
        }
        if err := writeHTML(w, `<div class="badge badge-outline">HTMX</div>`); err != nil {
            return fmt.Errorf("Index template (stack badge HTMX): %w", err)
        }
        if err := writeHTML(w, `<div class="badge badge-outline">Alpine.js</div>`); err != nil {
            return fmt.Errorf("Index template (stack badge Alpine): %w", err)
        }
        if err := writeHTML(w, `<div class="badge badge-outline">Tailwind CSS</div>`); err != nil {
            return fmt.Errorf("Index template (stack badge Tailwind): %w", err)
        }
        if err := writeHTML(w, `<div class="badge badge-outline">DaisyUI</div>`); err != nil {
            return fmt.Errorf("Index template (stack badge DaisyUI): %w", err)
        }
        if err := writeHTML(w, `</div></div></div></section>`); err != nil {
            return fmt.Errorf("Index template (stack section close): %w", err)
        }

        // COUNTER DEMO
        if err := writeHTML(w, `<section id="counter" class="mx-auto max-w-7xl px-4 md:px-6 mt-16">`); err != nil {
            return fmt.Errorf("Index template (counter section start): %w", err)
        }
        if err := writeHTML(w, `<div class="card bg-base-200/60 border border-white/10 rounded-box shadow-xl ring-1 ring-white/10">`); err != nil {
            return fmt.Errorf("Index template (counter card): %w", err)
        }
        if err := writeHTML(w, `<div class="card-body">`); err != nil {
            return fmt.Errorf("Index template (counter card body): %w", err)
        }
        if err := writeHTML(w, `<h2 class="card-title">Counter Demo</h2>`); err != nil {
            return fmt.Errorf("Index template (counter title): %w", err)
        }
        if err := writeHTML(w, `<div x-data="counter" class="grid gap-4">`); err != nil {
            return fmt.Errorf("Index template (counter Alpine container): %w", err)
        }
        if err := writeHTML(w, `<div class="stats bg-base-100 shadow">`); err != nil {
            return fmt.Errorf("Index template (counter stats): %w", err)
        }
        if err := writeHTML(w, `<div class="stat"><div class="stat-title">Local (Alpine)</div><div class="stat-value" x-text="c">0</div><div class="stat-desc">increments instantly</div></div>`); err != nil {
            return fmt.Errorf("Index template (counter local stat): %w", err)
        }
        if err := writeHTML(w, `<div id="server-count" class="stat"><div class="stat-title">Server (HTMX)</div><div id="server-count-value" role="status" aria-live="polite" class="stat-value">0</div><div class="stat-desc">updates 5s after last click</div></div>`); err != nil {
            return fmt.Errorf("Index template (counter server stat): %w", err)
        }
        if err := writeHTML(w, `</div>`); err != nil {
            return fmt.Errorf("Index template (counter stats close): %w", err)
        }
        if err := writeHTML(w, `<div class="join"><button class="btn btn-primary join-item" @click="bump()">+1</button><button class="btn join-item" @click="reset()">Reset</button></div>`); err != nil {
            return fmt.Errorf("Index template (counter buttons): %w", err)
        }
        if err := writeHTML(w, `</div></div></section>`); err != nil {
            return fmt.Errorf("Index template (counter section close): %w", err)
        }

        // HOW IT WORKS
        if err := writeHTML(w, `<section class="mx-auto max-w-7xl px-4 md:px-6 mt-16">`); err != nil {
            return fmt.Errorf("Index template (steps section start): %w", err)
        }
        if err := writeHTML(w, `<ul class="steps steps-vertical md:steps-horizontal w-full">`); err != nil {
            return fmt.Errorf("Index template (steps list): %w", err)
        }
        if err := writeHTML(w, `<li class="step step-primary">Clone</li>`); err != nil {
            return fmt.Errorf("Index template (step 1): %w", err)
        }
        if err := writeHTML(w, `<li class="step step-primary">gforge dev</li>`); err != nil {
            return fmt.Errorf("Index template (step 2): %w", err)
        }
        if err := writeHTML(w, `<li class="step">Edit app/</li>`); err != nil {
            return fmt.Errorf("Index template (step 3): %w", err)
        }
        if err := writeHTML(w, `<li class="step">gforge deploy</li>`); err != nil {
            return fmt.Errorf("Index template (step 4): %w", err)
        }
        if err := writeHTML(w, `</ul></section>`); err != nil {
            return fmt.Errorf("Index template (steps section close): %w", err)
        }

        // CTA BAND
        if err := writeHTML(w, `<section class="mx-auto max-w-7xl px-4 md:px-6 mt-16">`); err != nil {
            return fmt.Errorf("Index template (CTA section start): %w", err)
        }
        if err := writeHTML(w, `<div class="hero bg-base-200/60 rounded-box border border-white/10 ring-1 ring-white/10">`); err != nil {
            return fmt.Errorf("Index template (CTA hero): %w", err)
        }
        if err := writeHTML(w, `<div class="hero-content text-center">`); err != nil {
            return fmt.Errorf("Index template (CTA hero content): %w", err)
        }
        if err := writeHTML(w, `<div class="max-w-2xl"><h3 class="text-3xl font-bold">Build fast with Gothic Forge</h3><p class="opacity-80 mt-2">Edit files in <code class='kbd'>/app</code>. Use <span class='badge badge-primary'>gforge</span> for everything else.</p><div class="mt-6 flex justify-center gap-3"><a href="#counter" class="btn btn-primary">Try counter</a><a href="https://github.com/gerrymoeis/gothic_forge" target="_blank" rel="noopener" class="btn btn-outline">View source</a></div></div>`); err != nil {
            return fmt.Errorf("Index template (CTA content): %w", err)
        }
        if err := writeHTML(w, `</div></div></section>`); err != nil {
            return fmt.Errorf("Index template (CTA section close): %w", err)
        }
        return nil
    })
    return templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
        // Configurable SEO keywords via env, fallback to defaults
        kw := strings.TrimSpace(env.Get("SEO_KEYWORDS", ""))
        if kw == "" {
            kw = "Kompetisi pemrograman Indonesia, Pelatihan coding mahasiswa, Innovation Lab, Gemastik, Olivia competition, UI/UX design learning, Web development training, C++ programming education"
        }
        ctx = templ.WithChildren(ctx, body)
        return LayoutSEO(SEO{
            Title:       "Gothic Forge v3 — Lean Go starter (Templ + HTMX + Tailwind)",
            Description: "Lean, batteries-included Go starter with Templ + HTMX + Tailwind (no Node). Build fast, iterate faster.",
            Canonical:   "/",
            Image:       "",
            Keywords:    kw,
            JSONLD:      `{
  "@context": "https://schema.org",
  "@type": "WebSite",
  "name": "Gothic Forge v3",
  "url": "/",
  "potentialAction": {
    "@type": "SearchAction",
    "target": "/?q={search_term_string}",
    "query-input": "required name=search_term_string"
  }
}`,
        }).Render(ctx, w)
    })
}
