import { useEffect } from 'react';

const useScript = ({
  url,
  crossOrigin,
  type = 'text/javascript',
  async = true,
  defer = false,
  id = '',
  scriptInnerHTML = ''
}: {
  url: string;
  crossOrigin: string;
  type: string;
  async: boolean;
  defer: boolean;
  id: string;
  scriptInnerHTML?: string;
}) => {
  useEffect(() => {
    const script = document.createElement('script');

    if (scriptInnerHTML) {
      script.innerHTML = scriptInnerHTML;
    } else {
      script.id = id;
      script.async = async;
      script.type = type;
      script.defer = defer;
      script.crossOrigin = crossOrigin;
      script.src = url;
    }
    document.body.appendChild(script);

    return () => {
      document.body.removeChild(script);
    };
  }, [url, async, crossOrigin, type, defer, id, scriptInnerHTML]);
};

export default useScript;
