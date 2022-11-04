import { useRef, useEffect } from 'react';

const useAutoFocus = () => {
  const inputRef = useRef(null);
  useEffect(() => {
    let timeout;
    if (inputRef?.current) {
      timeout = setTimeout(() => {
        if (inputRef?.current) {
          inputRef.current?.focus();
        }
      }, 100);
    }
    return () => {
      if (timeout) clearTimeout(timeout);
    };
  }, []);

  return inputRef;
};

export default useAutoFocus;
