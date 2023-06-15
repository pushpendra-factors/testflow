import { useEffect } from 'react';

const useKey = (targetKeys: string[], callback: (key: string) => void) => {
  useEffect(() => {
    // Event handler for keydown event
    const handleKeyDown = (event: KeyboardEvent) => {
      if (targetKeys.includes(event.key)) {
        callback(event.key);
      }
    };
    // Add event listener
    document.addEventListener('keydown', handleKeyDown);

    // Clean up the event listener
    return () => {
      document.removeEventListener('keydown', handleKeyDown);
    };
  }, [targetKeys, callback]);
};

export default useKey;
