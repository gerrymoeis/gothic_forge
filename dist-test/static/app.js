// Alpine component definitions for Gothic Forge v3
// Works with Alpine CSP build (no inline/eval).

document.addEventListener('alpine:init', () => {
  Alpine.data('counter', () => ({
    c: 0,
    t: null,
    bump() {
      this.c++;
      this.schedule();
    },
    reset() {
      this.c = 0;
      if (this.t) clearTimeout(this.t);
      this.t = null;
      if (window.htmx) {
        htmx.ajax('POST', '/counter/sync', {
          target: '#server-count-value',
          swap: 'innerHTML',
          values: { count: this.c },
        });
      }
    },
    schedule() {
      if (this.t) clearTimeout(this.t);
      this.t = setTimeout(() => {
        if (window.htmx) {
          htmx.ajax('POST', '/counter/sync', {
            target: '#server-count-value',
            swap: 'innerHTML',
            values: { count: this.c },
          });
        }
      }, 5000);
    },
  }));
});
