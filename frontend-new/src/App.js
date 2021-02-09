import React, { useEffect, lazy, Suspense } from "react";
// import AppLayout from "./Views/AppLayout";
// import Login from "./Views/Pages/Login";
// import ForgotPassword from "./Views/Pages/ForgotPassword";
// import ResetPassword from "./Views/Pages/ResetPassword";
// import SignUp from "./Views/Pages/SignUp";
// import Activate from "./Views/Pages/Activate";
// import FactorsInsights from "./Views/Factors/FactorsInsights";
import { connect } from "react-redux";
import {
  BrowserRouter as Router,
  Route,
  Switch,
  Redirect,
} from "react-router-dom";
import PageSuspenseLoader from "./components/SuspenseLoaders/PageSuspenseLoader";
import * as Sentry from "@sentry/react"; 

const AppLayout = lazy(() => import("./Views/AppLayout"));
const Login = lazy(() => import("./Views/Pages/Login"));
const ForgotPassword = lazy(() => import("./Views/Pages/ForgotPassword"));
const ResetPassword = lazy(() => import("./Views/Pages/ResetPassword"));
const SignUp = lazy(() => import("./Views/Pages/SignUp"));
const Activate = lazy(() => import("./Views/Pages/Activate"));
const FactorsInsights = lazy(() => import("./Views/Factors/FactorsInsights"));

function App({ isAgentLoggedIn, agent_details }) {
  useEffect(() => { 

 

    Sentry.setUser({  
      username: agent_details?.first_name,
      email: agent_details?.email,
      id: agent_details?.uuid,
    });



    if (!process.env.NODE_ENV || process.env.NODE_ENV === 'development') {
      // DEV env
    } else {
      // PROD ENV

      //intercom init and passing logged-in user-data 
    var APP_ID = "rvffkuu7";
    window.intercomSettings = {
      app_id: APP_ID,
      name: agent_details?.first_name,
      email: agent_details?.email,
      user_id: agent_details?.uuid,
    };

    (function () { 
      var w = window;
      var ic = w.Intercom;
      if (typeof ic === "function") {
        ic("reattach_activator");
        ic("update", w.intercomSettings);
      } else {
        var d = document;
        var i = function () {
          i.c(arguments);
        };
        i.q = [];
        i.c = function (args) {
          i.q.push(args);
        };
        w.Intercom = i;
        var l = function () {
          var s = d.createElement("script");
          s.type = "text/javascript";
          s.async = true;
          s.src = "https://widget.intercom.io/widget/" + APP_ID;
          var x = d.getElementsByTagName("script")[0];
          x.parentNode.insertBefore(s, x);
        };
        if (document.readyState === "complete") {
          l();
        } else if (w.attachEvent) {
          w.attachEvent("onload", l);
        } else {
          w.addEventListener("load", l, false);
        }
      }
    })();

      (function() {
        window.__insp = window.__insp || [];
        __insp.push(['wid', 1994835818]);
        __insp.push(['identify', agent_details?.email]);
        __insp.push(['tagSession', {
          name: agent_details?.first_name,
          email: agent_details?.email,
          user_id: agent_details?.uuid
        }]);
        var ldinsp = function(){
        if(typeof window.__inspld != "undefined") return; window.__inspld = 1; var insp = document.createElement('script'); insp.type = 'text/javascript'; insp.async = true; insp.id = "inspsync"; insp.src = ('https:' == document.location.protocol ? 'https' : 'http') + '://cdn.inspectlet.com/inspectlet.js?wid=1994835818&r=' + Math.floor(new Date().getTime()/3600000); var x = document.getElementsByTagName('script')[0]; x.parentNode.insertBefore(insp, x); };
        setTimeout(ldinsp, 0);
      })();

    } 
  });

  return (
    <div className="App">
      <Suspense fallback={<PageSuspenseLoader />}>
        <Router>
          <Switch>
            <Route exact path="/signup" name="login" component={SignUp} />
            <Route
              exact
              path="/activate"
              name="Activate"
              component={Activate}
            />
            <Route
              exact
              path="/setpassword"
              name="login"
              component={ResetPassword}
            />
            <Route
              exact
              path="/forgotpassword"
              name="login"
              component={ForgotPassword}
            />
            <Route exact path="/login" name="login" component={Login} />
            {isAgentLoggedIn ? (
              <Route
                exact
                path="/explain/insights"
                name="login"
                component={FactorsInsights}
              />
            ) : (
              <Redirect to="/login" />
            )}
            {isAgentLoggedIn ? (
              <Route path="/" name="Home" component={AppLayout} />
            ) : (
              <Redirect to="/login" />
            )}
          </Switch>
        </Router>
      </Suspense>
    </div>
  );
}

const mapStateToProps = (state) => ({
  isAgentLoggedIn: state.agent.isLoggedIn,
  agent_details: state.agent.agent_details,
});

export default connect(mapStateToProps)(App);
