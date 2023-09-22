import { useEffect } from 'react';

const useScript = ({
  url,
  crossOrigin,
  type = 'text/javascript',
  async = true,
  defer = false,
  id = ''
}: {
  url: string;
  crossOrigin: string;
  type: string;
  async: boolean;
  defer: boolean;
  id: string;
}) => {
  useEffect(() => {
    const script = document.createElement('script');

    script.src = url;
    script.id = id;
    script.async = async;
    script.type = type;
    script.defer = defer;
    script.crossOrigin = crossOrigin;

    document.body.appendChild(script);

    return () => {
      document.body.removeChild(script);
    };
  }, [url, async, crossOrigin, type, defer, id]);
};

export default useScript;
