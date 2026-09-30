package lib

import (
	"bytes"
	"strings"
	"text/template"
)

const (
	DefaultEmailSubjectRegister = "Welcome to Our Platform! Please Verify Your Email"
	DefaultBodyRegister         = `
<div style="font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif; max-width: 600px; margin: 0 auto; background-color: #ffffff; border: 1px solid #e0e0e0; border-radius: 8px; overflow: hidden;">
    <div style="background-color: #2563eb; padding: 20px; text-align: center;">
        <h2 style="color: #ffffff; margin: 0; font-size: 24px;">Welcome Aboard!</h2>
    </div>
    <div style="padding: 30px; line-height: 1.6; color: #333333;">
        <p style="font-size: 16px;">Hello <strong>{{.Name}}</strong>,</p>
        <p style="font-size: 16px;">We're thrilled to have you here! To get started and access all our features, please verify your email address by clicking the button below:</p>
        <div style="text-align: center; margin: 30px 0;">
            <a href="{{.VerificationLink}}" style="background-color: #2563eb; color: #ffffff; padding: 12px 24px; text-decoration: none; border-radius: 5px; font-weight: bold; display: inline-block; font-size: 16px;">Verify My Account</a>
        </div>
        <p style="font-size: 14px; color: #666666;">If you have any questions, feel free to reply to this email.</p>
    </div>
    <div style="background-color: #f9fafb; padding: 20px; text-align: center; border-top: 1px solid #eeeeee;">
        <p style="font-size: 12px; color: #999999; margin: 0;">Sent with by <strong>Ofren Dialsa</strong></p>
    </div>
</div>`

	DefaultEmailSubjectResend = "New Verification Link"
	DefaultBodyResend         = `
<div style="font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif; max-width: 600px; margin: 0 auto; background-color: #ffffff; border: 1px solid #e0e0e0; border-radius: 8px; overflow: hidden;">
    <div style="background-color: #4b5563; padding: 20px; text-align: center;">
        <h2 style="color: #ffffff; margin: 0; font-size: 24px;">Verification Link</h2>
    </div>
    <div style="padding: 30px; line-height: 1.6; color: #333333;">
        <p style="font-size: 16px;">Hello <strong>{{.Name}}</strong>,</p>
        <p style="font-size: 16px;">As requested, here is your new verification link to activate your account. This link will expire shortly.</p>
        <div style="text-align: center; margin: 30px 0;">
            <a href="{{.VerificationLink}}" style="background-color: #2563eb; color: #ffffff; padding: 12px 24px; text-decoration: none; border-radius: 5px; font-weight: bold; display: inline-block; font-size: 16px;">Verify Email Address</a>
        </div>
        <p style="font-size: 13px; color: #ef4444; background-color: #fef2f2; padding: 10px; border-radius: 4px; border: 1px solid #fee2e2;">
            <strong>Note:</strong> If you didn't request this, please ignore this email.
        </p>
    </div>
    <div style="background-color: #f9fafb; padding: 20px; text-align: center; border-top: 1px solid #eeeeee;">
        <p style="font-size: 12px; color: #999999; margin: 0;">Security notification by <strong>Ofren Dialsa</strong></p>
    </div>
</div>`

	DefaultEmailSubjectVerifyEmailSuccess = "Email Verified Successfully"
	DefaultBodyVerifyEmailSuccess         = `
<div style="font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif; max-width: 600px; margin: 0 auto; background-color: #ffffff; border: 1px solid #e0e0e0; border-radius: 8px; overflow: hidden;">
    <div style="background-color: #10b981; padding: 20px; text-align: center;">
        <h2 style="color: #ffffff; margin: 0; font-size: 24px;">Account Verified!</h2>
    </div>
    <div style="padding: 30px; line-height: 1.6; color: #333333;">
        <p style="font-size: 16px;">Hello <strong>{{.Name}}</strong>,</p>
        <p style="font-size: 16px;">Great news! Your email address has been successfully verified. Your account is now fully active and ready to use.</p>
        
        <div style="text-align: center; margin: 30px 0;">
            <a href="{{.LoginLink}}" style="background-color: #2563eb; color: #ffffff; padding: 12px 24px; text-decoration: none; border-radius: 5px; font-weight: bold; display: inline-block; font-size: 16px;">Login to Your Account</a>
        </div>

        <p style="font-size: 14px; color: #666666;">You can now access all features of our platform. We're excited to have you on board!</p>
    </div>
    <div style="background-color: #f9fafb; padding: 20px; text-align: center; border-top: 1px solid #eeeeee;">
        <p style="font-size: 12px; color: #999999; margin: 0;">This is an automated notification.<br>Sent with by <strong>Ofren Dialsa</strong></p>
    </div>
</div>`

	DefaultEmailSubjectResetPassword = "Reset Your Password"
	DefaultBodyResetPassword         = `
<div style="font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif; max-width: 600px; margin: 0 auto; background-color: #ffffff; border: 1px solid #e0e0e0; border-radius: 8px; overflow: hidden;">
    <div style="background-color: #2563eb; padding: 20px; text-align: center;">
        <h2 style="color: #ffffff; margin: 0; font-size: 24px;">Password Reset</h2>
    </div>
    <div style="padding: 30px; line-height: 1.6; color: #333333;">
        <p style="font-size: 16px;">Hello <strong>{{.Name}}</strong>,</p>
        <p style="font-size: 16px;">A request was made to reset your password. If this was you, please click the button below to set a new password:</p>
        <div style="text-align: center; margin: 30px 0;">
            <a href="{{.ResetLink}}" style="background-color: #2563eb; color: #ffffff; padding: 12px 24px; text-decoration: none; border-radius: 5px; font-weight: bold; display: inline-block; font-size: 16px;">Reset Password</a>
        </div>
        <p style="font-size: 14px; color: #666666;">If the button above does not work, please copy and paste the following link into your browser:</p>
        <div style="background-color: #f8f9fa; padding: 15px; border-radius: 4px; word-break: break-all; font-size: 13px; color: #2563eb; border: 1px solid #eeeeee;">{{.ResetLink}}</div>
    </div>
    <div style="background-color: #f9fafb; padding: 20px; text-align: center; border-top: 1px solid #eeeeee;">
        <p style="font-size: 12px; color: #999999; margin: 0;">If you did not request this, you can safely ignore this email.<br>Security notification by <strong>Ofren Dialsa</strong></p>
    </div>
</div>`

	DefaultEmailSubjectPasswordResetSuccess = "Password Reset Successful"
	DefaultBodyPasswordResetSuccess         = `
<div style="font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif; max-width: 600px; margin: 0 auto; background-color: #ffffff; border: 1px solid #e0e0e0; border-radius: 8px; overflow: hidden;">
    <div style="background-color: #10b981; padding: 20px; text-align: center;">
        <h2 style="color: #ffffff; margin: 0; font-size: 24px;">Reset Successful</h2>
    </div>
    <div style="padding: 30px; line-height: 1.6; color: #333333;">
        <p style="font-size: 16px;">Hello <strong>{{.Name}}</strong>,</p>
        <p style="font-size: 16px;">Your password has been <strong>successfully reset</strong>. You can now use your new password to log in to your account.</p>
        <div style="background-color: #ecfdf5; padding: 15px; border-radius: 4px; color: #065f46; border: 1px solid #d1fae5; font-size: 14px; margin-top: 20px;">
            <strong>Security Tip:</strong> If you did not perform this action, please contact our support team immediately to secure your account.
        </div>
    </div>
    <div style="background-color: #f9fafb; padding: 20px; text-align: center; border-top: 1px solid #eeeeee;">
        <p style="font-size: 12px; color: #999999; margin: 0;">This is an automated security notification.<br>Sent by <strong>Ofren Dialsa</strong></p>
    </div>
</div>`

	DefaultEmailSubjectPasswordChangeSuccess = "Password Changed Successfully"
	DefaultBodyPasswordChangeSuccess         = `
<div style="font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif; max-width: 600px; margin: 0 auto; background-color: #ffffff; border: 1px solid #e0e0e0; border-radius: 8px; overflow: hidden;">
    <div style="background-color: #10b981; padding: 20px; text-align: center;">
        <h2 style="color: #ffffff; margin: 0; font-size: 24px;">Password Changed</h2>
    </div>
    <div style="padding: 30px; line-height: 1.6; color: #333333;">
        <p style="font-size: 16px;">Hello <strong>{{.Name}}</strong>,</p>
        <p style="font-size: 16px;">Your password was changed successfully from your account settings.</p>
        <div style="background-color: #fffbeb; padding: 15px; border-radius: 4px; color: #92400e; border: 1px solid #fef3c7; font-size: 14px; margin-top: 20px;">
            <strong>Note:</strong> If you did not make this change, please reset your password immediately and check your active sessions.
        </div>
    </div>
    <div style="background-color: #f9fafb; padding: 20px; text-align: center; border-top: 1px solid #eeeeee;">
        <p style="font-size: 12px; color: #999999; margin: 0;">Security notification sent by <strong>Ofren Dialsa</strong></p>
    </div>
</div>`
)

