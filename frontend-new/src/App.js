import React from 'react';
import AppLayout from './Views/AppLayout';
import Login from './Views/Pages/Login';
import ForgotPassword from './Views/Pages/ForgotPassword';
import ResetPassword from './Views/Pages/ResetPassword';
import SignUp from './Views/Pages/SignUp';
import Activate from './Views/Pages/Activate';
import { connect } from 'react-redux';
import {
  HashRouter, Route, Switch, Redirect
} from 'react-router-dom';

function App({ isAgentLoggedIn }) {
  return (
    <div className="App">
      <HashRouter>
        <Switch>
          <Route exact path="/signup" name="login" component={SignUp} />
          <Route exact path="/activate" name="Activate" component={Activate} />
          <Route exact path="/setpassword" name="login" component={ResetPassword} />
          <Route exact path="/forgotpassword" name="login" component={ForgotPassword} />
          <Route exact path="/login" name="login" component={Login} />
          {
            isAgentLoggedIn ? <Route path="/" name="Home" component={AppLayout} /> : <Redirect to="/login" />
          }
        </Switch>
      </HashRouter>
    </div>
  );
}

const mapStateToProps = (state) => ({
  isAgentLoggedIn: state.agent.isLoggedIn
});

export default connect(mapStateToProps)(App);
