package usecase

import (
	"errors"
	"fmt"
	"log"
	"strings"
	"super-app-chonburi-go/internal/domain"
	"super-app-chonburi-go/pkg/jwtutil"
	"super-app-chonburi-go/pkg/mail"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type authUseCase struct {
	adminRepo   domain.AdminRepository
	rtRepo      domain.AdminRefreshTokenRepository
	mailSender  mail.EmailSender
	frontendURL string
}

func NewAuthUseCase(adminRepo domain.AdminRepository, rtRepo domain.AdminRefreshTokenRepository, mailSender mail.EmailSender, frontendURL string) domain.AuthUseCase {
	return &authUseCase{
		adminRepo:   adminRepo,
		rtRepo:      rtRepo,
		mailSender:  mailSender,
		frontendURL: frontendURL,
	}
}

func (u *authUseCase) Login(email, password string) (string, string, domain.User, error) {
	admin, err := u.adminRepo.GetByEmail(email)
	if err != nil {
		return "", "", nil, errors.New("user not found")
	}

	if admin == nil {
		return "", "", nil, errors.New("user not found")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(admin.PasswordHash), []byte(password)); err != nil {
		return "", "", nil, errors.New("invalid credentials")
	}

	token, err := jwtutil.GenerateToken(admin)
	if err != nil {
		return "", "", nil, err
	}

	refreshToken := domain.NewUUID()
	rt := &domain.AdminRefreshToken{
		ID:          domain.NewUUID(),
		Token:       refreshToken,
		ExpiryTime:  time.Now().Add(7 * 24 * time.Hour),
		AdminUserId: admin.ID,
		CreatedBy:   admin.ID,
		UpdatedBy:   admin.ID,
		CreatedDate: time.Now(),
		UpdatedDate: time.Now(),
	}

	if err := u.rtRepo.Create(rt); err != nil {
		return "", "", nil, err
	}

	return token, refreshToken, admin, nil
}

func (u *authUseCase) RefreshToken(token string) (string, string, domain.User, error) {
	rt, err := u.rtRepo.GetByToken(token)
	if err != nil {
		return "", "", nil, errors.New("invalid refresh token")
	}

	if time.Now().After(rt.ExpiryTime) {
		_ = u.rtRepo.DeleteByToken(token)
		return "", "", nil, errors.New("refresh token expired")
	}

	admin, err := u.adminRepo.GetByID(rt.AdminUserId)
	if err != nil {
		return "", "", nil, err
	}

	newAccessToken, err := jwtutil.GenerateToken(admin)
	if err != nil {
		return "", "", nil, err
	}

	return newAccessToken, token, admin, nil
}

func (u *authUseCase) Logout(token string) error {
	return u.rtRepo.DeleteByToken(token)
}

func (u *authUseCase) Me(id string) (*domain.Admin, []string, error) {
	admin, err := u.adminRepo.GetByID(id)
	if err != nil {
		return nil, nil, err
	}

	permissions := []string{}

	if admin.Role != nil && (strings.EqualFold(admin.Role.Type, "superadmin") || strings.EqualFold(admin.Role.Type, "super_admin")) {
		permissions = append(permissions, "MANAGE_CITY", "MANAGE_ADMINS", "MANAGE_DEPARTMENTS", "VIEW_ALL_REPORTS")
	}

	uniqueKeys := make(map[string]bool)
	for _, dept := range admin.Departments {
		for _, module := range dept.Modules {
			if module.Key != nil && *module.Key != "" {
				uniqueKeys[*module.Key] = true
			}

			if module.ID == "d01b2ce5-34a9-498b-bba0-b1b8360f1ea9" ||
				module.NameTh == "ศูนย์ร้องทุกข์" ||
				module.NameEn == "Complaint Center" {
				uniqueKeys["ModuleComplaintCenter"] = true
			}
		}
	}

	for key := range uniqueKeys {
		permissions = append(permissions, key)
	}

	return admin, permissions, nil
}

func (u *authUseCase) ForgotPassword(email string) error {
	admin, err := u.adminRepo.GetByEmail(email)
	if err != nil || admin == nil {
		return errors.New("ไม่พบบัญชีผู้ใช้งานที่ผูกกับอีเมลนี้ในระบบ")
	}

	resetToken := domain.NewUUID()
	if err := u.adminRepo.UpdateFields(admin.ID, map[string]interface{}{
		"verify_forgot_password_token": resetToken,
		"updated_date":                 time.Now(),
	}); err != nil {
		return errors.New("เกิดข้อผิดพลาดในการสร้างคำขอรีเซ็ตรหัสผ่าน กรุณาลองใหม่อีกครั้ง")
	}

	if u.mailSender != nil {
		go func(targetEmail, name, lastName, token string) {
			resetLink := fmt.Sprintf("%s/reset-password?token=%s&type=reset", u.frontendURL, token)
			subject := "คำขอรีเซ็ตรหัสผ่านระบบ อบจ. ชลบุรี (Super App Chonburi)"
			body := fmt.Sprintf(`
			<div style="font-family: 'Prompt', 'Segoe UI', Arial, sans-serif; background-color: #f1f5f9; padding: 40px 15px;">
				<div style="max-width: 540px; margin: 0 auto; background: #ffffff; border-radius: 16px; overflow: hidden; box-shadow: 0 4px 20px rgba(0,0,0,0.06); border: 1px solid #e2e8f0;">
					<div style="background: linear-gradient(135deg, #00A5D6 0%%, #0284c7 100%%); padding: 32px 24px; text-align: center;">
						<h1 style="color: #ffffff; margin: 0; font-size: 24px; font-weight: 700; letter-spacing: -0.5px;">องค์การบริหารส่วนจังหวัดชลบุรี</h1>
						<p style="color: #e0f2fe; margin: 6px 0 0 0; font-size: 14px;">ระบบศูนย์กลางบริหารจัดการ Super App Chonburi</p>
					</div>
					<div style="padding: 32px 28px;">
						<h2 style="color: #1e293b; font-size: 18px; margin-top: 0; margin-bottom: 16px;">เรียน คุณ %s %s,</h2>
						<p style="color: #475569; font-size: 15px; line-height: 1.6; margin-bottom: 24px;">
							เราได้รับคำขอรีเซ็ตรหัสผ่านสำหรับบัญชีผู้ใช้งานของคุณ หากคุณเป็นผู้ส่งคำขอนี้ กรุณากดปุ่มด้านล่างเพื่อตั้งรหัสผ่านใหม่
						</p>
						<div style="text-align: center; margin: 32px 0;">
							<a href="%s" style="display: inline-block; background: #00A5D6; color: #ffffff; text-decoration: none; padding: 14px 32px; border-radius: 30px; font-size: 15px; font-weight: 600; box-shadow: 0 4px 12px rgba(0,165,214,0.35);">
								รีเซ็ตรหัสผ่านใหม่
							</a>
						</div>
						<p style="color: #64748b; font-size: 13px; line-height: 1.5; margin-bottom: 8px;">
							หรือคัดลอกลิงก์นี้ไปวางในเบราว์เซอร์ของคุณ:
						</p>
						<p style="background: #f8fafc; border: 1px solid #e2e8f0; padding: 10px 14px; border-radius: 8px; font-size: 12px; color: #0284c7; word-break: break-all; margin-top: 0;">
							%s
						</p>
						<div style="margin-top: 28px; padding-top: 20px; border-top: 1px solid #f1f5f9; text-align: center;">
							<p style="color: #94a3b8; font-size: 12px; margin: 0;">
								* หากคุณไม่ได้เป็นผู้ส่งคำขอนี้ สามารถเพิกเฉยต่ออีเมลฉบับนี้ได้ (ลิงก์มีอายุ 24 ชั่วโมง)
							</p>
						</div>
					</div>
				</div>
			</div>
			`, name, lastName, resetLink, resetLink)

			if err := u.mailSender.SendHTML([]string{targetEmail}, subject, body); err != nil {
				log.Printf("⚠️ [Mail Error] Failed to send forgot password email to %s: %v", targetEmail, err)
			} else {
				log.Printf("📧 [Mail Success] Forgot password email sent to %s", targetEmail)
			}
		}(admin.Email, admin.Name, admin.LastName, resetToken)
	}

	return nil
}

func (u *authUseCase) VerifyToken(token, tokenType string) (*domain.Admin, error) {
	if token == "" {
		return nil, errors.New("token is required")
	}

	if tokenType == "activation" {
		admin, err := u.adminRepo.GetByVerifyRegistrationToken(token)
		if err != nil || admin == nil {
			return nil, errors.New("ลิงก์เปิดใช้งานไม่ถูกต้องหรือถูกใช้งานไปแล้ว")
		}
		return admin, nil
	}

	// Default: reset password token
	admin, err := u.adminRepo.GetByVerifyForgotPasswordToken(token)
	if err != nil || admin == nil {
		return nil, errors.New("ลิงก์รีเซ็ตรหัสผ่านไม่ถูกต้องหรือหมดอายุแล้ว")
	}
	return admin, nil
}

func (u *authUseCase) ResetPassword(token, newPassword, tokenType string) error {
	if len(newPassword) < 8 {
		return errors.New("รหัสผ่านต้องมีความยาวอย่างน้อย 8 ตัวอักษร")
	}

	admin, err := u.VerifyToken(token, tokenType)
	if err != nil {
		return err
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	updateFields := map[string]interface{}{
		"password_hash":                string(hashed),
		"verify_forgot_password_token": nil,
		"updated_date":                 time.Now(),
	}
	if tokenType == "activation" {
		updateFields["verify_registration_token"] = ""
	}

	if err := u.adminRepo.UpdateFields(admin.ID, updateFields); err != nil {
		return errors.New("เกิดข้อผิดพลาดในการบันทึกรหัสผ่านใหม่ กรุณาลองใหม่อีกครั้ง")
	}

	return nil
}