type EmailVerificationData struct {
	Name             string
	VerificationLink string
}

type EmailDataResetPassword struct {
	Name      string
	ResetLink string
}

type EmailDataPasswordResetSuccess struct {
	Name string
}

type EmailDataVerifyEmailSuccess struct {
	Name      string
	LoginLink string
}

type EmailDataPasswordChangeSuccess struct {
	Name string
}

func BuildEmailBodyRegister(name, verificationLink string) (string, error) {
	htmlTemplate := DefaultBodyRegister

	tmpl, err := template.New("email").Parse(htmlTemplate)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	data := EmailVerificationData{
		Name:             name,
		VerificationLink: verificationLink,
	}

	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

func BuildEmailBodyResendVerification(name, verificationLink string) (string, error) {
	htmlTemplate := DefaultBodyResend

	tmpl, err := template.New("email").Parse(htmlTemplate)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	data := EmailVerificationData{
		Name:             name,
		VerificationLink: verificationLink,
	}

	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

func BuildEmailBodyVerifyEmailSuccess(name, loginLink string) (string, error) {
	htmlTemplate := DefaultBodyVerifyEmailSuccess

	tmpl, err := template.New("email").Parse(htmlTemplate)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	data := EmailDataVerifyEmailSuccess{
		Name:      name,
		LoginLink: loginLink,
	}

	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

func BuildEmailBodyResetPassword(name, resetPassword string) (string, error) {
	htmlTemplate := DefaultBodyResetPassword

	tmpl, err := template.New("email").Parse(htmlTemplate)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	data := EmailDataResetPassword{
		Name:      name,
		ResetLink: resetPassword,
	}

	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

func BuildEmailBodyPasswordResetSuccess(name string) (string, error) {
	htmlTemplate := DefaultBodyPasswordResetSuccess

	tmpl, err := template.New("email").Parse(htmlTemplate)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	data := EmailDataPasswordResetSuccess{
		Name: name,
	}

	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

func BuildEmailBodyPasswordChangeSuccess(name string) (string, error) {
	htmlTemplate := DefaultBodyPasswordChangeSuccess

	tmpl, err := template.New("email").Parse(htmlTemplate)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	data := EmailDataPasswordChangeSuccess{
		Name: name,
	}

	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

func ExtractNameFromEmail(email string) string {
	if parts := strings.SplitN(email, "@", 2); len(parts) > 0 && parts[0] != "" {
		return parts[0]
	}
	return email
}
