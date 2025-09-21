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
})();
