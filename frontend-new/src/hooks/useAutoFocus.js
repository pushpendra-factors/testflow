import { useRef, useEffect } from 'react';

const useAutoFocus = (setAutoFocus = true) => {
  const inputRef = useRef(null);
  useEffect(() => {
    let timeout;

    if (setAutoFocus) {
      if (inputRef?.current) {
        timeout = setTimeout(() => {
          if (inputRef?.current) {
            inputRef.current?.focus();
          }
        }, 100);
      }
    }

    return () => {
      if (timeout) clearTimeout(timeout);
    };
  }, [setAutoFocus]);

  return inputRef;
};

export default useAutoFocus;
