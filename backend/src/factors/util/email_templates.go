package util

import "fmt"

// Golang html/template pkg can be used later for parsing & creating complex email templates
// https://hackernoon.com/sending-html-email-using-go-c464d03a26a6

const (
	AgentAccountActivation = "Activate your account on factors.ai"
	Heading                = "Welcome to factors.ai!"
	Line1                  = "To activate your account, please follow this link:"
	Line2                  = "Once there, create a password for your account, and you're good to go!"
	Line3                  = "Please feel free to contact us by replying to this email if you face any issues."
	Footer1                = "Regards,"
	Footer2                = "factors.ai Team"
)

/*
Welcome to factors.ai!

To activate your account, please follow this link:

http://factors-dev.com:3000/#/activate?token=eyJhdSI6ImY2OTE5MGQwLWQ0N2YtNDUyMS04ODJiLTViOWEwZWU4MDZkYyIsInBmIjoiTVRVMU1UUXpPRFE0TW54eFpWbzFZbkJoZDJ4cmNrTlZVRXhCV20wMVoyRkJNMFEyUlY5VFNGVnNWbFJpTWtGMVFtSlhNbUUyZVdJdFIwRlVjSFppUWpOTWRscDJlRUpFZHpod2NYZEtVa3hXUVU5U2FGTjFOVE5PWjB0blBUMTh2UUludnNPWk1xeExFSXFuQXZ1cU13RkJhcFhqVktNSTFtdi1zWUJIbUJ3PSJ9

Once there, create a password for your account, and you're good to go!
Please feel free to contact us by replying to this email if you face any issues.

Regards,
factors.ai Team
*/
func CreateActivationTemplate(link string) (subject, text, html string) {
	subject = AgentAccountActivation
	text = fmt.Sprintf("%s\n\n%s\n\n%s\n\n%s\n%s\n\n%s\n%s", Heading, Line1, link, Line2, Line3, Footer1, Footer2)
	html = fmt.Sprintf("%s<br><br>%s<br><br>%s<br><br>%s<br>%s<br><br>%s<br>%s", Heading, Line1, link, Line2, Line3, Footer1, Footer2)
	fmt.Println(text)
	return
}

const (
	AgentAccResetPassword = "factors.ai account password reset"
	AFPHeading            = "Password Change Request."
	AFPLine1              = "Forgot your Password? Just visit the link below to create a new password. Your login email is "
	AFPLine2              = "If you did not request a Password change, ignore this mail and your Password will remain same."
	AFPFooter1            = "Thanks!"
	AFPFooter2            = "factors.ai Team"
)

/*
Password Change Request.

Forgot your Password? Just visit the link below to create a new password. Your login email is (ankit@factors.ai).

http://factors-dev.com:3000/#/setpassword?token=eyJhdSI6IjljZmY4OTcyLWQ2ZmEtNGMwOC04NThhLTkyN2RjNmU2MmZhMSIsInBmIjoiTVRVMU1UUXpPVGN5Tlh4S2NXRXdWWGcyUTB4b2MxOTFjbTV4V0Y5Q2J6QlpaVmhWYVZsd2FubDFXakEwY1c5RE1HNDBiVVZZYlc5TVNXRkpOSEpEWkdZMFNqVkhlRTlpVldKWE4yc3hVM1ZzWW5sMk5ucHhkRUU5UFh5SVFBbHgxNGxxc3gtaFVZVnlxTWVHcERyYmlzenZpcDllMWNTTl9EMUxfUT09In0=

If you did not request a Password change, ignore this mail and your Password will remain same.

Thanks!
factors.ai Team
*/
func CreateForgotPasswordTemplate(agentEmail, link string) (subject, text, html string) {
	subject = AgentAccResetPassword
	text = fmt.Sprintf("%s\n\n%s(%s).\n\n%s\n\n%s\n\n%s\n%s", AFPHeading, AFPLine1, agentEmail, link, AFPLine2, AFPFooter1, AFPFooter2)
	html = fmt.Sprintf("%s<br><br>%s(%s).<br><br>%s<br><br>%s<br><br>%s<br>%s", AFPHeading, AFPLine1, agentEmail, link, AFPLine2, AFPFooter1, AFPFooter2)
	return
}
