package database

import (
	"log"
	"super-app-chonburi-go/internal/domain"
)

func Seed() {
	log.Println("Running database seeding...")
	seedTaxRates()
	seedTaxBusinesses()
}

func seedTaxRates() {
	rates := []domain.TaxRate{
		{
			TaxType:   "hotel_fee",
			NameTH:    "ค่าธรรมเนียมบำรุง อบจ. จากผู้เข้าพักโรงแรม",
			RateValue: 0.5,
			RateUnit:  "percentage",
			IsActive:  true,
		},
		{
			TaxType:   "oil_gas_tax",
			NameTH:    "ภาษีบำรุง อบจ.จากการค้าน้ำมัน/ก๊าซ",
			RateValue: 4.54,
			RateUnit:  "percentage",
			IsActive:  true,
		},
		{
			TaxType:   "tobacco_tax",
			NameTH:    "ภาษีบำรุง อบจ.จากการค้ายาสูบ",
			RateValue: 1.0,
			RateUnit:  "percentage",
			IsActive:  true,
		},
	}

	for _, rate := range rates {
		var existing domain.TaxRate
		err := DB.Where("tax_type = ?", rate.TaxType).First(&existing).Error
		if err != nil {

			if err := DB.Create(&rate).Error; err != nil {
				log.Printf("Failed to seed tax rate %s: %v", rate.TaxType, err)
			} else {
				log.Printf("Seeded tax rate: %s", rate.TaxType)
			}
		} else {
			log.Printf("Tax rate %s already exists, skipping.", rate.TaxType)
		}
	}
}

func seedTaxBusinesses() {
	businesses := []domain.TaxBusiness{
		{
			BusinessRegNumber: "1234567",
			NameTH:            "โรงแรมชลบุรี แกรนด์ พลาซ่า",
			TaxType:           "hotel_fee",
		},
		{
			BusinessRegNumber: "7654321",
			NameTH:            "สถานีบริการน้ำมันชลบุรีพลังงาน",
			TaxType:           "oil_gas_tax",
		},
		{
			BusinessRegNumber: "9999999",
			NameTH:            "ร้านค้าส่งยาสูบชลบุรี",
			TaxType:           "tobacco_tax",
		},
	}

	nameEN1 := "Chonburi Grand Plaza Hotel"
	owner1 := "นายสมชาย ใจดี"
	id1 := "1100100123456"
	phone1 := "0812345678"
	email1 := "somchai@example.com"
	addr1 := "123 หมู่ 1 ต.เสม็ด อ.เมืองชลบุรี จ.ชลบุรี 20000"
	businesses[0].NameEN = &nameEN1
	businesses[0].OwnerName = &owner1
	businesses[0].OwnerIdentityNumber = &id1
	businesses[0].ContactPhone = &phone1
	businesses[0].ContactEmail = &email1
	businesses[0].AddressDetail = &addr1

	nameEN2 := "Chonburi Energy Gas Station"
	owner2 := "นางสาวสมศรี สวยงาม"
	id2 := "1100100654321"
	phone2 := "0898765432"
	email2 := "somsri@example.com"
	addr2 := "456 ต.แสนสุข อ.เมืองชลบุรี จ.ชลบุรี 20130"
	businesses[1].NameEN = &nameEN2
	businesses[1].OwnerName = &owner2
	businesses[1].OwnerIdentityNumber = &id2
	businesses[1].ContactPhone = &phone2
	businesses[1].ContactEmail = &email2
	businesses[1].AddressDetail = &addr2

	nameEN3 := "Chonburi Tobacco Wholesaler"
	owner3 := "นายประหยัด มัธยัสถ์"
	id3 := "1100100999999"
	phone3 := "0867890123"
	email3 := "prayad@example.com"
	addr3 := "789 ต.บ้านสวน อ.เมืองชลบุรี จ.ชลบุรี 20000"
	businesses[2].NameEN = &nameEN3
	businesses[2].OwnerName = &owner3
	businesses[2].OwnerIdentityNumber = &id3
	businesses[2].ContactPhone = &phone3
	businesses[2].ContactEmail = &email3
	businesses[2].AddressDetail = &addr3

	for _, biz := range businesses {
		var existing domain.TaxBusiness
		err := DB.Where("business_reg_number = ?", biz.BusinessRegNumber).First(&existing).Error
		if err != nil {

			if err := DB.Create(&biz).Error; err != nil {
				log.Printf("Failed to seed tax business %s: %v", biz.BusinessRegNumber, err)
			} else {
				log.Printf("Seeded tax business: %s (%s)", biz.NameTH, biz.BusinessRegNumber)
			}
		} else {
			log.Printf("Tax business %s already exists, skipping.", biz.BusinessRegNumber)
		}
	}
}
