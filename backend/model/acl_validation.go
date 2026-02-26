package model

import "slices"

// ValidateContentACL validates ACL rules for content (pages/folders)
func ValidateContentACL(acl []AccessRule) error {
	return validateACL(acl, ValidContentOps)
}

// ValidateConfigACL validates ACL rules for global config
func ValidateConfigACL(acl []AccessRule) error {
	return validateACL(acl, ValidConfigOps)
}

func validateACL(acl []AccessRule, validOps []AccessOp) error {
	for _, rule := range acl {
		if err := validateSubject(rule.Subject); err != nil {
			return err
		}
		if err := validateOperations(rule.Operations, validOps); err != nil {
			return err
		}
	}
	return nil
}

func validateSubject(subject string) error {
	if subject == "anonymous" || subject == "all" {
		return nil
	}
	if len(subject) > 5 && subject[:5] == "user:" {
		userID := subject[5:]
		if userID != "" {
			return nil
		}
	}
	return ErrInvalidACLSubject
}

func validateOperations(ops []AccessOp, validOps []AccessOp) error {
	for _, op := range ops {
		if !slices.Contains(validOps, op) {
			return ErrInvalidACLOperation
		}
	}
	return nil
}
