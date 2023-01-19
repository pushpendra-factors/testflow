import React, { useEffect } from 'react';
import { useSelector } from 'react-redux';
import { Redirect, Route, useLocation } from 'react-router-dom';

/*
  This function is to Capitalize first string
*/
function capitalize(string) {
  return string[0].toUpperCase() + string.slice(1).toLowerCase();
}
function PrivateRoute({ component: Component, ...restOfProps }) {
  const { isLoggedIn } = useSelector((state) => state.agent);
  let location = useLocation();
  useEffect(() => {
    let pageName = '';
    if (location.pathname == '/') {
      pageName = 'Dashboard';
    } else {
      let initialPaths = location.pathname.split('/');
      let n = initialPaths.length;

      pageName = capitalize(initialPaths[n - 1]);
    }
    document.title = pageName + ' - FactorsAI';
    console.log(location);
  }, [location]);
  return (
    <Route
      {...restOfProps}
      render={(props) =>
        isLoggedIn ? <Component {...props} /> : <Redirect to='/login' />
      }
    />
  );
}

export default PrivateRoute;
