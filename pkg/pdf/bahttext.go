package pdf

import (
	"fmt"
	"math"
)

var thaiNumbers = []string{"ศูนย์", "หนึ่ง", "สอง", "สาม", "สี่", "ห้า", "หก", "เจ็ด", "แปด", "เก้า"}
var thaiPositions = []string{"", "สิบ", "ร้อย", "พัน", "หมื่น", "แสน", "ล้าน"}

func ConvertToThaiBaht(amount float64) string {
	if amount == 0 {
		return "ศูนย์บาทถ้วน"
	}

	amount = math.Round(amount*100) / 100

	bahtPart := math.Floor(amount)
	satangPart := math.Round((amount - bahtPart) * 100)

	bahtText := convertNumberToThaiText(int64(bahtPart))
	satangText := ""

	if satangPart > 0 {
		satangText = convertNumberToThaiText(int64(satangPart)) + "สตางค์"
	}

	if bahtText != "" {
		if satangText == "" {
			return bahtText + "บาทถ้วน"
		}
		return bahtText + "บาท" + satangText
	}

	if satangText != "" {
		return satangText
	}

	return "ศูนย์บาทถ้วน"
}

func convertNumberToThaiText(number int64) string {
	if number == 0 {
		return ""
	}

	text := ""
	numberStr := fmt.Sprintf("%d", number)
	length := len(numberStr)

	for i := 0; i < length; i++ {
		digit := int(numberStr[i] - '0')
		pos := length - i - 1

		if digit == 0 {
			continue
		}

		if pos > 0 && pos%6 == 0 {
			text += thaiNumbers[digit] + "ล้าน"
			continue
		}

		realPos := pos % 6
		numText := thaiNumbers[digit]

		if realPos == 0 && digit == 1 && length > 1 && int(numberStr[i-1]-'0') != 0 {
			numText = "เอ็ด"
		}

		if realPos == 1 && digit == 2 {
			numText = "ยี่"
		}

		if realPos == 1 && digit == 1 {
			numText = ""
		}

		text += numText + thaiPositions[realPos]
	}

	if number >= 1000000 {

	}

	return text
}
