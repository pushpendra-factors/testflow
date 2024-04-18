import { lazy } from 'react';

const lazyWithRetry = (componentImport) =>
  lazy(async () => {
    try {
      const component = await componentImport();

      return component;
    } catch (error) {
      const lastReloadedTime = window.localStorage.getItem(
        'page-error-reloaded-last-time'
      );
      if (
        !lastReloadedTime ||
        Date.now() - Number(lastReloadedTime) > 60 * 1 * 1000
      ) {
        // Assuming that the user is not on the latest version of the application.
        // Let's refresh the page immediately.
        window.localStorage.setItem(
          'page-error-reloaded-last-time',
          Date.now()
        );
        if (process.env.NODE_ENV !== 'development')
          return window.location.reload();
      }

      // The page has already been reloaded
      // Assuming that user is already using the latest version of the application.
      // Let's let the application crash and raise the error.
      throw error;
    }
  });

export default lazyWithRetry;
