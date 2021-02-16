import React, { Component } from 'react';
import { connect } from 'react-redux';
import { bindActionCreators } from 'redux';
import {
  Input,
  Card,
  CardBody,
  CardHeader,
  Button,
} from 'reactstrap';
import Toggle from 'react-toggle';
import facebookSvg from '../../assets/img/settings/logo-social-fb-facebook-icon.svg';
import FacebookLogin from 'react-facebook-login';
// import LinkedinLogin from 'react-linkedin-login-oauth2';
import { addLinkedinAccessToken, fetchProjectSettings } from '../../actions/projectsActions'
import Loading from '../../loading';
import Select from 'react-select';
import { createSelectOpts, makeSelectOpt, getHostURL } from '../../util';
import linkedinSVG from './../../assets/img/integrations/linkedin.svg'

var uniqid = require('uniqid')
const mapStateToProps = store => {
  return {
    currentProjectId: store.projects.currentProjectId,
    currentProjectSettings: store.projects.currentProjectSettings
    }
}
const mapDispatchToProps = dispatch => {
  return bindActionCreators({addLinkedinAccessToken, fetchProjectSettings}, dispatch)
}


class Linkedin extends Component {
    constructor(props) {
        super(props);
        this.state = {
          response: "",
          adAccounts : "",
          SelectedAdAccount : "",
          code : '',
          oauthResponse: {},
        }
    }
    componentWillMount= () => {
      this.props.fetchProjectSettings(this.props.currentProjectId)
        .then((r) => {
          this.setState({ loaded: true });
        })
        .catch((r) => {
          this.setState({loaded: true, error: r.payload });
        });
    }
    componentDidMount = ()=> {
      let urlSplit = window.location.hash.split("?code=")
      let code = ''
      let state = ''
      if(urlSplit.length >1 ) {
        code = urlSplit[1].split('&state=')[0]
        state = urlSplit[1].split('&state=')[1]
      }
      if(code != '' && state === 'factors') {
        let url= getHostURL()+'integrations/linkedin/auth'
        fetch(url,{
          method: 'POST',
          body: JSON.stringify({
            'code': code
          })
        }).then(response => {
          response.json().then(e=>
            { 
              this.setState({
              oauthResponse: e
            })
            fetch(getHostURL()+'integrations/linkedin/ad_accounts', {
              method: 'POST',
              body: JSON.stringify({
                'access_token': this.state.oauthResponse['access_token']
              })
            }
            ).then(response => {
              response.json().then(res => {
                let jsonRes = JSON.parse(res)
                let adAccounts = this.getAdAccounts(jsonRes.elements)
                this.setState({
                  adAccounts: adAccounts
                })
              })
            })
          })
        })
      }
    }
    
    getAdAccounts = (jsonRes) => {
      let adAccounts = jsonRes.map(res => {
        return {"value": res.id, "name": res.name}
      })
      return adAccounts
    }
    renderLinkedinLogin = () => {
       if(!(this.props.currentProjectSettings.int_linkedin_access_token)) {
        let hostname = window.location.hostname
        let protocol = window.location.protocol
        let port = window.location.port
        let redirect_uri = protocol + "//" + hostname + ":" + port
        if(port === undefined || port === '') {
          redirect_uri = protocol + "//" + hostname
        }
        let href = `https://www.linkedin.com/oauth/v2/authorization?response_type=code&client_id=${BUILD_CONFIG.linkedin_client_id}&redirect_uri=${redirect_uri}&state=factors&scope=r_basicprofile%20r_liteprofile%20r_ads_reporting%20r_ads`
        return (
          <a href={href}>
            <Button hidden={this.props.currentProjectSettings.int_linkedin_access_token} color='primary' style={{ marginTop: '-3px' }} 
                  outline> 
              <img src={linkedinSVG} style={{ marginRight: '6px', marginBottom: '3px', width: '15px' }}></img>Enable with LinkedIn
            </Button></a>
        )
      }
      else {
        return (
          <div>Logged In</div>
        )
      }
    }
    getAdAccountsOptSrc() {
      let opts = {}
      for(let i in this.state.adAccounts) {
        let adAccount = this.state.adAccounts[i];
        opts[adAccount.value] = adAccount.label;
      }
      return opts;
    }
    handleChange = e => {
      this.setState({
        SelectedAdAccount : e
      })
    }
    handleSubmit = e => {
      e.preventDefault();
      if (this.state.SelectAdAccount != "" ) {
        const data = {
          "int_linkedin_ad_account": this.state.SelectedAdAccount.value,
          "int_linkedin_refresh_token": this.state.oauthResponse["refresh_token"],
          "int_linkedin_refresh_token_expiry": this.state.oauthResponse["refresh_token_expires_in"],
          "project_id": this.props.currentProjectId.toString(),
          "int_linkedin_access_token": this.state.oauthResponse['access_token'],
          "int_linkedin_access_token_expiry": this.state.oauthResponse["expires_in"]
        }
        this.props.addLinkedinAccessToken(data).then(()=> {
          console.log("access token added")
        }).catch((e)=> console.log(e))
      }
    }
    formComponent = () => {
      if (!(this.props.currentProjectSettings.int_linkedin_access_token)) {
        if (this.state.adAccounts != "" && this.state.adAccounts.length != 0) {
          return (
            <div className="p-2">
              <form onSubmit={e => this.handleSubmit(e)}>
                <div className="w-50 pb-2">
                  <h5>Choose your ad account:</h5>
                  <Select
                  value={this.state.SelectedAdAccount}
                  onChange={this.handleChange}
                  options={createSelectOpts(this.getAdAccountsOptSrc())}
                  />
                </div>
                <input className="btn btn-primary shadow-none" type="submit" value="Submit"/>
              </form>
            </div>
            )
        }
        if (this.state.adAccounts != "" && this.state.adAccounts.length == 0) {
          return <div>You don't have any ad accounts associated to the id you logged in with.</div>
        }
    }else {
      if(this.props.currentProjectSettings.int_linkedin_ad_account !== "" || this.props.currentProjectSettings.int_linkedin_ad_account !== undefined) {
        return <h5 className="p-2 m-2">Selected ad account: {this.props.currentProjectSettings.int_linkedin_ad_account}</h5>
      }
    }
      return
    }
    render() {
        return (
          <div className='fapp-content fapp-content-margin'>
            <Card className='fapp-bordered-card'>
                <div>
                    <CardHeader className='fapp-button-header' style={{ marginBottom: '0' }}>
                        <strong>LinkedIn</strong>
                        <div style={{display: 'inline-block', float: 'right'}}>
                        {this.renderLinkedinLogin()}    
                        </div>
                    </CardHeader>
                </div>
                {this.formComponent()}
            </Card>
          </div>
        )
      }
}
export default connect(mapStateToProps, mapDispatchToProps)(Linkedin);