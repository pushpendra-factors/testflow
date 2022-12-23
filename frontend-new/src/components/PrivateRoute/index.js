import React from 'react';
import { useSelector } from 'react-redux';
import { Redirect, Route } from 'react-router-dom';

function PrivateRoute({ component: Component, ...restOfProps }) {
  const { isLoggedIn } = useSelector((state) => state.agent);

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
