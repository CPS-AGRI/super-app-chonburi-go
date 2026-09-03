package usecase

import (
	"errors"
	"fmt"
	"log"
	"super-app-chonburi-go/internal/domain"
	"super-app-chonburi-go/pkg/mail"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type adminUseCase struct {
	adminRepo   domain.AdminRepository
	mailSender  mail.EmailSender
	frontendURL string
}

func NewAdminUseCase(adminRepo domain.AdminRepository, mailSender mail.EmailSender, frontendURL string) domain.AdminUseCase {
	return &adminUseCase{
		adminRepo:   adminRepo,
		mailSender:  mailSender,
		frontendURL: frontendURL,
	}
}

func (u *adminUseCase) GetAdmins(query domain.AdminQuery) (*domain.PaginatedAdminResponse, error) {
	if query.PageNumber < 1 {
		query.PageNumber = 1
	}
	if query.PageSize < 1 {
		query.PageSize = 20
	}
	return u.adminRepo.GetPaginated(query)
}

func (u *adminUseCase) GetAdminByID(id string) (*domain.Admin, error) {
	admin, err := u.adminRepo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if admin == nil {
		return nil, errors.New("admin not found")
	}
	return admin, nil
}

func (u *adminUseCase) CreateAdmin(admin *domain.Admin) error {
	if admin.Email == "" {
		return errors.New("email is required")
	}

	// 1. Validate Email Uniqueness (Must not exist in system)
	existingAdmin, _ := u.adminRepo.GetByEmail(admin.Email)
	if existingAdmin != nil {
		return errors.New("email already exists")
	}

	// 2. Generate One-Time Activation Token
	activationToken := domain.NewUUID()
	admin.VerifyRegistrationToken = activationToken

	// 3. Set Initial Password Hash (if provided hash it, otherwise hash random UUID)
	if admin.Password != "" {
		hashed, err := bcrypt.GenerateFromPassword([]byte(admin.Password), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		admin.PasswordHash = string(hashed)
	} else {
		// Temporary placeholder hash until user sets password via activation link
		randomTempPass := domain.NewUUID()
		hashed, err := bcrypt.GenerateFromPassword([]byte(randomTempPass), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		admin.PasswordHash = string(hashed)
	}

	admin.ID = domain.NewUUID()
	admin.CreatedBy = "system"
	admin.UpdatedBy = "system"
	admin.CreatedDate = time.Now()
	admin.UpdatedDate = time.Now()

	admin.Departments = []domain.Department{}
	for _, deptID := range admin.DepartmentIds {
		admin.Departments = append(admin.Departments, domain.Department{ID: deptID})
	}

	if err := u.adminRepo.Create(admin); err != nil {
		return err
	}

	// 4. Send Activation Email to Staff
	if u.mailSender != nil {
		go func(targetEmail, name, lastName, token string) {
			activationLink := fmt.Sprintf("%s/reset-password?token=%s&type=activation", u.frontendURL, token)
			subject := "คำเชิญเข้าใช้งานระบบ อบจ. ชลบุรี (Super App Chonburi)"
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
							ผู้ดูแลระบบได้สร้างบัญชีผู้ใช้งานสำหรับคุณในระบบจัดการของ <strong>องค์การบริหารส่วนจังหวัดชลบุรี</strong> เรียบร้อยแล้ว กรุณากดปุ่มด้านล่างเพื่อเปิดใช้งานบัญชีและตั้งรหัสผ่านของคุณ
						</p>
						<div style="text-align: center; margin: 32px 0;">
							<a href="%s" style="display: inline-block; background: #00A5D6; color: #ffffff; text-decoration: none; padding: 14px 32px; border-radius: 30px; font-size: 15px; font-weight: 600; box-shadow: 0 4px 12px rgba(0,165,214,0.35);">
								เปิดใช้งานบัญชีและตั้งรหัสผ่าน
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
								* ลิงก์นี้มีอายุการใช้งาน 48 ชั่วโมง เพื่อความปลอดภัยของข้อมูล
							</p>
						</div>
					</div>
				</div>
			</div>
			`, name, lastName, activationLink, activationLink)

			if err := u.mailSender.SendHTML([]string{targetEmail}, subject, body); err != nil {
				log.Printf("⚠️ [Mail Error] Failed to send activation email to %s: %v", targetEmail, err)
			} else {
				log.Printf("📧 [Mail Success] Activation email sent to %s", targetEmail)
			}
		}(admin.Email, admin.Name, admin.LastName, activationToken)
	}

	return nil
}

func (u *adminUseCase) UpdateAdmin(admin *domain.Admin) error {
	existingAdmin, err := u.adminRepo.GetByID(admin.ID)
	if err != nil || existingAdmin == nil {
		return errors.New("admin not found")
	}

	existingAdmin.Email = admin.Email
	existingAdmin.Name = admin.Name
	existingAdmin.LastName = admin.LastName
	existingAdmin.Phone = admin.Phone
	existingAdmin.Position = admin.Position
	existingAdmin.RoleId = admin.RoleId
	existingAdmin.UpdatedBy = "system"
	existingAdmin.UpdatedDate = time.Now()

	if admin.Password != "" {
		hashed, err := bcrypt.GenerateFromPassword([]byte(admin.Password), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		existingAdmin.PasswordHash = string(hashed)
	}

	existingAdmin.Departments = []domain.Department{}
	for _, deptID := range admin.DepartmentIds {
		existingAdmin.Departments = append(existingAdmin.Departments, domain.Department{ID: deptID})
	}

	return u.adminRepo.Update(existingAdmin)
}

func (u *adminUseCase) DeleteAdmin(id string) error {
	return u.adminRepo.Delete(id)
}
