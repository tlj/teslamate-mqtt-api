package datapoints

var (
	ValidDatapoints map[string]bool

	basicDatapoints = map[string]bool{
		"battery_level":          true,
		"charge_energy_added":    true,
		"charge_limit_soc":       true,
		"display_name":           true,
		"est_battery_range_km":   true,
		"exterior_color":         true,
		"ideal_battery_range_km": true,
		"inside_temp":            true,
		"is_climate_on":          true,
		"is_preconditioning":     true,
		"outside_temp":           true,
		"model":                  true,
		"plugged_in":             true,
		"rated_battery_range_km": true,
		"spoiler_type":           true,
		"state":                  true,
		"time_to_full_charge":    true,
		"update_available":       true,
		"update_version":         true,
		"usable_battery_level":   true,
		"version":                true,
		"wheel_type":             true,
	}

	authenticatedDatapoints = map[string]bool{
		"doors_open":      true,
		"elevation":       true,
		"is_user_present": true,
		"latitude":        true,
		"longitude":       true,
		"locked":          true,
		"odometer":        true,
		"sentry_mode":     true,
		"speed":           true,
		"trunk_open":      true,
	}
)

func CalculateValidDatapoints(authenticated bool) map[string]bool {
	dps := make(map[string]bool)
	for k, v := range basicDatapoints {
		dps[k] = v
	}

	if authenticated {
		for k, v := range authenticatedDatapoints {
			dps[k] = v
		}
	}

	return dps
}
