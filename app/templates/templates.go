package templates

import (
    "context"
    "io"
    "strings"

    templ "github.com/a-h/templ"
    "gothicforge3/internal/env"
)

func Index() templ.Component {
    body := templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
        // HERO
        _, _ = io.WriteString(w, `<section class="mx-auto max-w-7xl px-4 md:px-6 relative">`)
        _, _ = io.WriteString(w, `<div class="hero min-h-[60vh] text-center hero-orb">`)
        _, _ = io.WriteString(w, `<div class="hero-content flex-col">`)
        _, _ = io.WriteString(w, `<div class="badge badge-outline mb-3 border-white/20 text-white/80">New</div>`)
        _, _ = io.WriteString(w, `<h1 class="text-5xl md:text-7xl font-extrabold tracking-tight bg-gradient-to-r from-[#4F46E5] to-[#EC4899] bg-clip-text text-transparent">Gothic Forge v3</h1>`)
        _, _ = io.WriteString(w, `<p class="mt-4 max-w-2xl mx-auto opacity-80">Lean, batteries-included Go starter with Templ + HTMX + Tailwind + DaisyUI. No Node required for rendering.</p>`)
        _, _ = io.WriteString(w, `<div class="mt-6 flex gap-3 justify-center"><a href="#counter" class="btn btn-primary">Try the demo</a><a href="https://github.com/gerrymoeis/gothic_forge" target="_blank" rel="noopener" class="btn btn-outline">View source</a></div>`)
        // Auth links (only show Login if OAuth configured)
        oauthEnabled := strings.TrimSpace(env.Get("GITHUB_CLIENT_ID", "")) != "" && strings.TrimSpace(env.Get("GITHUB_CLIENT_SECRET", "")) != ""
        if oauthEnabled {
            _, _ = io.WriteString(w, `<div class="mt-3 text-sm opacity-90">`+
                `<a href="/auth/github/login" class="link link-hover text-primary">Sign in with GitHub</a>`+
                ` <span class="opacity-50">·</span> `+
                `<a href="/auth/logout" class="link link-hover">Logout</a>`+
                `</div>`)
        }
        _, _ = io.WriteString(w, `</div></div>`)
        _, _ = io.WriteString(w, `</section>`)

        // FEATURES
        _, _ = io.WriteString(w, `<section class="mx-auto max-w-7xl px-4 md:px-6 mt-12">`)
        _, _ = io.WriteString(w, `<div class="grid gap-6 md:grid-cols-3">`)
        // Card 1
        _, _ = io.WriteString(w, `<div class="card bg-base-200/60 border border-white/10 rounded-box shadow-xl ring-1 ring-white/10"><div class="card-body"><h3 class="card-title">Type-safe UI</h3><p>Build with Templ and Go — no runtime JS required for rendering.</p></div></div>`)
        // Card 2
        _, _ = io.WriteString(w, `<div class="card bg-base-200/60 border border-white/10 rounded-box shadow-xl ring-1 ring-white/10"><div class="card-body"><h3 class="card-title">Progressive interactivity</h3><p>HTMX for hypermedia, Alpine for local state where needed.</p></div></div>`)
        // Card 3
        _, _ = io.WriteString(w, `<div class="card bg-base-200/60 border border-white/10 rounded-box shadow-xl ring-1 ring-white/10"><div class="card-body"><h3 class="card-title">Zero Node toolchain</h3><p>Tailwind compiled with gotailwindcss; DaisyUI via CDN.</p></div></div>`)
        _, _ = io.WriteString(w, `</div></section>`)

        // STACK TRIBUTE
        _, _ = io.WriteString(w, `<section class="mx-auto max-w-7xl px-4 md:px-6 mt-12">`)
        _, _ = io.WriteString(w, `<div class="card bg-base-200/60 border border-white/10 rounded-box shadow-xl ring-1 ring-white/10">`)
        _, _ = io.WriteString(w, `<div class="card-body">`)
        _, _ = io.WriteString(w, `<h3 class="card-title">Core Stack</h3><p class="opacity-80">Type-safe UI and progressive interactivity.</p>`)
        _, _ = io.WriteString(w, `<div class="flex flex-wrap gap-2 mt-2">`)
        _, _ = io.WriteString(w, `<div class="badge badge-outline">Go</div>`)
        _, _ = io.WriteString(w, `<div class="badge badge-outline">Templ</div>`)
        _, _ = io.WriteString(w, `<div class="badge badge-outline">HTMX</div>`)
        _, _ = io.WriteString(w, `<div class="badge badge-outline">Alpine.js</div>`)
        _, _ = io.WriteString(w, `<div class="badge badge-outline">Tailwind CSS</div>`)
        _, _ = io.WriteString(w, `<div class="badge badge-outline">DaisyUI</div>`)
        _, _ = io.WriteString(w, `</div></div></div></section>`)

        // COUNTER DEMO
        _, _ = io.WriteString(w, `<section id="counter" class="mx-auto max-w-7xl px-4 md:px-6 mt-16">`)
        _, _ = io.WriteString(w, `<div class="card bg-base-200/60 border border-white/10 rounded-box shadow-xl ring-1 ring-white/10">`)
        _, _ = io.WriteString(w, `<div class="card-body">`)
        _, _ = io.WriteString(w, `<h2 class="card-title">Counter Demo</h2>`)
        _, _ = io.WriteString(w, `<div x-data="counter" class="grid gap-4">`)
        _, _ = io.WriteString(w, `<div class="stats bg-base-100 shadow">`)
        _, _ = io.WriteString(w, `<div class="stat"><div class="stat-title">Local (Alpine)</div><div class="stat-value" x-text="c">0</div><div class="stat-desc">increments instantly</div></div>`)
        _, _ = io.WriteString(w, `<div id="server-count" class="stat"><div class="stat-title">Server (HTMX)</div><div id="server-count-value" role="status" aria-live="polite" class="stat-value">0</div><div class="stat-desc">updates 5s after last click</div></div>`)
        _, _ = io.WriteString(w, `</div>`)
        _, _ = io.WriteString(w, `<div class="join"><button class="btn btn-primary join-item" @click="bump()">+1</button><button class="btn join-item" @click="reset()">Reset</button></div>`)
        _, _ = io.WriteString(w, `</div></div></section>`)

        // HOW IT WORKS
        _, _ = io.WriteString(w, `<section class="mx-auto max-w-7xl px-4 md:px-6 mt-16">`)
        _, _ = io.WriteString(w, `<ul class="steps steps-vertical md:steps-horizontal w-full">`)
        _, _ = io.WriteString(w, `<li class="step step-primary">Clone</li>`)
        _, _ = io.WriteString(w, `<li class="step step-primary">gforge dev</li>`)
        _, _ = io.WriteString(w, `<li class="step">Edit app/</li>`)
        _, _ = io.WriteString(w, `<li class="step">gforge deploy</li>`)
        _, _ = io.WriteString(w, `</ul></section>`)

        // CTA BAND
        _, _ = io.WriteString(w, `<section class="mx-auto max-w-7xl px-4 md:px-6 mt-16">`)
        _, _ = io.WriteString(w, `<div class="hero bg-base-200/60 rounded-box border border-white/10 ring-1 ring-white/10">`)
        _, _ = io.WriteString(w, `<div class="hero-content text-center">`)
        _, _ = io.WriteString(w, `<div class="max-w-2xl"><h3 class="text-3xl font-bold">Build fast with Gothic Forge</h3><p class="opacity-80 mt-2">Edit files in <code class='kbd'>/app</code>. Use <span class='badge badge-primary'>gforge</span> for everything else.</p><div class="mt-6 flex justify-center gap-3"><a href="#counter" class="btn btn-primary">Try counter</a><a href="https://github.com/gerrymoeis/gothic_forge" target="_blank" rel="noopener" class="btn btn-outline">View source</a></div></div>`)
        _, _ = io.WriteString(w, `</div></div></section>`)
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
