export default function retryDynamicImport(fn, retriesLeft = 5, interval = 1000) {
    return new Promise((resolve, reject) => {
      fn()
        .then(resolve)
        .catch((error) => {
          setTimeout(() => {
            if (retriesLeft === 1) {
              // reject('Unable to fetch the bundle, throw error');
              reject(error);
              return;
            } 
            retryDynamicImport(fn, retriesLeft - 1, interval).then(resolve, reject);
          }, interval);
        });
    });
  }
 