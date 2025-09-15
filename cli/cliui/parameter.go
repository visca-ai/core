package cliui

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/wirtualdev/pretty"
	"github.com/wirtualdev/serpent"
	"github.com/wirtualdev/wirtualdev/v2/wirtualsdk"
)

func RichParameter(inv *serpent.Invocation, templateVersionParameter wirtualsdk.TemplateVersionParameter, defaultOverrides map[string]string) (string, error) {
	label := templateVersionParameter.Name
	if templateVersionParameter.DisplayName != "" {
		label = templateVersionParameter.DisplayName
	}

	if templateVersionParameter.Ephemeral {
		label += pretty.Sprint(DefaultStyles.Warn, " (build option)")
	}

	_, _ = fmt.Fprintln(inv.Stdout, Bold(label))

	if templateVersionParameter.DescriptionPlaintext != "" {
		_, _ = fmt.Fprintln(inv.Stdout, "  "+strings.TrimSpace(strings.Join(strings.Split(templateVersionParameter.DescriptionPlaintext, "\n"), "\n  "))+"\n")
	}

	defaultValue := templateVersionParameter.DefaultValue
	if v, ok := defaultOverrides[templateVersionParameter.Name]; ok {
		defaultValue = v
	}

	var err error
	var value string
	if templateVersionParameter.Type == "list(string)" {
		// Move the cursor up a single line for nicer display!
		_, _ = fmt.Fprint(inv.Stdout, "\033[1A")

		var options []string
		err = json.Unmarshal([]byte(templateVersionParameter.DefaultValue), &options)
		if err != nil {
			return "", err
		}

		values, err := MultiSelect(inv, MultiSelectOptions{
			Options:  options,
			Defaults: options,
		})
		if err == nil {
			v, err := json.Marshal(&values)
			if err != nil {
				return "", err
			}

			_, _ = fmt.Fprintln(inv.Stdout)
			pretty.Fprintf(
				inv.Stdout,
				DefaultStyles.Prompt, "%s\n", strings.Join(values, ", "),
			)
			value = string(v)
		}
	} else if len(templateVersionParameter.Options) > 0 {
		// Move the cursor up a single line for nicer display!
		_, _ = fmt.Fprint(inv.Stdout, "\033[1A")
		var richParameterOption *wirtualsdk.TemplateVersionParameterOption
		richParameterOption, err = RichSelect(inv, RichSelectOptions{
			Options:    templateVersionParameter.Options,
			Default:    defaultValue,
			HideSearch: true,
		})
		if err == nil {
			_, _ = fmt.Fprintln(inv.Stdout)
			pretty.Fprintf(inv.Stdout, DefaultStyles.Prompt, "%s\n", richParameterOption.Name)
			value = richParameterOption.Value
		}
	} else {
		text := "Enter a value"
		if !templateVersionParameter.Required {
			text += fmt.Sprintf(" (default: %q)", defaultValue)
		}
		text += ":"

		value, err = Prompt(inv, PromptOptions{
			Text: Bold(text),
			Validate: func(value string) error {
				return validateRichPrompt(value, templateVersionParameter)
			},
		})
		value = strings.TrimSpace(value)
	}
	if err != nil {
		return "", err
	}

	// If they didn't specify anything, use the default value if set.
	if len(templateVersionParameter.Options) == 0 && value == "" {
		value = defaultValue
	}

	return value, nil
}

func validateRichPrompt(value string, p wirtualsdk.TemplateVersionParameter) error {
	return wirtualsdk.ValidateWorkspaceBuildParameter(p, &wirtualsdk.WorkspaceBuildParameter{
		Name:  p.Name,
		Value: value,
	}, nil)
}
