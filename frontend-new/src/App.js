import React, { useEffect } from 'react';
import AppLayout from './Views/AppLayout';
import Login from './Views/Pages/Login';
import ForgotPassword from './Views/Pages/ForgotPassword';
import ResetPassword from './Views/Pages/ResetPassword';
import SignUp from './Views/Pages/SignUp';
import Activate from './Views/Pages/Activate';
import FactorsInsights from './Views/Factors/FactorsInsights';
import { connect } from 'react-redux';
import {
  BrowserRouter as Router, Route, Switch, Redirect
} from 'react-router-dom';

function App({ isAgentLoggedIn, agent_details }) {

  useEffect(()=>{
  
  var APP_ID = "rvffkuu7";
  window.intercomSettings = {
      app_id: APP_ID,
      name: agent_details?.first_name,
      email: agent_details?.email,
      user_id: agent_details?.uuid
    };
  (function(){var w=window;var ic=w.Intercom;if(typeof ic==="function"){ic('reattach_activator');ic('update',w.intercomSettings);}else{var d=document;var i=function(){i.c(arguments);};i.q=[];i.c=function(args){i.q.push(args);};w.Intercom=i;var l=function(){var s=d.createElement('script');s.type='text/javascript';s.async=true;s.src='https://widget.intercom.io/widget/' + APP_ID;var x=d.getElementsByTagName('script')[0];x.parentNode.insertBefore(s, x);};if(document.readyState==='complete'){l();}else if(w.attachEvent){w.attachEvent('onload',l);}else{w.addEventListener('load',l,false);}}})();

  });

  return (
    <div className="App">
      <Router>
        <Switch>
          <Route exact path="/signup" name="login" component={SignUp} />
          <Route exact path="/activate" name="Activate" component={Activate} />
          <Route exact path="/setpassword" name="login" component={ResetPassword} />
          <Route exact path="/forgotpassword" name="login" component={ForgotPassword} />
          <Route exact path="/login" name="login" component={Login} />
          {isAgentLoggedIn ? <Route exact path="/explain/insights" name="login" component={FactorsInsights} /> : <Redirect to="/login" />}
          {isAgentLoggedIn ? <Route path="/" name="Home" component={AppLayout} /> : <Redirect to="/login" /> }
        </Switch>
      </Router>
    </div>
  );
}

const mapStateToProps = (state) => ({
  isAgentLoggedIn: state.agent.isLoggedIn,
  agent_details: state.agent.agent_details
});

export default connect(mapStateToProps)(App);
