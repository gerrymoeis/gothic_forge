(function(){
  function getCookie(name){
    var m = document.cookie.match('(^|;)\\s*' + name + '\\s*=\\s*([^;]+)');
    return m ? m.pop() : '';
  }
  function onConfig(evt){
    var token = getCookie('_gforge_csrf');
    if (token) {
      evt.detail.headers['X-CSRF-Token'] = token;
    }
  }
  if (document.body) {
    document.body.addEventListener('htmx:configRequest', onConfig);
  } else {
    document.addEventListener('DOMContentLoaded', function(){
      document.body && document.body.addEventListener('htmx:configRequest', onConfig);
    });
  }

  // Prefetch on hover/focus for links with data-prefetch="1" (same-origin only)
  // Minimal, CSP-friendly: uses fetch HEAD and respects connect-src 'self'
  var __gforge_prefetch = Object.create(null);
  var __gforge_hoverTimer = null;

  function findAnchor(el){
    while (el && el !== document.body) {
      if (el.tagName && el.tagName.toLowerCase() === 'a') return el;
      el = el.parentNode;
    }
    return null;
  }
  function prefetchURL(url){
    try {
      if (__gforge_prefetch[url]) return; // already primed
      __gforge_prefetch[url] = 1;
      fetch(url, { method: 'HEAD', credentials: 'same-origin' }).catch(function(){ /* ignore */ });
    } catch (_) { /* ignore */ }
  }
  function onIntent(e){
    var a = findAnchor(e.target);
    if (!a || a.dataset.prefetch !== '1') return;
    var href = a.getAttribute('href');
    if (!href || href.charAt(0) === '#' || href.startsWith('mailto:') || href.startsWith('tel:')) return;
    try {
      var u = new URL(href, window.location.origin);
      if (u.origin !== window.location.origin) return; // same-origin only
      if (__gforge_hoverTimer) clearTimeout(__gforge_hoverTimer);
      __gforge_hoverTimer = setTimeout(function(){ prefetchURL(u.href); }, 100);
    } catch (_) { /* ignore */ }
  }
  document.addEventListener('mouseover', onIntent, { passive: true });
  document.addEventListener('focusin', onIntent);
})();
