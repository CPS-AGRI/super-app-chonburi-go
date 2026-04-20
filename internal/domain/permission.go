package domain

type Permission string

const (
	ManageComplaints      Permission = "MANAGE_COMPLAINTS"
	ManageTaxes           Permission = "MANAGE_TAXES"
	ManageCitySettings    Permission = "MANAGE_CITY_SETTINGS"
	ManagePublicRelations Permission = "MANAGE_PUBLIC_RELATIONS"
	VerifyCitizens        Permission = "VERIFY_CITIZENS"
	ManageWeatherAlerts   Permission = "MANAGE_WEATHER_ALERTS"
)

func (p Permission) String() string {
	return string(p)
}
