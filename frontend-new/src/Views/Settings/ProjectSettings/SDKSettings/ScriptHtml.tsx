import React from 'react';

const ScriptHtml = ({
  projectToken,
  assetURL,
  apiURL
}: {
  projectToken: string;
  assetURL: string;
  apiURL: string;
}) => {
  return (
    <>
      <span style={{ color: '#2F80ED' }}>{`<script>`}</span>
      {`
window.faitracker=window.faitracker||function(){this.q=[];var t=new CustomEvent("FAITRACKER_QUEUED_EVENT");return this.init=function(t,e,a){this.TOKEN=t,this.INIT_PARAMS=e,this.INIT_CALLBACK=a,window.dispatchEvent(new CustomEvent("FAITRACKER_INIT_EVENT"))},this.call=function(){var e={k:"",a:[]};if(arguments&&arguments.length>=1){for(var a=1;a<arguments.length;a++)e.a.push(arguments[a]);e.k=arguments[0]}this.q.push(e),window.dispatchEvent(t)},this.message=function(){window.addEventListener("message",function(t){"faitracker"===t.data.origin&&this.call("message",t.data.type,t.data.message)})},this.message(),this.init("`}
      <span style={{ color: '#EB5757' }}>{projectToken}</span>
      {`",{host:"${apiURL}"}),this}(),function(){var t=document.createElement("script");t.type="text/javascript",t.src="${assetURL}",t.async=!0,(d=document.getElementsByTagName("script")[0]).parentNode.insertBefore(t,d)}();
`}
      <span style={{ color: '#2F80ED' }}>{`</script>`}</span>
    </>
  );
};

export default ScriptHtml;
