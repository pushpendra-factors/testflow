import React from 'react';
import AppLayout from './Views/AppLayout';
import Login from './Views/Pages/Login';
import ForgotPassword from './Views/Pages/ForgotPassword';
import ResetPassword from './Views/Pages/ResetPassword';
import { HashRouter, Route, Switch } from 'react-router-dom';

function App() {
  return (
    <div className="App">
      <HashRouter>
        <Switch>
          <Route exact path="/resetpassword" name="login" component={ResetPassword} />
          <Route exact path="/forgotpassword" name="login" component={ForgotPassword} />
          <Route exact path="/login" name="login" component={Login} />
          <Route path="/" name="Home" component={AppLayout} />
        </Switch>
      </HashRouter>
    </div>
  );
}

export default App;
