import React, {Component} from 'react';
import { Formik, Form, Field, ErrorMessage } from 'formik';
import { Container, Input, Button, Row, Col, CardGroup, CardBody, Card} from 'reactstrap';
import { bindActionCreators } from 'redux';
import { connect } from 'react-redux';
import * as yup from 'yup';
import queryString from 'query-string';
import { setPassword } from "../../../actions/agentActions";
import  {  MissingPassword, PasswordsDoNotMatch, PasswordMinEightChars } from '../ValidationMessages';
import SubmissionError from '../SubmissionError';

const mapDispatchToProps = dispatch => {
    return bindActionCreators({setPassword}, dispatch)
}

class SetPassword extends Component {
    constructor(props) {
        super(props);
        this.state = {
            error: null
        }
    }

    renderForm(token) {
        return(
            <Formik
                    initialValues={{password:'', ReenterPassword:''}}
                    validationSchema = {
                        yup.object().shape({
                            password: yup.string().required(MissingPassword).min(8, PasswordMinEightChars),
                            ReenterPassword: yup.string().required().oneOf([yup.ref('password'),null], PasswordsDoNotMatch)
                        })
                    }
                    onSubmit={(values, {setSubmitting}) => {
                        this.props.setPassword(values.password, token)
                        .then(() => {
                            setSubmitting(false);
                            this.props.history.push("/login");
                        })
                        .catch((msg) => {
                            setSubmitting(false);
                            this.setState({ error: msg });
                        });                             
                    }}
                >
                    {({isSubmitting, touched})=> (
                        <Form noValidate>
                            <h3 style={{textAlign: 'center', marginBottom: '30px', color: '#484848'}}>Reset Password</h3>
                            <SubmissionError message={this.state.error} />
                            <Input className='fapp-page-input fapp-big-font' style={{marginBottom: '20px'}} tag={Field} type="password" name="password" placeholder="Password"/>
                            {
                                touched.password &&
                                <ErrorMessage name="password">
                                    {msg => <span className='fapp-error-span' style={{marginTop: '-15px'}}>{msg}</span>}    
                                </ErrorMessage>
                            }
                            <Input className='fapp-page-input fapp-big-font' style={{marginBottom: '20px'}} tag={Field} type="password" name="ReenterPassword" placeholder="Re-enter Password"/>
                            {   
                                touched.ReenterPassword &&
                                <ErrorMessage name="ReenterPassword">
                                    {msg => <span className='fapp-error-span' style={{marginTop: '-15px'}}>{msg}</span>}    
                                </ErrorMessage>
                            }
                            <div style={{textAlign: 'center'}}>
                                <Button color='success' type='submit' disabled={isSubmitting} className='fapp-cta-button' style={{marginTop: '15px'}}>RESET PASSWORD</Button>
                            </div>
                        </Form>
                    )}
                </Formik>
        );
    }
    render(){
        let parsed = queryString.parse(this.props.location.search)
        return (
                <Container fluid>
                    <Row style={{backgroundColor: '#F7F8FD', height: '100vh'}}>
                        <Col md={{size: 6, offset: 3}}>
                        <Card style={{marginTop: '8rem', width: '65%', marginLeft: '15%'}} className="p-4 fapp-block-shadow">
                            <CardBody>
                            { this.renderForm(parsed.token) }
                            </CardBody>
                        </Card>
                        </Col>
                    </Row>
                </Container>
            );
        }
}

export default connect(null, mapDispatchToProps)(SetPassword);