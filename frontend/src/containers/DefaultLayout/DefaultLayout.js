import React, { Component } from 'react';
import { connect } from "react-redux";
import { bindActionCreators } from 'redux';
import { Redirect, Route, Switch } from 'react-router-dom';
import { Collapse, Nav, NavItem, NavLink, Container, Button, Navbar } from 'reactstrap';
import InternalRoute from '../../routes.internal';

// sidebar nav config
import {sideBarItems,  internalSideBarItems } from '../../_nav';
// routes config
import {routes, internalRoutes}  from '../../routes';
import DefaultHeader from './DefaultHeader';
import {
  fetchProjects,
  fetchProjectModels
} from "../../actions/projectsActions";
import Loading from '../../loading';
import factorsicon from '../../assets/img/brand/factors-icon.svg';
import { fetchAgentInfo, fetchAgentBillingAccount, setLoginToken } from '../../actions/agentActions';
import { hotjar } from 'react-hotjar';
import { isProduction, isFromFactorsDomain, isTokenLogin } from '../../util';

// inits factorsai sdk for app.
import factorsai from '../../common/factorsaiObj';

const isHotJarExcludedEmail = (email) => {
  const excludedEmailDomains = [ "factors.ai" ];
  return email && excludedEmailDomains.indexOf(email.split("@")[1]) > -1;
}

const mapStateToProps = store => {
  return {
    currentProjectId: store.projects.currentProjectId,
    projects: store.projects.projects,
    isAgentLoggedIn: store.agents.isLoggedIn,
    agent: store.agents.agent,
    eventNames: store.projects.currentProjectEventNames,
  }
}

const mapDispatchToProps = dispatch => {
  return bindActionCreators({ 
    fetchProjects,
    fetchProjectModels,
    fetchAgentInfo,
    fetchAgentBillingAccount,
    setLoginToken,
  }, dispatch);
}

class DefaultLayout extends Component {
  constructor(props) {
    super(props);
    this.state = {
      projects: {
        loaded: false,
        error: null
      },
      agent: {
        loaded: false,
        error: null
      },
      showSetupProjectNotification: false,
      showSetupProjectModal: false,
    }
  }

  getLoginTokenFromQueryParams() {
    let qParams = window.location.href.split("?");
    if (qParams.length < 2) return "";

    // no other params allowed when token given.
    let params = qParams[1].split("=");
    if (params.length < 2 || params[0] != "token") return "";

    return params[1].trim();
  }

  componentWillMount() {
    let loginToken = this.getLoginTokenFromQueryParams();
    if (loginToken != "") this.props.setLoginToken(loginToken);
    this.setActiveIndexFromUrl();

    if (window.fcWidget) {
      window.fcWidget.init({
        token: "3208785c-3624-47c7-be9a-4f60aa0e60f9",
        host: "https://wchat.freshchat.com"
      });
    }

    this.props.fetchProjects()
      .then((action) => {
        this.setState({ 
          projects: {
            loaded: true
          }
        });
      })
      .catch((response) => {
        this.setState({ 
          projects: { 
            loaded: true, 
            error: response.payload
          } 
        });
      });

    this.props
      .fetchAgentInfo()
      .then((r) => {
        this.setState({
          agent: {
            loaded: true,
          },
        });

        factorsai.identify(r.data.email);
        factorsai.addUserProperties({
          email: r.data.email,
          firstName: r.data.first_name,
          lastName: r.data.last_name,
        });

        if (window.fcWidget) {
          window.fcWidget.setExternalId(r.data.email);
          window.fcWidget.user.setEmail(r.data.email);
          window.fcWidget.user.setFirstName(r.data.first_name);
        }

        if (window.gr && typeof window.gr === 'function') {
          window.gr('track', 'conversion', { email: r.data.email });
        }
      })
      .catch((r) => {
        this.setState({
          agent: {
            loaded: true,
            error: 'Failed to get agent information',
          },
        });
      });

    this.props.fetchAgentBillingAccount();    
  }

  componentDidUpdate(prevProps, prevState, snapshot) {
    if (prevProps.agent && this.props.agent && prevProps.agent.email != this.props.agent.email) {
      if (isProduction() && !isHotJarExcludedEmail(this.props.agent.email))
        hotjar.initialize(1259925, 6);
    }

    if (this.props.location !== prevProps.location) {
      this.setActiveIndexFromUrl();
    }

    if (window.fcWidget && this.props.currentProjectId && this.props.projects) {
      window.fcWidget.user.setProperties({
        "Project Id": this.props.currentProjectId,
        "Project Name": this.props.projects[this.props.currentProjectId].name,
      });

      factorsai.addUserProperties({
        projectId: this.props.currentProjectId,
      });
    }
  }

  refresh = () => {
    this.props.history.push('/refresh');
  }

  changeViewToAccountSettings = () => {
    this.props.history.push('/account_settings');
  }

  changeViewToUserProfile = () => {
    this.props.history.push('/profile');
  }

  isLoaded() {
    return this.state.projects.loaded && this.state.agent.loaded;
  }

  isAgentLoggedIn(){
    return this.props.isAgentLoggedIn;
  }

  getAgentName = () => {
    return (this.props.agent 
      && this.props.agent.first_name) ? this.props.agent.first_name : '';
  }

