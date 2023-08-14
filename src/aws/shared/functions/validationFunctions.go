package main

import (
	"fmt"
	"regexp"
	"strings"
	"github.com/badoux/checkmail"
)

func validateUserInputs(username string, email string) ([]string, error) {
	errors, err := checkUniqueAttribute("email", email, "email")
	if err != nil {
		return nil, err
	}

	usernameErrors, err := checkUniqueAttribute("username", username, "")
	if err != nil {
		return nil, err
	}

	errors = append(errors, usernameErrors...)
	usernameErrors, err = validateInput(username, "Username")
	if err != nil {
		return nil, err
	}

	errors = append(errors, usernameErrors...)
	emailErrors, err := validateInput(email, "Email")
	if err != nil {
		return nil, err
	}

	errors = append(errors, emailErrors...)
	return errors, nil
}

func validateInput(value string, valueName string) ([]string, error) {
	var errors []string
	switch valueName {
		case "Bio":
			errors = append(errors, validateLength(value, valueName, 0, 1000))
		case "Email":
			errors = append(errors,
				validateNotEmpty(value, valueName),
				validateLength(value, valueName, 3, 256),
				validateIsEmail(value))
		case "FirstName", "LastName":
			errors = append(errors,
				validateLength(value, valueName, 0, 32),
				validateAgainstRegex(
					value,
					valueName,
					regexp.MustCompile(`[!"`+"`"+`'#%&,:;<>=@{}~$()*+/\\?[\]^|0-9]`),
					"cannot contain numbers or special characters"))
		case "Password":
			errors = append(errors,
				validateNotEmpty(value, valueName),
				validateLength(value, valueName, 8, 32))
		case "Rating":
			errors = append(errors,
				validateIsWholeNumber(value))
		case "Review":
			errors = append(errors,
				validateLength(value, valueName, 0, 500))
		case "Username":
			errors = append(errors,
				validateNotEmpty(value),
				validateLength(value),
				validateAgainstRegex(
					value,
					valueName,
					regexp.MustCompile(`[^A-Za-z0-9]+`),
					"cannot contain special characters"))
		default:
			return nil, fmt.Errorf("invalid value name")
	}

	return errors, nil
}

func validateAgainstRegex(value string, name string, regex *regexp.Regexp, message string) string {
	if regex.MatchString(value) {
		return fmt.Sprintf("%s %s", name, message)
	}
	return ""
}

func validateIsEmail(value string) string {
	if err := checkmail.ValidateFormat(value); err != nil {
		return "Email must be valid"
	}
	return ""
}

func validateIsWholeNumber(value string) string {
	if _, err := strconv.Atoi(value); err != nil {
		return fmt.Sprintf("%s must be a whole number", name)
	}
	return ""
}

func validateLength(value string, valueName string, min int64, max int64) string {
	if int64(len(value)) < min || int64(len(value)) > max {
		return fmt.Sprintf("%s must be between %d and %d characters", valueName)
	}
	return ""
}

func validateNotEmpty(value string) string {
	if len(strings.TrimSpace(value)) == 0 {
		return fmt.Sprintf("%s must not be empty", name)
	}
	return ""
}

func validateNotValue(value string) string {
	if value == match {
		return fmt.Sprintf("%s must not be %s", name)
	}
	return ""
}

func validateValue(value float64) string {
	if value < lowerBound || value > upperBound {
		return fmt.Sprintf("%s must be between %f and %f", name)
	}
	return ""
}

func validateWholeNumber(value float64) string {
	if math.Mod(float64(int64(value)), float64(1)) != 0.0 {
        return fmt.Sprintf("%s must be a whole number", name)
    }
    return ""
}
