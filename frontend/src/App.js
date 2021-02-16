import React, { Component } from 'react';
import { HashRouter, Route, Switch, useLocation, Redirect } from 'react-router-dom';
import './App.css';
// Styles
// CoreUI Icons Set
import '@coreui/icons/css/coreui-icons.min.css';
// Import Flag Icons Set
import 'flag-icon-css/css/flag-icon.min.css';
// Import Font Awesome Icons Set
import 'font-awesome/css/font-awesome.min.css';
// Import Simple Line Icons Set
import 'simple-line-icons/css/simple-line-icons.css';
// Import Main styles for this application
import './scss/style.css'

// Containers
import { DefaultLayout } from './containers';
// Pages
import { Login, Page404, Page500, Signup, Activate, SetPassword, ForgotPassword } from './views/Pages';

// import { renderRoutes } from 'react-router-config';
class App extends Component {
  componentWillMount = ()=> {
    const queryParams = new URLSearchParams(window.location.search)
    const code = queryParams.get("code")
    const state  = queryParams.get("state")
    if(code != null) {
      let splitUrl = window.location.href.split('/')
      let hostname = splitUrl[0] + "//" + splitUrl[2]
      window.location.href= `${hostname}/#/settings/linkedin/?code=${code}&state=${state}`
    }

  }
  render() {
    return (
      <HashRouter>
        <Switch>
          <Route exact path="/login" name="Login Page" component={Login} />
          <Route exact path="/signup" name="Signup Page" component={Signup} />
          <Route exact path="/setpassword" name="Set Password" component={SetPassword} />
          <Route exact path="/forgotpassword" name="Forgot Password" component={ForgotPassword} />
          <Route exact path="/activate" name="Activate" component={Activate} />
          <Route exact path="/404" name="Page 404" component={Page404} />
          <Route exact path="/500" name="Page 500" component={Page500} />
          <Route path="/" name="Home" component={DefaultLayout} />
        </Switch>
      </HashRouter>
    );
  }
}

export default App;
