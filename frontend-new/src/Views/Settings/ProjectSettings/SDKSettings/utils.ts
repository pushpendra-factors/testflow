export const generateSdkScriptCode = (
  assetURL: string,
  projectToken: string,
  apiURL: string,
  includeScriptTag = true
) => {
  const codeWithOutScriptTag = `window.faitracker=window.faitracker||function(){this.q=[];var t=new CustomEvent("FAITRACKER_QUEUED_EVENT");return this.init=function(t,e,a){this.TOKEN=t,this.INIT_PARAMS=e,this.INIT_CALLBACK=a,window.dispatchEvent(new CustomEvent("FAITRACKER_INIT_EVENT"))},this.call=function(){var e={k:"",a:[]};if(arguments&&arguments.length>=1){for(var a=1;a<arguments.length;a++)e.a.push(arguments[a]);e.k=arguments[0]}this.q.push(e),window.dispatchEvent(t)},this.message=function(){window.addEventListener("message",function(t){"faitracker"===t.data.origin&&this.call("message",t.data.type,t.data.message)})},this.message(),this.init("${projectToken}",{host:"${apiURL}"}),this}(),function(){var t=document.createElement("script");t.type="text/javascript",t.src="${assetURL}",t.async=!0,(d=document.getElementsByTagName("script")[0]).parentNode.insertBefore(t,d)}();`;
  if (includeScriptTag) return `<script>${codeWithOutScriptTag}</script>`;
  return codeWithOutScriptTag;
};

export const generateSdkScriptCodeForPdf = (
  assetURL: string,
  projectToken: string,
  apiURL: string,
  includeScriptTag = true
) => {
  const codeWithOutScriptTag = `window.faitracker = window.faitracker || function() { this.q = []; var t = new CustomEvent("FAITRACKER_QUEUED_EVENT"); return this.init = function(t, e, a) { this.TOKEN = t, this.INIT_PARAMS = e, this.INIT_CALLBACK = a, window.dispatchEvent(new CustomEvent("FAITRACKER_INIT_EVENT")) }, this.call = function() { var e = { k: "", a: [] }; if (arguments && arguments.length >= 1) { for (var a = 1; a < arguments.length; a++) e.a.push(arguments[a]); e.k = arguments[0] } this.q.push(e), window.dispatchEvent(t) }, this.message = function() { window.addEventListener("message", function(t) { "faitracker" === t.data.origin && this.call("message", t.data.type, t.data.message) }) }, this.message(), this.init("${projectToken}", { host: "${apiURL}" }), this }(), function() { var t = document.createElement("script"); t.type = "text/javascript", t.src = "${assetURL}", t.async = !0, (d = document.getElementsByTagName("script")[0]).parentNode.insertBefore(t, d) }()`;
  if (includeScriptTag) return `<script>${codeWithOutScriptTag}</script>`;
  return codeWithOutScriptTag;
};

export const JavascriptHeadDocumentation =
  'https://help.factors.ai/en/collections/8535799-installing-factors-sdk';
