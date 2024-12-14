// prompt package provides a set of functions to prompt the user to enter a value.
package prompt

import (
	"fmt"
	"strconv"
	"time"

	"github.com/manifoldco/promptui"
)

var (
	// ErrorMaxAttemptExceeded is returned when the maximum number of attempts is exceeded.
	ErrorMaxAttemptExceeded = fmt.Errorf("max attempts exceeded")
	// ErrorPasswordMismatch is returned when the password and the confirmation password do not match.
	ErrorPasswordMismatch = fmt.Errorf("password mismatch")
	// ErrorTimeDurationFormat is returned when the time duration format is invalid.
	ErrorTimeDurationFormat = fmt.Errorf("invalid time duration format")
	// ErrorTimeFormat is returned when the time format is invalid.
	ErrorTimeFormat = fmt.Errorf("invalid time format")
	// ErrorNumberFormat is returned when the number format is invalid.
	ErrorNumberFormat = fmt.Errorf("invalid number format")
)

// PromptAgainCallback is a callback function that is called when the user needs to be prompted again.
type PromptAgainCallback func(error) bool

type timeUnit string

const (
	// Nanosecond
	Nanosecond timeUnit = "ns"
	// Microsecond
	Microsecond timeUnit = "us"
	// Millisecond
	Millisecond timeUnit = "ms"
	// Second
	Second timeUnit = "s"
	// Minute
	Minute timeUnit = "m"
	// Hour
	Hour timeUnit = "h"
	// Day
	Day timeUnit = "d"
	// Week
	Week timeUnit = "w"
	// Month
	Month timeUnit = "M"
	// Year
	Year timeUnit = "y"
)

// PromptNumber prompts the user to enter a number.
func PromptNumber(
	label string,
	maxAttempts int,
	promptAgainCallback PromptAgainCallback,
	isConfirm bool,
) (int, error) {
	prompt := promptui.Prompt{
		Label:     label,
		IsConfirm: isConfirm,
	}
	var value int
	var success bool
	for i := 0; i < maxAttempts; i++ {
		result, err := prompt.Run()
		if err != nil {
			return 0, err
		}
		value, err = strconv.Atoi(result)
		if err != nil {
			if pass := promptAgainCallback(ErrorNumberFormat); !pass {
				continue
			}
		}
		success = true
		break
	}
	if !success {
		return 0, ErrorMaxAttemptExceeded
	}
	return value, nil
}

// PromptSelect prompts the user to select an item from a list.
func PromptSelect(
	label string,
	items []string,
) (string, error) {
	prompt := promptui.Select{
		Label: label,
		Items: items,
	}
	index, _, err := prompt.Run()
	if err != nil {
		return "", err
	}
	return items[index], nil
}

// PromptDuration prompts the user to enter a duration.
func PromptDuration(
	label string,
	timeUnit timeUnit,
	maxAttempts int,
	promptAgainCallback PromptAgainCallback,
	isConfirm bool,
) (time.Duration, error) {
	prompt := promptui.Prompt{
		Label:     label,
		IsConfirm: isConfirm,
	}
	var value time.Duration
	var success bool
	for i := 0; i < maxAttempts; i++ {
		result, err := prompt.Run()
		if err != nil {
			return 0, err
		}
		value, err = time.ParseDuration(result + string(timeUnit))
		if err != nil {
			if pass := promptAgainCallback(ErrorTimeDurationFormat); !pass {
				continue
			}
		}
		success = true
		break
	}
	if !success {
		return 0, ErrorMaxAttemptExceeded
	}
	return value, nil
}

// PromptBool prompts the user to enter a boolean value.
func PromptBool(label string) (bool, error) {
	prompt := promptui.Select{
		Label: label,
		Items: []string{"yes", "no"},
	}
	_, result, err := prompt.Run()
	if err != nil {
		return false, err
	}
	return result == "Yes", nil
}

// PromptPassword prompts the user to enter a password.
func PromptPassword(
	label string,
	confirmLabel string,
	maxAttempts int,
	promptAgainCallback PromptAgainCallback,
	isConfirm bool,
) (string, error) {
	prompt := promptui.Prompt{
		Label: label,
		Mask:  '*',
	}
	confirmPrompt := promptui.Prompt{
		Label: confirmLabel,
		Mask:  '*',
	}
	var value string
	var success bool
	for i := 0; i < maxAttempts; i++ {
		result, err := prompt.Run()
		if err != nil {
			return "", err
		}
		value = result
		if isConfirm {
			confirmValue, err := confirmPrompt.Run()
			if err != nil {
				return "", err
			}
			if value != confirmValue {
				if pass := promptAgainCallback(ErrorPasswordMismatch); !pass {
					continue
				}
			}
		}
		success = true
		break
	}
	if !success {
		return "", ErrorMaxAttemptExceeded
	}
	return value, nil
}

// PromptText prompts the user to enter a text.
func PromptText(
	label string,
	isConfirm bool,
) (string, error) {
	prompt := promptui.Prompt{
		Label:     label,
		IsConfirm: isConfirm,
	}
	result, err := prompt.Run()
	if err != nil {
		return "", err
	}
	return result, nil
}

// PromptTime prompts the user to enter a time.
func PromptTime(
	label string,
	layout string,
	maxAttempts int,
	promptAgainCallback PromptAgainCallback,
	isConfirm bool,
) (time.Time, error) {
	prompt := promptui.Prompt{
		Label:     label,
		IsConfirm: isConfirm,
	}
	var value time.Time
	var success bool
	for i := 0; i < maxAttempts; i++ {
		result, err := prompt.Run()
		if err != nil {
			return time.Time{}, err
		}
		value, err = time.Parse(layout, result)
		if err != nil {
			if pass := promptAgainCallback(ErrorTimeFormat); !pass {
				continue
			}
		}
		success = true
		break
	}

	if !success {
		return time.Time{}, ErrorMaxAttemptExceeded
	}
	return value, nil
}