  renderProjectsDropdown() {    
    const selectableProjects = Array.from(
      Object.values(this.props.projects), 
      // selectable_projects object structure.
      project => ({ "label": project.name, "value": project.id }) 
    )
    if (selectableProjects.length == 0 ){
      return <DefaultHeader 
        refresh={this.refresh}
        changeViewToAccountSettings={this.changeViewToAccountSettings}
        changeViewToUserProfile={this.changeViewToUserProfile}
        getProfileName={this.getAgentName} 
        currentAgent={this.props.agent}
        showSetupProjectNotification={this.showSetupProjectNotification}
        showSetupProjectModal={this.state.showSetupProjectModal}
        closeSetupProjectModal={this.closeSetupProjectModal}
      />
    }

    return <DefaultHeader
      refresh={this.refresh}
      changeViewToAccountSettings={this.changeViewToAccountSettings}
      changeViewToUserProfile={this.changeViewToUserProfile}
      selectableProjects={selectableProjects}
      selectedProject={{
        label: this.props.projects[this.props.currentProjectId].name,
        value: this.props.currentProjectId 
      }}
      getProfileName={this.getAgentName}
      currentAgent={this.props.agent}
      showSetupProjectNotification={this.showSetupProjectNotification}
      showSetupProjectModal={this.state.showSetupProjectModal}
      closeSetupProjectModal={this.closeSetupProjectModal}
    />
  }

  showInternalSideBarItems() {
    return isFromFactorsDomain(this.props.agent.email);
  }

  popBookDemo = () => {
    Calendly.initPopupWidget({ url: 'https://calendly.com/factorsai/demo' });
    factorsai.track('clicked_book_demo', { 'source': 'app' });
    return false;
  }

  setActiveIndexFromUrl = () => {
    const url = window.location.hash.split("#")[1];
    
    const actNavIndex = sideBarItems.findIndex(el => el.url === url);

    this.setState({
      activeNavIndex: actNavIndex
    })
  }

  activateNav = (index) => {
    const actIndex = Number(index.currentTarget.attributes.getNamedItem("index").nodeValue);
    this.setState({
      activeNavIndex: actIndex
    });
  }

  renderSetupProjectNotification() {
    if (!this.state.showSetupProjectNotification) return null;
  
    return (
      <div style={{ width: '105%', textAlign: 'center', background: '#6610f2', margin: '0 -30px', 
        padding: '2px', fontSize: '13px', fontWeight: '700', background: '#6f42c1', color: '#FFF' }}>
        Setup your project to use FactorsAI
        <Button color='warning' style={{ margin: '5px', marginLeft: '10px', padding: '1px 7px', fontSize: '12px', color: '#333' }} 
          onClick={() => { this.toggleSetupProjectModal(); factorsai.track('clicked_complete_project_setup') }}>Setup Project</Button>
        <Button color='warning' style={{ margin: '5px', padding: '1px 7px', fontSize: '12px', color: '#333' }}
          onClick={this.popBookDemo}>Book A Demo</Button>
      </div>
    );
  }

  showSetupProjectNotification = (value) => {
    this.setState({ showSetupProjectNotification: value });
  }

  toggleSetupProjectModal = () => {
    this.setState({ showSetupProjectModal: !this.state.showSetupProjectModal });
  }

  closeSetupProjectModal = () => {
    this.setState({ showSetupProjectModal: false });
  }

  renderErrorPage() {
    return <div style={{ textAlign: "center", width: "100%", fontWeight: "700", marginTop: "250px"}}>
      <div style={{color: '#AAA', fontSize: '35px'}}>Oops! Something went wrong.</div>
      <a href="https://app.factors.ai/" style={{ fontWeight: "500", fontSize: "20px" }}>Go back to home</a>
    </div>;
  }

  renderSideBar(sideBarItemsToDisplay) {
    const sideBarItems = [];
    sideBarItemsToDisplay.items.map((item, index) => {
      const className = this.state.activeNavIndex === index? "nav-link active" :  "nav-link";

      sideBarItems.push(
        <li key={item.name} className="nav-item">
          <a index={index} className={className} href={"#" + item.url} onClick={this.activateNav}>
            <i className={"nav-icon " + item.icon}></i>
            {item.name}
          </a>
        </li>
      );
    })
    
    return (
    <div className="fapp-sidebar sidebar">
      <img style={{marginTop: '12px', marginBottom: '20px'}} src={factorsicon} />
      <div className="scrollbar-container sidebar-nav ps">
        <ul className="nav">
          {sideBarItems}
        </ul>
      </div>
    </div>
    )
  }

  render() {
    if (!this.isAgentLoggedIn()) return <Redirect to='/login' />;
    if (!this.isLoaded()) return <Loading />;

    if (this.state.projects.loaded && this.state.projects.error) 
      return this.renderErrorPage();

    let sideBarItemsToDisplay = {items: sideBarItems};

    if(this.showInternalSideBarItems()){
      sideBarItemsToDisplay.items = [...sideBarItems , ...internalSideBarItems]
    }

    let appHomePath = "/dashboard";
    if (isTokenLogin()) appHomePath = "/factor";

    return (
      <div className="app">
        <div className="app-body fapp-body">
            { this.renderSideBar(sideBarItemsToDisplay) }
          <main className="main fapp-main">
            <header className="fapp-header app-header navbar">
              { this.renderProjectsDropdown() }
            </header>
            <Container className='fapp-right-pane' fluid>
              { this.renderSetupProjectNotification() }
              <Switch>
                {
                  routes.map((route, idx) => {
                    return route.component ? (<Route key={idx} path={route.path} exact={route.exact} name={route.name} 
                      render={props => (<route.component {...props} />)} />) : (null);
                  })
                }
                {
                  internalRoutes.map((route, idx) => {
                    return (<InternalRoute key={idx} path={route.path} exact={route.exact} name={route.name} component={route.component}/>)
                  })
                }

                <Redirect from="/" to={appHomePath} />
              </Switch>
            </Container>
          </main>
        </div>
      </div>
    );
  }
}

export default connect(mapStateToProps, mapDispatchToProps)(DefaultLayout);
